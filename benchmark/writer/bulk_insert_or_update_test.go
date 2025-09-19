package writer

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_BulkInsertOrUpdate(b *testing.B) {
	ctx := context.Background()
	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	defer benchmark.CleanupTestData(ctx, db)

	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	// Pre-insert some test data
	testUsers := benchmark.GenerateTestUsers(100)
	err = repo.BulkInsert(ctx, testUsers)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generate test data for each iteration
		users := benchmark.GenerateTestUsers(50)

		// Define filter function
		filterFn := func(doc benchmark.TestUser) bson.M {
			return bson.M{"email": doc.Email}
		}

		_, _, err = repo.BulkInsertOrUpdate(ctx, users, filterFn)
		if err != nil {
			b.Errorf("BulkInsertOrUpdate failed: %v", err)
		}
	}
}
