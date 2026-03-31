package mongokit

import (
	"context"
	"errors"
	"maps"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OpTypes defines MongoDB query operators
var OpTypes = newOpTypes()

func newOpTypes() *opTypes {
	return &opTypes{
		Eq:    "$eq",
		Gt:    "$gt",
		Gte:   "$gte",
		Lt:    "$lt",
		Lte:   "$lte",
		Ne:    "$ne",
		In:    "$in",
		Nin:   "$nin",
		Regex: "$regex",
	}
}

type opTypes struct {
	Eq    OpType
	Gt    OpType
	Gte   OpType
	Lt    OpType
	Lte   OpType
	Ne    OpType
	In    OpType
	Nin   OpType
	Regex OpType
}

type OpType string

// Reader interface for read operations
type Reader[T any] interface {
	Find(ctx context.Context, filter bson.M, sort *SortOption, pagination *Pagination, isDeleted ...*bool) ([]T, error)
	FindOne(ctx context.Context, filter bson.M, isDeleted ...*bool) (*T, error)
	FindWithCount(ctx context.Context, filter bson.M, sort *SortOption, pagination *Pagination, isDeleted ...*bool) ([]T, int64, error)
	Count(ctx context.Context, filter bson.M, isDeleted ...*bool) (int64, error)
	Distinct(ctx context.Context, field string, filter bson.M) ([]interface{}, error)
}

// Writer interface for write operations
type Writer[T any] interface {
	Insert(ctx context.Context, doc *T) (interface{}, error)
	BulkInsert(ctx context.Context, docs []T) error
	BulkInsertWithIDs(ctx context.Context, docs []T) ([]interface{}, error)
	InsertOrUpdate(ctx context.Context, filter bson.M, doc *T) (matched, upserted int64, err error)
	BulkInsertOrUpdate(ctx context.Context, docs []T, filterFn func(doc T) bson.M) (matched, upserted int64, err error)
	BulkInsertOrUpdateForFields(ctx context.Context, docs []T, filterFn func(doc T) bson.M, setFn func(doc T) bson.M) (matched, upserted int64, err error)
	UpdateOne(ctx context.Context, filter bson.M, update bson.M, upsert bool) error
	BulkUpdate(ctx context.Context, filter bson.M, update bson.M, upsert bool) (int64, error)
	FindOneAndUpdate(ctx context.Context, filter bson.M, update bson.M, returnAfter bool, upsert bool) (*T, error)
	DeleteOne(ctx context.Context, filter bson.M) error
	DeleteMany(ctx context.Context, filter bson.M) (int64, error)
	DeleteOneSoft(ctx context.Context, filter bson.M, deletedBy string) error
	DeleteManySoft(ctx context.Context, filter bson.M, deletedBy string) (int64, error)
}

// Aggregator interface for complex_query operations
type Aggregator[T any] interface {
	Aggregate(ctx context.Context, builder *AggregateBuilder) ([]bson.M, error)
	AggregateWithOptions(ctx context.Context, builder *AggregateBuilder, opts *options.AggregateOptions) ([]bson.M, error)
	AggregateRaw(ctx context.Context, pipeline mongo.Pipeline) ([]bson.M, error)
}

// Repository combines all interfaces
type Repository[T any] interface {
	Reader[T]
	Writer[T]
	Aggregator[T]
}

// Query represents a single query condition
type Query struct {
	Field string
	Op    OpType
	Value interface{}
}

func (q Query) ToBSON() bson.M {
	return bson.M{q.Field: bson.M{string(q.Op): q.Value}}
}

// QueryBuilder builds complex queries
type QueryBuilder struct {
	conditions []bson.M
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{conditions: []bson.M{}}
}

func (qb *QueryBuilder) Where(field string, op OpType, value interface{}) *QueryBuilder {
	for i, c := range qb.conditions {
		if cond, ok := c[field]; ok {
			m := cond.(bson.M)
			m[string(op)] = value
			qb.conditions[i] = bson.M{field: m}
			return qb
		}
	}
	qb.conditions = append(qb.conditions, bson.M{field: bson.M{string(op): value}})
	return qb
}

func (qb *QueryBuilder) WhereIn(field string, values ...interface{}) *QueryBuilder {
	if len(values) == 1 {
		switch v := values[0].(type) {
		case bson.A:
			values = v
		case []interface{}:
			values = v
		case []primitive.ObjectID:
			tmp := make([]interface{}, len(v))
			for i, x := range v {
				tmp[i] = x
			}
			values = tmp
		}
	}

	for i, c := range qb.conditions {
		if cond, ok := c[field]; ok {
			m := cond.(bson.M)
			m[string(OpTypes.In)] = values
			qb.conditions[i] = bson.M{field: m}
			return qb
		}
	}

	qb.conditions = append(qb.conditions, bson.M{
		field: bson.M{string(OpTypes.In): values},
	})
	return qb
}

func (qb *QueryBuilder) And(field string, op OpType, value interface{}) *QueryBuilder {
	return qb.Where(field, op, value)
}

func (qb *QueryBuilder) Or(orConditions ...Query) *QueryBuilder {
	var orBSON []bson.M
	for _, q := range orConditions {
		orBSON = append(orBSON, q.ToBSON())
	}
	qb.conditions = append(qb.conditions, bson.M{"$or": orBSON})
	return qb
}

func (qb *QueryBuilder) Not(query Query) *QueryBuilder {
	qb.conditions = append(qb.conditions, bson.M{query.Field: bson.M{"$not": bson.M{string(query.Op): query.Value}}})
	return qb
}

func (qb *QueryBuilder) Build() bson.M {
	if len(qb.conditions) == 0 {
		return bson.M{}
	}
	if len(qb.conditions) == 1 {
		return qb.conditions[0]
	}
	return bson.M{"$and": qb.conditions}
}

// SortOption defines sorting options
type SortOption struct {
	Field     string
	Ascending bool
}

// Pagination defines pagination options
type Pagination struct {
	Limit int64
	Skip  int64
}

// MongoRepository is a generic MongoDB repository implementation
type MongoRepository[T any] struct {
	Collection *mongo.Collection
}

func (r *MongoRepository[T]) Insert(ctx context.Context, doc *T) (interface{}, error) {
	res, err := r.Collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return res.InsertedID, nil
}

func (r *MongoRepository[T]) InsertOrUpdate(ctx context.Context, filter bson.M, doc *T) (matched, upserted int64, err error) {
	res, err := r.Collection.ReplaceOne(ctx, filter, doc, options.Replace().SetUpsert(true))
	if err != nil {
		return 0, 0, err
	}
	var ups int64
	if res.UpsertedID != nil {
		ups = 1
	}
	return res.MatchedCount, ups, nil
}

func (r *MongoRepository[T]) BulkInsert(ctx context.Context, docs []T) error {
	insertDocs := make([]interface{}, len(docs))
	for i, d := range docs {
		insertDocs[i] = d
	}
	_, err := r.Collection.InsertMany(ctx, insertDocs)
	return err
}

func (r *MongoRepository[T]) BulkInsertWithIDs(ctx context.Context, docs []T) ([]interface{}, error) {
	insertDocs := make([]interface{}, len(docs))
	for i, d := range docs {
		insertDocs[i] = d
	}

	res, err := r.Collection.InsertMany(ctx, insertDocs)
	if err != nil {
		return nil, err
	}
	return res.InsertedIDs, nil
}

func (r *MongoRepository[T]) BulkInsertOrUpdateForFields(
	ctx context.Context,
	docs []T,
	filterFn func(doc T) bson.M,
	setFn func(doc T) bson.M,
) (matched, upserted int64, err error) {

	if len(docs) == 0 {
		return 0, 0, nil
	}

	models := make([]mongo.WriteModel, 0, len(docs))
	for _, d := range docs {
		u := mongo.NewUpdateOneModel().
			SetFilter(filterFn(d)).
			SetUpdate(bson.M{"$set": setFn(d)}).
			SetUpsert(true)
		models = append(models, u)
	}

	res, err := r.Collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return 0, 0, err
	}
	return res.MatchedCount, int64(len(res.UpsertedIDs)), nil
}

func (r *MongoRepository[T]) BulkInsertOrUpdate(
	ctx context.Context,
	docs []T,
	filterFn func(doc T) bson.M,
) (matched, upserted int64, err error) {

	if len(docs) == 0 {
		return 0, 0, nil
	}

	models := make([]mongo.WriteModel, 0, len(docs))
	for _, d := range docs {
		// Convert document to BSON and remove _id field to avoid immutable field error
		docBSON, err := bson.Marshal(d)
		if err != nil {
			return 0, 0, err
		}

		var docMap bson.M
		if err := bson.Unmarshal(docBSON, &docMap); err != nil {
			return 0, 0, err
		}

		// Remove _id field to avoid immutable field error during replacement
		delete(docMap, "_id")

		m := mongo.NewReplaceOneModel().
			SetFilter(filterFn(d)).
			SetReplacement(docMap).
			SetUpsert(true)
		models = append(models, m)
	}

	res, err := r.Collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return 0, 0, err
	}
	return res.MatchedCount, int64(len(res.UpsertedIDs)), nil
}

func (r *MongoRepository[T]) Find(
	ctx context.Context,
	filter bson.M,
	sort *SortOption,
	pagination *Pagination,
	isDeleted ...*bool,
) ([]T, error) {
	f := cloneFilter(filter)
	applyDeleteFilter(f, "is_deleted", isDeleted...)

	findOptions := options.Find()
	if sort != nil {
		direction := 1
		if !sort.Ascending {
			direction = -1
		}
		findOptions.SetSort(bson.D{{Key: sort.Field, Value: direction}})
	}
	if pagination != nil {
		findOptions.SetLimit(pagination.Limit)
		findOptions.SetSkip(pagination.Skip)
	}

	cursor, err := r.Collection.Find(ctx, f, findOptions)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var results []T
	for cursor.Next(ctx) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoRepository[T]) FindOne(
	ctx context.Context,
	filter bson.M,
	isDeleted ...*bool,
) (*T, error) {
	f := cloneFilter(filter)
	applyDeleteFilter(f, "is_deleted", isDeleted...)

	result := r.Collection.FindOne(ctx, f)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var item T
	if err := result.Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MongoRepository[T]) UpdateOne(ctx context.Context, filter bson.M, update bson.M, upsert bool) error {
	opts := options.Update().SetUpsert(upsert)
	res, err := r.Collection.UpdateOne(ctx, filter, bson.M{"$set": update}, opts)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 && !upsert {
		return errors.New("no documents matched")
	}
	return nil
}

func (r *MongoRepository[T]) BulkUpdate(ctx context.Context, filter bson.M, update bson.M, upsert bool) (int64, error) {
	opts := options.Update().SetUpsert(upsert)
	res, err := r.Collection.UpdateMany(ctx, filter, bson.M{"$set": update}, opts)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (r *MongoRepository[T]) DeleteOne(ctx context.Context, filter bson.M) error {
	res, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("no documents deleted")
	}
	return nil
}

func (r *MongoRepository[T]) DeleteMany(ctx context.Context, filter bson.M) (int64, error) {
	res, err := r.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	if res.DeletedCount == 0 {
		return 0, errors.New("no documents deleted")
	}
	return res.DeletedCount, nil
}

func (r *MongoRepository[T]) DeleteOneSoft(ctx context.Context, filter bson.M, deletedBy string) error {
	if filter == nil {
		filter = bson.M{}
	}
	if _, ok := filter["is_deleted"]; !ok {
		filter["is_deleted"] = bson.M{"$ne": true}
	}

	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{
		"is_deleted": true,
		"deleted_at": &now,
		"deleted_by": deletedBy,
	}}

	res, err := r.Collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("no documents matched")
	}
	if res.ModifiedCount == 0 {
		return errors.New("document matched but not modified")
	}
	return nil
}

func (r *MongoRepository[T]) DeleteManySoft(ctx context.Context, filter bson.M, deletedBy string) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	if _, ok := filter["is_deleted"]; !ok {
		filter["is_deleted"] = bson.M{"$ne": true}
	}

	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{
		"is_deleted": true,
		"deleted_at": &now,
		"deleted_by": deletedBy,
	}}

	res, err := r.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	if res.MatchedCount == 0 {
		return 0, errors.New("no documents matched")
	}
	if res.ModifiedCount == 0 {
		return 0, errors.New("documents matched but not modified")
	}
	return res.ModifiedCount, nil
}

func (r *MongoRepository[T]) Aggregate(ctx context.Context, builder *AggregateBuilder) ([]bson.M, error) {
	pipeline := builder.Build()

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// AggregateBuilder builds complex_query pipelines
type AggregateBuilder struct {
	pipeline mongo.Pipeline
}

func NewAggregateBuilder() *AggregateBuilder {
	return &AggregateBuilder{pipeline: mongo.Pipeline{}}
}

func (ab *AggregateBuilder) Match(filter bson.M) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$match", Value: filter}})
	return ab
}

func (ab *AggregateBuilder) Group(id interface{}, fields bson.M) *AggregateBuilder {
	group := bson.D{{Key: "$group", Value: bson.M{"_id": id}}}
	for k, v := range fields {
		group[0].Value.(bson.M)[k] = v
	}
	ab.pipeline = append(ab.pipeline, group)
	return ab
}

func (ab *AggregateBuilder) Sort(field string, direction int) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: field, Value: direction}}}})
	return ab
}

type SortField struct {
	Field string
	Asc   bool
}

type SortOptions struct{ Fields []SortField }

func Sorts(so *SortOptions) bson.D {
	if so == nil {
		return nil
	}
	d := bson.D{}
	for _, f := range so.Fields {
		v := 1
		if !f.Asc {
			v = -1
		}
		d = append(d, bson.E{Key: f.Field, Value: v})
	}
	return d
}

func (ab *AggregateBuilder) Project(fields bson.M) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$project", Value: fields}})
	return ab
}

func (ab *AggregateBuilder) Limit(n int64) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$limit", Value: n}})
	return ab
}

func (ab *AggregateBuilder) Skip(n int64) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$skip", Value: n}})
	return ab
}

func (ab *AggregateBuilder) Unwind(field string) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$unwind", Value: "$" + field}})
	return ab
}

func (ab *AggregateBuilder) Lookup(from, localField, foreignField, as string) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$lookup", Value: bson.M{
		"from":         from,
		"localField":   localField,
		"foreignField": foreignField,
		"as":           as,
	}}})
	return ab
}

func (ab *AggregateBuilder) Count(alias string) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$count", Value: alias}})
	return ab
}

func (ab *AggregateBuilder) Build() mongo.Pipeline {
	return ab.pipeline
}

func (ab *AggregateBuilder) UnwindPreserveNull(field string) *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$unwind", Value: bson.M{
		"path": "$" + field, "preserveNullAndEmptyArrays": true,
	}}})
	return ab
}

func (ab *AggregateBuilder) ProjectKeep(fields ...string) *AggregateBuilder {
	if len(fields) == 0 {
		return ab
	}
	m := bson.M{}
	for _, f := range fields {
		m[f] = 1
	}
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$project", Value: m}})
	return ab
}

func (ab *AggregateBuilder) ProjectAliases(pairs ...string) *AggregateBuilder {
	if len(pairs)%2 != 0 {
		return ab
	}
	m := bson.M{}
	for i := 0; i < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$addFields", Value: m}})
	return ab
}

func (ab *AggregateBuilder) ExcludeID() *AggregateBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$project", Value: bson.M{"_id": 0}}})
	return ab
}

func (r *MongoRepository[T]) AggregateWithOptions(
	ctx context.Context,
	builder *AggregateBuilder,
	opts *options.AggregateOptions,
) ([]bson.M, error) {
	cursor, err := r.Collection.Aggregate(ctx, builder.Build(), opts)
	if err != nil {
		return nil, err
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *MongoRepository[T]) Count(ctx context.Context, filter bson.M, isDeleted ...*bool) (int64, error) {
	f := cloneFilter(filter)
	applyDeleteFilter(f, "is_deleted", isDeleted...)
	return r.Collection.CountDocuments(ctx, f)
}

func MakePagination(page, pageSize int64) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (page - 1) * pageSize
	return &Pagination{Limit: pageSize, Skip: skip}
}

func (r *MongoRepository[T]) FindWithCount(
	ctx context.Context,
	filter bson.M,
	sort *SortOption,
	pagination *Pagination,
	isDeleted ...*bool,
) ([]T, int64, error) {
	f := cloneFilter(filter)
	applyDeleteFilter(f, "is_deleted", isDeleted...)

	total, err := r.Collection.CountDocuments(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	items, err := r.Find(ctx, f, sort, pagination, isDeleted...)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *MongoRepository[T]) Distinct(
	ctx context.Context,
	field string,
	filter bson.M,
) ([]interface{}, error) {

	result, err := r.Collection.Distinct(ctx, field, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MongoRepository[T]) AggregateRaw(
	ctx context.Context,
	pipeline mongo.Pipeline,
) ([]bson.M, error) {

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func applyDeleteFilter(filter bson.M, fieldName string, isDeleted ...*bool) {
	if filter == nil {
		return
	}
	if _, ok := filter[fieldName]; ok {
		return
	}
	if len(isDeleted) == 0 {
		return
	}
	if isDeleted[0] == nil {
		return
	}
	filter[fieldName] = *isDeleted[0]
}

func cloneFilter(src bson.M) bson.M {
	if src == nil {
		return bson.M{}
	}
	return maps.Clone(src)
}
