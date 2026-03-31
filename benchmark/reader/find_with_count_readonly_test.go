package reader

import (
	"context"
	"testing"

	"github.com/darkcoretech/mongo-repository/benchmark"
	mongokit "github.com/darkcoretech/mongo-repository/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_FindWithCount_ReadOnly_CountMatchesCountAPI(t *testing.T) {
	ctx := context.Background()

	db, err := benchmark.SetupBenchmark()
	if err != nil {
		t.Skipf("benchmark setup unavailable, skipping integration test: %v", err)
	}

	collection := db.Collection(benchmark.TestUsers)
	repo := &mongokit.MongoRepository[benchmark.TestUser]{Collection: collection}

	filter := bson.M{}

	t.Run("without_isDeleted_param", func(t *testing.T) {
		items, total, err := repo.FindWithCount(ctx, filter, nil, nil)
		if err != nil {
			t.Fatalf("FindWithCount failed: %v", err)
		}

		count, err := repo.Count(ctx, filter)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if total != count {
			t.Fatalf("total mismatch: findWithCount=%d count=%d", total, count)
		}
		if int64(len(items)) != total {
			t.Fatalf("items length mismatch: len(items)=%d total=%d", len(items), total)
		}
	})

	t.Run("isDeleted_false", func(t *testing.T) {
		isDeleted := false

		items, total, err := repo.FindWithCount(ctx, filter, nil, nil, &isDeleted)
		if err != nil {
			t.Fatalf("FindWithCount failed: %v", err)
		}

		count, err := repo.Count(ctx, filter, &isDeleted)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if total != count {
			t.Fatalf("total mismatch: findWithCount=%d count=%d", total, count)
		}
		if int64(len(items)) != total {
			t.Fatalf("items length mismatch: len(items)=%d total=%d", len(items), total)
		}
	})

	t.Run("isDeleted_true", func(t *testing.T) {
		isDeleted := true

		items, total, err := repo.FindWithCount(ctx, filter, nil, nil, &isDeleted)
		if err != nil {
			t.Fatalf("FindWithCount failed: %v", err)
		}

		count, err := repo.Count(ctx, filter, &isDeleted)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if total != count {
			t.Fatalf("total mismatch: findWithCount=%d count=%d", total, count)
		}
		if int64(len(items)) != total {
			t.Fatalf("items length mismatch: len(items)=%d total=%d", len(items), total)
		}
	})
}
