package elencho

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Database struct {
	instance *sql.DB
}

func Make() *Database {
	return &Database{
		instance: nil,
	}
}

func (db *Database) Open() error {
	database, err := sql.Open("postgres", os.Getenv(databaseUrlEnv))
	if err != nil {
		return fmt.Errorf("error opening database: %q", err)
	}
	db.instance = database

	return nil
}

func (db *Database) Close() error {
	err := db.instance.Close()
	if err != nil {
		return fmt.Errorf("error while closing databse: %q", err)
	}

	return nil
}

func (db *Database) Insert(query sq.InsertBuilder) error {
	_, err := query.PlaceholderFormat(sq.Dollar).RunWith(db.instance).Exec()
	if err != nil {
		return fmt.Errorf("error while getting performing insert query '%s': %q", query, err)
	}

	return nil
}

func (db *Database) Select(query sq.SelectBuilder, block func(*sql.Rows) (interface{}, error)) ([]interface{}, error) {
	rows, err := query.PlaceholderFormat(sq.Dollar).RunWith(db.instance).Query()
	if err != nil {
		return nil, fmt.Errorf("error while getting performing select query '%s': %q", query, err)
	}

	mappedRows := make([]interface{}, 0)
	for rows.Next() {
		value, err := block(rows)
		if err != nil {
			return nil, fmt.Errorf("error while reading columns: %q", err)
		}

		mappedRows = append(mappedRows, value)
	}

	return mappedRows, nil
}

func (db *Database) Truncate(tableNames []string) error {
	tx, err := db.instance.Begin()
	if err != nil {
		return fmt.Errorf("error while beginning transaction: %q", err)
	}

	for _, v := range tableNames {
		stmt, err := tx.Prepare("TRUNCATE " + v + " RESTART IDENTITY CASCADE")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error while preparing statement: %q", err)
		}

		if _, err := stmt.Exec(); err != nil {
			tx.Rollback()
			return fmt.Errorf("error while executing statement: %q", err)
		}

		err = stmt.Close()
		if err != nil {
			return fmt.Errorf("error while close the statement: %q", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error while committing the transaction: %q", err)
	}

	return nil
}

// TODO: return error not Fatal
func connect(url string) []map[string]interface{} {
	client := http.Client{
		Timeout: time.Second * 10, // Maximum of 10 seconds because we don't need quick response time.
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("connecting to %s", url)
	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	fmt.Printf("connection successful")

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	var j interface{}
	parseErr := json.Unmarshal(body, &j)
	if parseErr != nil {
		log.Fatal(parseErr)
	}

	r := make([]map[string]interface{}, 0)
	for _, v := range j.([]interface{}) {
		r = append(r, v.(map[string]interface{}))
	}

	return r
}

func Scrape(url string, goquerySelector string, block func(e *colly.HTMLElement)) error {
	c := colly.NewCollector()
	c.OnHTML(goquerySelector, block)
	err := c.Visit(url)
	if err != nil {
		return fmt.Errorf("an error occurred while scraping the unibz website: %q", err)
	}
	return nil
}
