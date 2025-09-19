package complex_query

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_CountWithMatch(b *testing.B) {
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
		builder := mongokit.NewAggregateBuilder().
			Match(bson.M{"status": true}).
			Count("total_active_users")

		_, err = repo.Aggregate(ctx, builder)
		if err != nil {
			b.Errorf("CountWithMatch failed: %v", err)
		}
	}
}
