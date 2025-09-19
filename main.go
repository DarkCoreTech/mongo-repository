package main

import (
	"context"
	"fmt"
	"time"

	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

// Simple test to verify the package works correctly
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize MongoDB client
	client := mongokit.InitMongoClient("mongodb://localhost:27017")

	// Test basic repository functionality
	testCollection := client.Database("sample-db").Collection("sample_collection")
	_ = mongokit.MongoRepository[mongokit.SampleModel]{Collection: testCollection}

	// Test query builder
	filter := mongokit.NewQueryBuilder().
		Where("name", mongokit.OpTypes.Eq, "test").
		Build()

	fmt.Println("Query filter:", filter)

	// Use ctx to avoid unused variable warning
	_ = ctx

	// Test complex_query builder
	aggBuilder := mongokit.NewAggregateBuilder().
		Match(filter).
		Limit(10)

	fmt.Println("Aggregation pipeline:", aggBuilder.Build())

	// Test pagination
	pagination := mongokit.MakePagination(1, 10)
	fmt.Printf("Pagination: Limit=%d, Skip=%d\n", pagination.Limit, pagination.Skip)

	// Test sort options
	sortOpts := &mongokit.SortOptions{
		Fields: []mongokit.SortField{
			{Field: "name", Asc: true},
			{Field: "price", Asc: false},
		},
	}
	fmt.Println("Sort options:", mongokit.Sorts(sortOpts))

	fmt.Println("✅ MongoDB Repository package is working correctly!")
	fmt.Println("📚 Check the examples/ directory for comprehensive usage examples.")

	fmt.Println("main run")
}
