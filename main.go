package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mocksi/diffdash-drone/storage"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-repo>", os.Args[1])
	}

	config := storage.Config{
		BaseBranch:  "main",
		StoragePath: "/tmp/diff_dash/repos.duckdb",
		RepoPath:    os.Args[1],
	}

	// TODO: allow StoragePath to be configurable by the user
	repoDatabase, err := storage.SetupDuckDb(config)
	if err != nil {
		log.Fatalf("Could not create DB: %v", err)
	}
	err = repoDatabase.ExtractCommits()
	if err != nil {
		log.Fatalf("Error processing commits: %v", err)
	}

	fmt.Println("Successfully processed all commits.")
	defer repoDatabase.Close()
}
