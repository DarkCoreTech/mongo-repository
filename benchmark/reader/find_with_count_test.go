package reader

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_FindWithCount(b *testing.B) {
	ctx := context.Background()
	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	defer benchmark.CleanupTestData(ctx, db)

	// Setup test data
	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	// Insert test data
	testUsers := benchmark.GenerateTestUsers(1000)
	err = repo.BulkInsert(ctx, testUsers)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := bson.M{"status": true}
		sort := &mongokit.SortOption{Field: "created_at", Ascending: false}
		pagination := &mongokit.Pagination{Limit: 10, Skip: 0}

		_, _, err := repo.FindWithCount(ctx, filter, sort, pagination)
		if err != nil {
			b.Errorf("FindWithCount failed: %v", err)
		}
	}
}
