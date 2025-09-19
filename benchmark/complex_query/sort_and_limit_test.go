package complex_query

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	"go.mongodb.org/mongo-driver/bson"

	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

func Benchmark_Aggregator_SortAndLimit(b *testing.B) {
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
			Sort("created_at", -1).
			Limit(50).
			Project(bson.M{
				"name":       1,
				"email":      1,
				"created_at": 1,
			})

		_, err = repo.Aggregate(ctx, builder)
		if err != nil {
			b.Errorf("SortAndLimit failed: %v", err)
		}
	}
}
