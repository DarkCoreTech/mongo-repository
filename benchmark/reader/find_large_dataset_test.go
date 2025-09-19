package reader

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_Find_LargeDataset(b *testing.B) {
	ctx := context.Background()
	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	
	defer benchmark.CleanupTestData(ctx, db)

	// Setup test data
	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	// Insert large test data
	testUsers := benchmark.GenerateTestUsers(10000)
	err = repo.BulkInsert(ctx, testUsers)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := bson.M{"status": true}
		sort := &mongokit.SortOption{Field: "created_at", Ascending: false}
		pagination := &mongokit.Pagination{Limit: 100, Skip: 0}

		_, err := repo.Find(ctx, filter, sort, pagination)
		if err != nil {
			b.Errorf("Find with large dataset failed: %v", err)
		}
	}
}
