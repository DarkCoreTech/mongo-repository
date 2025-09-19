package writer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Benchmark_UpdateOne(b *testing.B) {
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
		user := testUsers[i%len(testUsers)]
		filter := bson.M{"email": user.Email}
		update := bson.M{
			"name":    fmt.Sprintf("Updated User %d", i),
			"updated": time.Now(),
		}

		err = repo.UpdateOne(ctx, filter, update, false)
		if err != nil {
			b.Errorf("UpdateOne failed: %v", err)
		}
	}
}
