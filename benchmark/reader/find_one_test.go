package reader

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	"go.mongodb.org/mongo-driver/bson"

	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

func Benchmark_FindOne(b *testing.B) {
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
	testUsers := benchmark.GenerateTestUsers(100)
	err = repo.BulkInsert(ctx, testUsers)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		filter := bson.M{"email": testUsers[i%len(testUsers)].Email}

		_, err := repo.FindOne(ctx, filter)
		if err != nil {
			b.Errorf("FindOne failed: %v", err)
		}
	}
}
