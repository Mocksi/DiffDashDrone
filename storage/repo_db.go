package storage

import (
	"database/sql"

	"github.com/go-git/go-git/v5"
	_ "github.com/marcboeker/go-duckdb"
)

type RepoDatabase struct {
	Repository *git.Repository
	Database   *sql.DB
}

func SetupDuckDb(repoPath string, storagePath string) (rd *RepoDatabase, err error) {

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("duckdb", "/tmp/diff_dash/repos.duckdb")
	if err != nil {
		return nil, err
	}

	defer db.Close()
	return &RepoDatabase{
		Repository: repo,
		Database:   db,
	}, err
}
