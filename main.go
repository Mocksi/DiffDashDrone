package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mocksi/diffdash-drone/storage"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-repo>", os.Args[0])
	}

	repoPath := os.Args[1]

	repoDatabase, err := storage.SetupDuckDb(repoPath, "/tmp/diff_dash/repos.duckdb")
	if err != nil {
		log.Fatalf("Could not create DB: %v", err)
	}
	db := repoDatabase.Database
	r := repoDatabase.Repository

	fmt.Println("Creating Schema")
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS commits (hash STRING, filename STRING, message STRING)"); err != nil {
		log.Fatalf("Could not create table: %v", err)
	}

	ref, err := r.Head()
	if err != nil {
		log.Fatalf("Could not get HEAD: %v", err)
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Fatalf("Could not get commit log: %v", err)
	}

	fmt.Println("Capturing commits. May take up to 10 minutes.")
	err = cIter.ForEach(func(c *object.Commit) error {
		fIter, err := c.Files()

		if err != nil {
			return fmt.Errorf("could not find files: %v", err)
		}
		fIter.ForEach(func(f *object.File) error {
			_, err := db.Exec("INSERT INTO commits VALUES (?, ?, ?)", c.Hash.String(), f.Name, c.Message)
			if err != nil {
				return fmt.Errorf("could not insert commit: %v", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("could not find file: %v", err)
		}

		return err
	})
	if err != nil {
		log.Fatalf("Error processing commits: %v", err)
	}

	fmt.Println("Successfully processed all commits.")
	defer db.Close()
}
