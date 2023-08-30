package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
)

type Data struct {
	Timestamp  time.Time `json:"timestamp"`
	AssetID    int64     `json:"assetid"`
	AssetType  string    `json:"assettype"`
	MetricType string    `json:"metrictype"`
	Locations  []float64 `json:"locations"`
	Values     []float64 `json:"values"`
}

func MeasurePostgres() {
	// Open a connection to the PostgreSQL database
	connStr := "postgres://postgres:postgres@localhost/local?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist already
	createTable(db)

	// Read data from the JSON file
	jsonData, err := os.ReadFile("data.json")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the JSON data into a slice of Data structs
	var data []Data
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Fatal(err)
	}
	// Insert the data into the PostgreSQL table
	if insertTime, err := insertData(db, data); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Time to insert %d rows: %v\n", len(data), *insertTime)
	}

	fmt.Println("Data inserted into postgres successfully!")

	// Query the latest data for asset id = 1, asset type = circuit, metric type = red-distributed-temperature within the last 15 minutes
	fetchingTime, latestData, err := getLatestData(db, 1, "Circuit", "red-distributed-temperature", 15)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Time to fetch %d rows: %v\n", len(data), *fetchingTime)

	fmt.Printf("Latest data for asset id 1, asset type 'Circuit', metric type 'red-distributed-temperature': %+v\n", latestData)

	tableSize, err := getTableSize(db, "data")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Table Size: %d bytes\n", tableSize)
}

func getTableSize(db *sql.DB, tableName string) (int64, error) {
	var size int64
	query := fmt.Sprintf("SELECT pg_total_relation_size('%s')", tableName)
	err := db.QueryRow(query).Scan(&size)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func createTable(db *sql.DB) {
	createTableQuery := `
        CREATE TABLE IF NOT EXISTS data (
            timestamp TIMESTAMPTZ,
            assetid BIGINT,
            assettype VARCHAR(255),
            metrictype VARCHAR(255),
            locations FLOAT[],
            values FLOAT[]
        );`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func insertData(db *sql.DB, data []Data) (*time.Duration, error) {
	startTime := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(pq.CopyIn("data", "timestamp", "assetid", "assettype", "metrictype", "locations", "values"))
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	for _, d := range data {
		_, err = stmt.Exec(d.Timestamp, d.AssetID, d.AssetType, d.MetricType, pq.Array(d.Locations), pq.Array(d.Values))
		if err != nil {
			_ = stmt.Close()
			_ = tx.Rollback()
			return nil, err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	return &duration, nil
}

func getLatestData(db *sql.DB, assetid int64, assettype string, metrictype string, minutesAgo int) (*time.Duration, []*Data, error) {
	startTime := time.Now()
	var latestData []*Data

	query := `
        SELECT timestamp, assetid, assettype, metrictype, locations, values
        FROM data
        WHERE assetid = $1 AND assettype = $2 AND metrictype = $3
        AND timestamp > NOW() - INTERVAL '15 minutes'
        ORDER BY timestamp DESC`

	rows, err := db.Query(query, assetid, assettype, metrictype)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		data := &Data{}
		err := rows.Scan(
			&data.Timestamp,
			&data.AssetID,
			&data.AssetType,
			&data.MetricType,
			pq.Array(&data.Locations),
			pq.Array(&data.Values),
		)
		if err != nil {
			return nil, nil, err
		}

		latestData = append(latestData, data)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	return &duration, latestData, nil
}
