package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SampleModel represents a generic model for demonstration purposes
type SampleModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email,omitempty"`
	Category  string             `bson:"category,omitempty"`
	Price     float64            `bson:"price,omitempty"`
	IsActive  bool               `bson:"is_active"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	IsDeleted bool               `bson:"is_deleted,omitempty"`
	DeletedBy string             `bson:"deleted_by,omitempty"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"`
}
