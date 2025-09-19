package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Benchmark_Insert(b *testing.B) {
	ctx := context.Background()

	client, err := SetupBenchmark()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testUser := TestUser{
			Email:   fmt.Sprintf("test%d@example.com", i),
			Name:    "Test User",
			Status:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}

		_, err = client.Collection(TestUsers).InsertOne(ctx, testUser)
		if err != nil {
			b.Errorf("Insert failed: %v", err)
		}
	}
}
