package writer

import (
	"context"
	"testing"
	"time"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_BulkUpdate(b *testing.B) {
	ctx := context.Background()
	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	defer benchmark.CleanupTestData(ctx, db)

	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	// Pre-insert some test data
	testUsers := benchmark.GenerateTestUsers(1000)
	err = repo.BulkInsert(ctx, testUsers)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := bson.M{"status": true}
		update := bson.M{
			"updated": time.Now(),
		}

		_, err = repo.BulkUpdate(ctx, filter, update, false)
		if err != nil {
			b.Errorf("BulkUpdate failed: %v", err)
		}
	}
}
