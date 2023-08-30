package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Data struct {
	Timestamp  time.Time `json:"timestamp"`
	AssetID    int64     `json:"assetID"`
	AssetType  string    `json:"assetType"`
	MetricType string    `json:"metricType"`
	Locations  []float64 `json:"locations"`
	Values     []float64 `json:"values"`
}

func MeasureMongo() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:example@localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// Load data from JSON file
	fileName := "data.json"
	dataBytes, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	var data []Data
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		log.Fatal(err)
	}

	// Insert data into MongoDB
	err = insertData(client, data)
	if err != nil {
		log.Fatal(err)
	}

	// Query latest data
	results, err := queryLatestData(client, 1, "Circuit", "red-distributed-temperature", time.Minute*15)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Latest data:")
	for _, result := range results {
		fmt.Printf("Timestamp: %v, AssetID: %v, AssetType: %v, MetricType: %v, Locations: %v, Values: %v\n",
			result.Timestamp, result.AssetID, result.AssetType, result.MetricType, result.Locations, result.Values)
	}
}

func insertData(client *mongo.Client, data []Data) error {
	collection := client.Database("mongo-db").Collection("mongo-collection")

	var docs []interface{}
	for _, d := range data {
		docs = append(docs, d)
	}

	_, err := collection.InsertMany(context.Background(), docs)
	return err
}

func queryLatestData(client *mongo.Client, assetID int64, assetType string, metricType string, duration time.Duration) ([]Data, error) {
	collection := client.Database("mongo-db").Collection("mongo-collection")

	filter := bson.M{
		"assetID":    assetID,
		"assetType":  assetType,
		"metricType": metricType,
		"timestamp": bson.M{
			"$gte": time.Now().Add(-duration),
		},
	}

	options := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(1)

	cur, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var results []Data
	for cur.Next(context.Background()) {
		var result Data
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
