package writer

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

func Benchmark_BulkInsert(b *testing.B) {
	ctx := context.Background()

	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generate test data for each iteration
		testUsers := benchmark.GenerateTestUsers(100)

		err = repo.BulkInsert(ctx, testUsers)
		if err != nil {
			b.Errorf("BulkInsert failed: %v", err)
		}
	}
}
