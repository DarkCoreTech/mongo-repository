package complex_query

import (
	"context"
	"testing"
	"time"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_ComplexPipeline(b *testing.B) {
	ctx := context.Background()

	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	defer benchmark.CleanupTestData(ctx, db)

	// Setup test data
	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	// Insert test data with more complex structure
	testUsers := benchmark.GenerateComplexTestUsers(1000)
	err = repo.BulkInsert(ctx, testUsers)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		builder := mongokit.NewAggregateBuilder().
			Match(bson.M{
				"status": true,
				"created_at": bson.M{
					"$gte": time.Now().AddDate(0, -1, 0),
				},
			}).
			Group(bson.M{
				"year":  bson.M{"$year": "$created_at"},
				"month": bson.M{"$month": "$created_at"},
			}, bson.M{
				"count": bson.M{"$sum": 1},
				"users": bson.M{"$push": "$name"},
			}).
			Sort("count", -1).
			Limit(10).
			Project(bson.M{
				"_id":        1,
				"count":      1,
				"user_count": bson.M{"$size": "$users"},
			})

		_, err = repo.Aggregate(ctx, builder)
		if err != nil {
			b.Errorf("Complex pipeline failed: %v", err)
		}
	}
}
