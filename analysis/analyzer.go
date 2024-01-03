package analysis

import (
	"database/sql"
	"fmt"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/mocksi/diffdash-drone/storage"
)

func FindTags(rd *storage.RepoDatabase, limit int) (*sql.Rows, error) {
	// FIXME: use better template or prepared statements
	// see:  https://duckdb.org/docs/sql/query_syntax/prepared_statements.html
	query := `
		WITH first_line_messages AS (
			SELECT LOWER(SPLIT_PART(message, '\n', 1)) AS first_line_lower
			FROM repos.commits
		), words AS (
			SELECT UNNEST(regexp_split_to_array(first_line_lower, '[^a-zA-Z]+')) AS word
			FROM first_line_messages
			WHERE first_line_lower IS NOT NULL
		), word_count AS (
			SELECT word, COUNT(*) as frequency
			FROM words
			WHERE word <> ''
			GROUP BY word
		)
		SELECT word, frequency
		FROM word_count
		WHERE frequency > 10
		AND word NOT IN (
		'the', 'be', 'to', 'of', 'and', 'a', 'in', 'that', 'have', 'i', 'for', 'from', 'is', 'on', 'by', 't', 'stories',
		'user', 'this', 'remove'
		)
		ORDER BY frequency DESC
		LIMIT 
	` + fmt.Sprintf("%d", limit) + ";"
	rows, err := rd.Database.Query(query)
	return rows, err
}

func FindBugspots(rd *storage.RepoDatabase) error {
	// TODO: allow user to choose their own fix tags
	query := `
		WITH all_commits AS (
			SELECT f.name as filename, COUNT(DISTINCT c.hash) as all_count
			FROM repos.commits c
			JOIN repos.file f ON c.hash = f.commit_hash
			GROUP BY f.name
		), fix_commits AS (
			SELECT f.name as filename, COUNT(DISTINCT c.hash) as fixes_count
			FROM repos.commits c
			JOIN repos.file f ON c.hash = f.commit_hash
			WHERE regexp_matches(LOWER(SPLIT_PART(message, '\n', 1)), '(fix?|fix(es|ed)?|close(s|d)?|revert(s|ed)?)')
			GROUP BY f.name
		)
		SELECT fc.filename, fc.fixes_count, ac.all_count,
			CASE WHEN ac.all_count > 0 THEN fc.fixes_count::FLOAT / ac.all_count ELSE 0 END as bug_spot_likelihood
		FROM fix_commits fc
		JOIN all_commits ac ON fc.filename = ac.filename
		ORDER BY fc.fixes_count DESC LIMIT 10;
	`
	rows, err := rd.Database.Query(query)
	if err != nil {
		return err
	}
	fmt.Println("filename", "fix commits", "all commits", "bugspot likelihood")
	for rows.Next() {
		var filename string
		var fixes_count string
		var all_count string
		var bug_spot_likelihood string
		err = rows.Scan(&filename, &fixes_count, &all_count, &bug_spot_likelihood)
		if err != nil {
			return err
		}
		fmt.Println(filename, fixes_count, all_count, bug_spot_likelihood)
	}
	return nil
}

func CountCommits(rd *storage.RepoDatabase) *sql.Row {
	query := `SELECT count(DISTINCT hash) FROM commits;`
	return rd.Database.QueryRow(query)
}

func QueryForExport(rd *storage.RepoDatabase) (*sql.Rows, error) {
	query := `
		WITH all_commits AS (
			SELECT f.name as filename, COUNT(DISTINCT c.hash) as all_count
			FROM repos.commits c
			JOIN repos.file f ON c.hash = f.commit_hash
			GROUP BY f.name
		), fix_commits AS (
			SELECT f.name as filename, COUNT(DISTINCT c.hash) as fixes_count
			FROM repos.commits c
			JOIN repos.file f ON c.hash = f.commit_hash
			WHERE regexp_matches(LOWER(SPLIT_PART(c.message, '\n', 1)), '(fix|fixes|fixed|close|closes|closed|revert|reverts|reverted)')
			GROUP BY f.name
		), bug_spot_scores AS (
			SELECT DISTINCT ac.filename,
				CASE WHEN ac.all_count > 0 THEN fc.fixes_count::FLOAT / ac.all_count ELSE 0 END as bug_spot_likelihood
			FROM all_commits ac
			LEFT JOIN fix_commits fc ON ac.filename = fc.filename
		)
		SELECT DISTINCT bss.filename, c.message, c.author_email, c.hash, bss.bug_spot_likelihood,  c.commit_timestamp
		FROM bug_spot_scores bss
		JOIN repos.file f ON bss.filename = f.name
		JOIN repos.commits c ON f.commit_hash = c.hash
		ORDER BY bss.filename, c.commit_timestamp DESC;
	`
	return rd.Database.Query(query)
}

type ExportRow struct {
	Filename          string `parquet:"name=filename, type=BYTE_ARRAY, encoding=PLAIN_DICTIONARY"`
	Message           string `parquet:"name=message, type=BYTE_ARRAY, encoding=PLAIN_DICTIONARY"`
	AuthorEmail       string `parquet:"name=author_email, type=BYTE_ARRAY, encoding=PLAIN_DICTIONARY"`
	Hash              string `parquet:"name=hash, type=BYTE_ARRAY, encoding=PLAIN_DICTIONARY"`
	BugSpotLikelihood string `parquet:"name=bug_spot_likelihood, type=BYTE_ARRAY, encoding=PLAIN_DICTIONARY"`
	CommitTimestamp   string `parquet:"name=commit_timestamp, type=BYTE_ARRAY, convertedtype=DATETIME"`
}

func AnalyzeWithLLM(rd *storage.RepoDatabase) error {
	rows, err := QueryForExport(rd)
	if err != nil {
		return err
	}

	fw, err := local.NewLocalFileWriter("output.parquet")
	if err != nil {
		return err
	}

	pw, err := writer.NewParquetWriter(fw, new(ExportRow), 2)
	if err != nil {
		return err
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		record := make([]string, len(columns))
		for i, val := range values {
			if val != nil {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		exportRow := ExportRow{
			Filename:          fmt.Sprintf("%v", record[0]),
			Message:           fmt.Sprintf("%v", record[1]),
			AuthorEmail:       fmt.Sprintf("%v", record[2]),
			Hash:              fmt.Sprintf("%v", record[3]),
			BugSpotLikelihood: fmt.Sprintf("%v", record[4]),
			CommitTimestamp:   fmt.Sprintf("%v", record[5]),
		}

		if err := pw.Write(exportRow); err != nil {
			fmt.Printf("Error writing row %v. Details : %v", exportRow, err)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}
	err = pw.WriteStop()
	defer rows.Close()
	defer fw.Close()
	return err
}
