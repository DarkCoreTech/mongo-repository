package writer

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

func Benchmark_BulkInsertWithIDs(b *testing.B) {
	ctx := context.Background()
	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	defer benchmark.CleanupTestData(ctx, db)

	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate test data for each iteration
		testUsers := benchmark.GenerateTestUsers(100)

		_, err = repo.BulkInsertWithIDs(ctx, testUsers)
		if err != nil {
			b.Errorf("BulkInsertWithIDs failed: %v", err)
		}
	}
}
