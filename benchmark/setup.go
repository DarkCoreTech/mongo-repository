package benchmark

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	mongokit "github.com/darkcoretech/mongo-repository/internal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	TestUsers = "benchmark_test_users"
)

type TestUser struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Email   string             `bson:"email"`
	Name    string             `bson:"name"`
	Status  bool               `bson:"status"`
	Created time.Time          `bson:"created_at"`
	Updated time.Time          `bson:"updated_at"`
}

func SetupBenchmark() (*mongo.Database, error) {
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DATABASE")

	mk := mongokit.InitMongoClient(uri)

	return mk.Database(dbName), nil
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	// Find the project root directory by looking for go.mod
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("could not find project root directory")
		}
		dir = parent
	}

	envFile := filepath.Join(dir, ".env")
	file, err := os.Open(envFile)
	if err != nil {
		return fmt.Errorf("could not open .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) > 1 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
