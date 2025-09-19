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

func Benchmark_Writer_InsertOrUpdate(b *testing.B) {
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

		// Update the user
		user.Name = fmt.Sprintf("Updated User %d", i)
		user.Updated = time.Now()

		_, _, err := repo.InsertOrUpdate(ctx, filter, &user)
		if err != nil {
			b.Errorf("InsertOrUpdate failed: %v", err)
		}
	}
}
