package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mocksi/diffdash-drone/analysis"
	"github.com/mocksi/diffdash-drone/storage"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-repo>", os.Args[1])
	}

	// TODO: allow config to be configurable by the user
	config := storage.Config{
		BaseBranch:  "main",
		StoragePath: "/tmp/diff_dash",
		RepoPath:    os.Args[1],
	}

	// TODO: allow StoragePath to be configurable by the user
	repoDatabase, err := storage.SetupDuckDb(config)
	if err != nil {
		log.Fatalf("Could not create DB: %v", err)
	}
	var numCommits int
	err = analysis.CountCommits(repoDatabase).Scan(&numCommits)
	if err != nil {
		log.Fatalf("Could not open DB: %v", err)
	}

	if numCommits < 1 {
		fmt.Println("No commits found. Extracting commits...")

		err = repoDatabase.ExtractCommits()
		if err != nil {
			log.Fatalf("Error processing commits: %v", err)
		}
	}

	fmt.Println("Successfully processed all commits.")
	err = analysis.FindBugspots(repoDatabase)
	if err != nil {
		log.Fatalf("Error finding bugspots commits: %v", err)
	}

	fmt.Println("Successfully found bugspots. Generating parquet file...")

	defer repoDatabase.Close()
}
