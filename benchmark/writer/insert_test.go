package writer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
)

func Benchmark_Insert(b *testing.B) {
	ctx := context.Background()

	db, err := benchmark.SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	collection := db.Collection(benchmark.TestUsers)

	for i := 0; i < b.N; i++ {
		testUser := benchmark.TestUser{
			Email:   fmt.Sprintf("test%d@example.com", i),
			Name:    "Test User",
			Status:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}

		repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

		_, err = repo.Insert(ctx, &testUser)
		if err != nil {
			b.Errorf("Insert failed: %v", err)
		}
	}
}
