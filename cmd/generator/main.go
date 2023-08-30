package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

type Data struct {
	Timestamp  string
	AssetID    int64
	AssetType  string
	MetricType string
	Locations  []float64
	Values     []float64
}

const (
	ROW_COUNT = 10
)

func main() {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	data := make([]Data, 0, ROW_COUNT)

	for i := 0; i < ROW_COUNT; i++ {
		timestamp := generateRandomTimestamp(rng)
		assetID := int64(i + 1)
		assetType := "Circuit"
		metricType := generateRandomMetricType(rng)
		locations := generateRandomFloats(rng, 30000)
		values := generateRandomFloats(rng, 30000)

		item := Data{
			Timestamp:  timestamp,
			AssetID:    assetID,
			AssetType:  assetType,
			MetricType: metricType,
			Locations:  locations,
			Values:     values,
		}

		data = append(data, item)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("data.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}
}

func generateRandomTimestamp(rng *rand.Rand) string {
	start := time.Now().Add(-24 * time.Hour) // Set start time to 24 hours ago
	end := time.Now()

	min := start.Unix()
	max := end.Unix()

	randomTime := rng.Int63n(max-min) + min
	timestamp := time.Unix(randomTime, 0).Format("2006-01-02T15:04:05Z07:00")

	return timestamp
}

func generateRandomMetricType(rng *rand.Rand) string {
	types := []string{"red-distributed-temperature", "yellow-distributed-temperature", "green-distributed-temperature"}
	return types[rng.Intn(len(types))]
}

func generateRandomFloats(rng *rand.Rand, count int) []float64 {
	floats := make([]float64, count)
	for i := 0; i < count; i++ {
		floats[i] = rng.Float64()
	}
	return floats
}
