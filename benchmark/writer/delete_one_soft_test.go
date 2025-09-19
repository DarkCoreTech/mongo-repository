package writer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/darkcoretech/mongo-repository/benchmark"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

func Benchmark_DeleteOneSoft(b *testing.B) {
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
		// Insert a user first
		user := benchmark.TestUser{
			ID:      primitive.NewObjectID(),
			Email:   fmt.Sprintf("soft_delete_test_%d@example.com", i),
			Name:    fmt.Sprintf("Soft Delete Test User %d", i),
			Status:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}

		_, err = repo.Insert(ctx, &user)
		if err != nil {
			b.Errorf("Insert failed: %v", err)
		}

		// Then soft delete it
		filter := bson.M{"email": user.Email}
		err = repo.DeleteOneSoft(ctx, filter, "benchmark_test")
		if err != nil {
			b.Errorf("DeleteOneSoft failed: %v", err)
		}
	}
}
