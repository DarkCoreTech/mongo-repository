package writer

import (
	"context"
	"fmt"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_DeleteMany(b *testing.B) {
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
		// Insert test data first
		testUsers := benchmark.GenerateTestUsers(50)
		for j := range testUsers {
			testUsers[j].Email = fmt.Sprintf("delete_many_test_%d_%d@example.com", i, j)
		}

		err = repo.BulkInsert(ctx, testUsers)
		if err != nil {
			b.Errorf("BulkInsert failed: %v", err)
		}

		// Then delete them
		filter := bson.M{
			"email": bson.M{
				"$regex": fmt.Sprintf("delete_many_test_%d_", i),
			},
		}

		_, err = repo.DeleteMany(ctx, filter)
		if err != nil {
			b.Errorf("DeleteMany failed: %v", err)
		}
	}
}
