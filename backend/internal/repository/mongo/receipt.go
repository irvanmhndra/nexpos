package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const receiptCollection = "receipts"

type receiptRepository struct {
	col *mongo.Collection
}

// NewReceiptRepository creates a MongoDB-backed ReceiptRepository and ensures indexes.
func NewReceiptRepository(db *mongo.Database) (repository.ReceiptRepository, error) {
	col := db.Collection(receiptCollection)
	if err := ensureReceiptIndexes(col); err != nil {
		return nil, fmt.Errorf("receipt: ensure indexes: %w", err)
	}
	return &receiptRepository{col: col}, nil
}

func ensureReceiptIndexes(col *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "order_id", Value: 1}, {Key: "company_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "company_id", Value: 1}},
		},
	})
	return err
}

func (r *receiptRepository) Save(ctx context.Context, rec *model.Receipt) error {
	_, err := r.col.InsertOne(ctx, rec)
	return err
}

func (r *receiptRepository) GetByOrderID(ctx context.Context, companyID, orderID int64) (*model.Receipt, error) {
	var rec model.Receipt
	err := r.col.FindOne(ctx, bson.M{
		"order_id":   orderID,
		"company_id": companyID,
	}).Decode(&rec)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &rec, err
}
