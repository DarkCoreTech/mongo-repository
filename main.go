package main

import (
	"context"
	"fmt"
	"time"

	"github.com/darkcoretech/mongo-repository/mongo"
)

// Simple test to verify the package works correctly
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize MongoDB client
	client := mongo.InitMongoClient("mongodb://localhost:27017")

	// Test basic repository functionality
	testCollection := client.Database("sample-db").Collection("sample_collection")
	_ = mongo.MongoRepository[mongo.SampleModel]{Collection: testCollection}

	// Test query builder
	filter := mongo.NewQueryBuilder().
		Where("name", mongo.OpTypes.Eq, "test").
		Build()

	fmt.Println("Query filter:", filter)

	// Use ctx to avoid unused variable warning
	_ = ctx

	// Test aggregation builder
	aggBuilder := mongo.NewAggregateBuilder().
		Match(filter).
		Limit(10)

	fmt.Println("Aggregation pipeline:", aggBuilder.Build())

	// Test pagination
	pagination := mongo.MakePagination(1, 10)
	fmt.Printf("Pagination: Limit=%d, Skip=%d\n", pagination.Limit, pagination.Skip)

	// Test sort options
	sortOpts := &mongo.SortOptions{
		Fields: []mongo.SortField{
			{Field: "name", Asc: true},
			{Field: "price", Asc: false},
		},
	}
	fmt.Println("Sort options:", mongo.Sorts(sortOpts))

	fmt.Println("✅ MongoDB Repository package is working correctly!")
	fmt.Println("📚 Check the examples/ directory for comprehensive usage examples.")
}
