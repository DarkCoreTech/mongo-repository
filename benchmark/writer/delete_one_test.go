package writer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Benchmark_DeleteOne(b *testing.B) {
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
			Email:   fmt.Sprintf("delete_test_%d@example.com", i),
			Name:    fmt.Sprintf("Delete Test User %d", i),
			Status:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}

		_, err = repo.Insert(ctx, &user)
		if err != nil {
			b.Errorf("Insert failed: %v", err)
		}

		// Then delete it
		filter := bson.M{"email": user.Email}
		err = repo.DeleteOne(ctx, filter)
		if err != nil {
			b.Errorf("DeleteOne failed: %v", err)
		}
	}
}
