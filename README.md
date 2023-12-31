![Diff Dash Drone@0 2x](https://github.com/Mocksi/DiffDashDrone/assets/63699/65e0f93e-2eaa-40b6-8388-995f177accd5)

# Your Wingman in Decoding and Defeating Bugs

## Overview

DiffDash Drone is a tool designed to analyze the git history of a repository and identify potential "bugspots".

It parses the `.git` history into a DuckDB database and generates a report in Parquet format highlighting files that are more likely to be bugspots, based on commit messages that indicate high churn and include keywords like "Revert" or "Fixes".

## Features

- **Git History Parsing**: Clone and parse git repository histories.
- **Bugspot Analysis**: Analyze commit messages and file churn to identify potential bugspots.
- **Report Generation**: Generate a detailed report in Parquet format for easy integration with data analysis tools.

## Requirements

- Go (Version 1.x or later)
- Access to a Git repository

## Dependencies

- `go-git`
- `go-github`
- `go-duckdb`
- `tqdm` (for progress bar visualization)

## Installation

Download the latest binaries from the [releases page](#TODO)

## Manual Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   ```
2. Install Go
3. Install dependencies:
   ```bash
   go mod download
   go mod verify
   ```
4. Build and Run:
   ```bash
   go build
   ./diffdash-drone
   ```

## Usage

    ```bash
    diffdash-drone [OPTIONS] <path-to-local-repository>
    ```

### Options

- `--help`: Display help information
- `--version`: Display version information
- `--parquet`: Only generate a Parquet file
- `--bugspots`: Only generate a Basic Bugspots report
- `--prediction`: Using a prediction model, generate an extended Bugspots report that can identify potential bugspots in the future

Note: If neither `--parquet` nor `--bugspots` are specified, both a Parquet file and a Basic Bugspots report will be generated.

## TODO:

- [x] Create a Storage package
- [x] Store all Git data in a DuckDB database.
- [x] Create an Analyzer package
- [ ] Create a Reporter package
- [ ] Hook up to the LLM cli
- [ ] Make magic happen
- [ ] Build release binaries
- [ ] Add a --help option
- [ ] Add a --version option
- [ ] Add a --parquet option
- [ ] Add a --bugspots option
- [ ] Add a --prediction option
- [ ] Enable having multiple repositories/multiple database files
