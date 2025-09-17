# MongoDB Repository Package

A comprehensive, type-safe MongoDB repository package for Go that provides a clean and intuitive API for MongoDB operations.

## Features

- **Generic Repository Pattern**: Type-safe operations with Go generics
- **Query Builder**: Fluent API for building complex MongoDB queries
- **Aggregation Builder**: Chainable aggregation pipeline builder
- **CRUD Operations**: Complete Create, Read, Update, Delete operations
- **Bulk Operations**: Efficient bulk insert, update, and upsert operations
- **Soft Delete**: Built-in soft delete functionality
- **Pagination**: Easy pagination support
- **Sorting**: Single and multi-field sorting
- **Connection Management**: Singleton MongoDB client with connection pooling

## Installation

```bash
go get github.com/darkcoretech/mongo-repository
```

## Quick Start

```go
package main

import (
    "context"
    "time"
    
    "github.com/darkcoretech/mongo-repository/mongo"
)

type SampleModel struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Email     string             `bson:"email,omitempty"`
    Category  string             `bson:"category,omitempty"`
    Price     float64            `bson:"price,omitempty"`
    IsActive  bool               `bson:"is_active"`
    CreatedAt time.Time          `bson:"created_at"`
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Initialize MongoDB client
    client := mongo.InitMongoClient("mongodb://localhost:27017")
    
    // Create repository
    collection := client.Database("sample-db").Collection("sample_collection")
    repo := mongo.MongoRepository[SampleModel]{Collection: collection}

    // Create a sample record
    sample := &SampleModel{
        Name:      "Sample Item",
        Email:     "sample@example.com",
        Category:  "Electronics",
        Price:     99.99,
        IsActive:  true,
        CreatedAt: time.Now(),
    }
    
    _, err := repo.Insert(ctx, sample)
    if err != nil {
        panic(err)
    }

    // Find samples with query builder
    filter := mongo.NewQueryBuilder().
        Where("category", mongo.OpTypes.Eq, "Electronics").
        Build()
    
    samples, err := repo.Find(ctx, filter, nil, nil)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d samples\n", len(samples))
}
```

## Query Builder

Build complex queries with a fluent API:

```go
// Simple query
filter := mongo.NewQueryBuilder().
    Where("age", mongo.OpTypes.Gte, 18).
    And("status", mongo.OpTypes.Eq, "active").
    Build()

// Complex OR query
orFilter := mongo.OrQuery(
    mongo.AndQuery(
        mongo.NewQueryBuilder().
            Where("role", mongo.OpTypes.Eq, "admin").
            Build(),
    ),
    mongo.AndQuery(
        mongo.NewQueryBuilder().
            Where("role", mongo.OpTypes.Eq, "moderator").
            And("verified", mongo.OpTypes.Eq, true).
            Build(),
    ),
)

// IN query
filter := mongo.NewQueryBuilder().
    WhereIn("status", "active", "pending", "verified").
    Build()
```

## Aggregation Builder

Build aggregation pipelines with a chainable API:

```go
pipeline := mongo.NewAggregateBuilder().
    Match(bson.M{"is_active": true}).
    Group("$category", bson.M{
        "count": bson.M{"$sum": 1},
        "avgPrice": bson.M{"$avg": "$price"},
    }).
    Sort("count", -1).
    Limit(10).
    Build()

results, err := repo.Aggregate(ctx, pipeline)
```

## CRUD Operations

### Create
```go
// Single insert
sample := &SampleModel{Name: "Sample Item", Price: 99.99, Category: "Electronics"}
id, err := repo.Insert(ctx, sample)

// Bulk insert
samples := []SampleModel{
    {Name: "Item 1", Price: 59.99, Category: "Electronics"},
    {Name: "Item 2", Price: 19.99, Category: "Books"},
}
err := repo.BulkInsert(ctx, samples)
```

### Read
```go
// Find one
user, err := repo.FindOne(ctx, filter)

// Find many with pagination
pagination := &mongo.Pagination{Limit: 10, Skip: 0}
sort := &mongo.SortOption{Field: "created_at", Ascending: false}
users, err := repo.Find(ctx, filter, sort, pagination)

// Find with count
users, total, err := repo.FindWithCount(ctx, filter, sort, pagination)
```

### Update
```go
// Update one
update := map[string]interface{}{
    "price": 89.99,
    "updated_at": time.Now(),
}
err := repo.UpdateOne(ctx, filter, update, false)

// Bulk update
updatedCount, err := repo.BulkUpdate(ctx, filter, update, false)

// Upsert
matched, upserted, err := repo.InsertOrUpdate(ctx, filter, &sample)
```

### Delete
```go
// Hard delete
err := repo.DeleteOne(ctx, filter)
deletedCount, err := repo.DeleteMany(ctx, filter)

// Soft delete
err := repo.DeleteOneSoft(ctx, filter, "admin")
deletedCount, err := repo.DeleteManySoft(ctx, filter, "admin")
```

## Pagination

```go
// Create pagination
pagination := mongo.MakePagination(1, 10) // page 1, 10 items per page

// Use with find
users, total, err := repo.FindWithCount(ctx, filter, sort, pagination)

// Calculate total pages
totalPages := (total + pagination.Limit - 1) / pagination.Limit
```

## Sorting

```go
// Single field sort
sort := &mongo.SortOption{
    Field:     "created_at",
    Ascending: false, // DESC
}

// Multi-field sort
sortOpts := &mongo.SortOptions{
    Fields: []mongo.SortField{
        {Field: "priority", Asc: false}, // DESC
        {Field: "created_at", Asc: true}, // ASC
    },
}
```

## Examples

Check the `examples/` directory for comprehensive usage examples:

- `basic_usage.go` - Basic CRUD operations and query building
- `advanced_usage.go` - Complex queries, aggregation, and advanced features
- `crud_operations.go` - Complete CRUD operations with all features

## API Reference

### Repository Interface

```go
type Repository[T any] interface {
    // Read operations
    Find(ctx context.Context, filter bson.M, sort *SortOption, pagination *Pagination, isDeleted ...*bool) ([]T, error)
    FindOne(ctx context.Context, filter bson.M, isDeleted ...*bool) (*T, error)
    FindWithCount(ctx context.Context, filter bson.M, sort *SortOption, pagination *Pagination, isDeleted ...*bool) ([]T, int64, error)
    Count(ctx context.Context, filter bson.M, isDeleted ...*bool) (int64, error)

    // Write operations
    Insert(ctx context.Context, doc *T) (interface{}, error)
    BulkInsert(ctx context.Context, docs []T) error
    UpdateOne(ctx context.Context, filter bson.M, update bson.M, upsert bool) error
    BulkUpdate(ctx context.Context, filter bson.M, update bson.M, upsert bool) (int64, error)
    DeleteOne(ctx context.Context, filter bson.M) error
    DeleteMany(ctx context.Context, filter bson.M) (int64, error)

    // Soft delete operations
    DeleteOneSoft(ctx context.Context, filter bson.M, deletedBy string) error
    DeleteManySoft(ctx context.Context, filter bson.M, deletedBy string) (int64, error)

    // Upsert operations
    InsertOrUpdate(ctx context.Context, filter bson.M, doc *T) (matched, upserted int64, err error)
    BulkInsertOrUpdate(ctx context.Context, docs []T, filterFn func(doc T) bson.M) (matched, upserted int64, err error)

    // Aggregation
    Aggregate(ctx context.Context, builder *AggregateBuilder) ([]bson.M, error)
}
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
