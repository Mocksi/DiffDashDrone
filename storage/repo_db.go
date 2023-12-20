package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	_ "github.com/marcboeker/go-duckdb"
)

type RepoDatabase struct {
	Repository *git.Repository
	Database   *sql.DB
	Config     Config
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetupDuckDb(c Config) (rd *RepoDatabase, err error) {
	repo, err := git.PlainOpen(c.RepoPath)
	if err != nil {
		return nil, err
	}

	err = createDirIfNotExist(c.StoragePath)
	if err != nil {
		return nil, err
	}

	// FIXME: enable handling multiple Git repositories
	dbPath := fmt.Sprintf("%s/repos.db", c.StoragePath)

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, err
	}

	err = CreateSchema(db)
	if err != nil {
		return nil, err
	}

	return &RepoDatabase{
		Repository: repo,
		Database:   db,
		Config:     c,
	}, err
}

func Query(db *sql.DB, query string) (*sql.Rows, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func CreateSchema(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS collaborator ( name STRING, email STRING)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS commits (hash STRING, author_email STRING, committer_email STRING, mergeTag STRING, message STRING, commit_timestamp TIMESTAMP_S)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS file (commit_hash STRING, name STRING)")

	return err
}

func (rd *RepoDatabase) InsertCollaborator(colab *object.Signature) error {
	_, err := rd.Database.Exec("INSERT INTO collaborator VALUES (?, ?)", colab.Name, colab.Email)
	return err
}

func (rd *RepoDatabase) InsertCommit(c *object.Commit) error {
	whenIsoFormat := c.Committer.When.Format(time.RFC3339)
	_, err := rd.Database.Exec("INSERT INTO commits VALUES (?, ?, ?, ?, ?, ?)", c.Hash.String(), c.Author.Email, c.Committer.Email, c.MergeTag, c.Message, whenIsoFormat)
	return err
}

func (rd *RepoDatabase) InsertFile(c *object.Commit, f *object.File) error {
	_, err := rd.Database.Exec("INSERT INTO file VALUES (?, ?)", c.Hash.String(), f.Name)
	if err != nil {
		return err
	}

	return nil
}

func (rd *RepoDatabase) ExtractCommits() error {
	fmt.Println("Capturing commits. May take up to 10 minutes.")

	referenceName := plumbing.ReferenceName("refs/heads/" + rd.Config.BaseBranch)
	ref, err := rd.Repository.Reference(referenceName, true)

	// FIXME: allow user to specify branch instead
	if err != nil {
		referenceName = plumbing.ReferenceName("refs/heads/master")
		ref, err = rd.Repository.Reference(referenceName, true)
	}

	if err != nil {
		return fmt.Errorf("could not find branch %s: %w", rd.Config.BaseBranch, err)
	}
	cIter, err := rd.Repository.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime, All: false})
	if err != nil {
		return err
	}

	var ErrExitLoop = errors.New("exit loop")

	count := 0
	// FIXME: remove limit or at least make it configurable.
	limit := 1000
	err = cIter.ForEach(func(c *object.Commit) error {
		if count >= 1000 {
			cIter.Close()
			return ErrExitLoop
		}
		collaborators := []*object.Signature{}
		collaborators = append(collaborators, &c.Author)
		collaborators = append(collaborators, &c.Committer)

		for _, collaborator := range collaborators {
			err = rd.InsertCollaborator(collaborator)
			if err != nil {
				return err
			}
		}

		err := rd.InsertCommit(c)
		if err != nil {
			return err
		}

		fIter, err := c.Files()

		if err != nil {
			return err
		}

		for {
			f, err := fIter.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			err = rd.InsertFile(c, f)
			if err != nil {
				return err
			}
		}
		fmt.Printf("\r%d/%d commits processed", count, limit)
		count++

		return nil
	})
	fmt.Println("Done!")
	if err != nil {
		if err == ErrExitLoop {
			return nil
		}
	}
	return err
}

func (rd *RepoDatabase) Close() {
	rd.Database.Close()
}
