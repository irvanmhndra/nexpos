package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type promotionRepository struct {
	db *sqlx.DB
}

func NewPromotionRepository(db *sqlx.DB) repository.PromotionRepository {
	return &promotionRepository{db: db}
}

func (r *promotionRepository) Create(ctx context.Context, promotion *model.Promotion) error {
	query := `
		INSERT INTO promotions (
			company_id, code, name, description, type, discount_type, discount_value,
			min_purchase, max_discount, start_at, end_at, priority, is_active,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowxContext(ctx, query,
		promotion.CompanyID,
		promotion.Code,
		promotion.Name,
		promotion.Description,
		promotion.Type,
		promotion.DiscountType,
		promotion.DiscountValue,
		promotion.MinPurchase,
		promotion.MaxDiscount,
		promotion.StartAt,
		promotion.EndAt,
		promotion.Priority,
		promotion.IsActive,
	).Scan(&promotion.ID, &promotion.CreatedAt, &promotion.UpdatedAt)
}

func (r *promotionRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Promotion, error) {
	var promo model.Promotion
	query := `SELECT * FROM promotions WHERE id = $1 AND company_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &promo, query, id, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &promo, nil
}

func (r *promotionRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, promoType, status string, limit, offset int) ([]*model.Promotion, int, error) {
	args := []any{companyID}
	conditions := []string{"company_id = $1"}
	argIdx := 2

	if search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE $%d OR code ILIKE $%d OR description ILIKE $%d)",
			argIdx, argIdx+1, argIdx+2,
		))
		like := "%" + search + "%"
		args = append(args, like, like, like)
		argIdx += 3
	}

	if isActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *isActive)
		argIdx++
	}

	if promoType != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, promoType)
		argIdx++
	}

	// Derived-status filter, evaluated against server time (NOW()) — matches how
	// promotions are actually applied at order time.
	switch status {
	case model.PromotionStatusInactive:
		conditions = append(conditions, "is_active = false")
	case model.PromotionStatusScheduled:
		conditions = append(conditions, "is_active = true AND start_at > NOW()")
	case model.PromotionStatusActive:
		conditions = append(conditions, "is_active = true AND start_at <= NOW() AND (end_at IS NULL OR end_at >= NOW())")
	case model.PromotionStatusExpired:
		conditions = append(conditions, "is_active = true AND end_at IS NOT NULL AND end_at < NOW()")
	}

	where := strings.Join(conditions, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM promotions WHERE %s", where)
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append(args, limit, offset)
	listQuery := fmt.Sprintf(
		"SELECT * FROM promotions WHERE %s ORDER BY priority DESC, created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1,
	)

	var promotions []*model.Promotion
	if err := conn(ctx, r.db).SelectContext(ctx, &promotions, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return promotions, total, nil
}

func (r *promotionRepository) Update(ctx context.Context, promotion *model.Promotion) error {
	query := `
		UPDATE promotions
		SET code = $1, name = $2, description = $3, type = $4, discount_type = $5,
		    discount_value = $6, min_purchase = $7, max_discount = $8, start_at = $9,
		    end_at = $10, priority = $11, is_active = $12, updated_at = NOW()
		WHERE id = $13 AND company_id = $14
		RETURNING updated_at
	`
	return conn(ctx, r.db).QueryRowxContext(ctx, query,
		promotion.Code,
		promotion.Name,
		promotion.Description,
		promotion.Type,
		promotion.DiscountType,
		promotion.DiscountValue,
		promotion.MinPurchase,
		promotion.MaxDiscount,
		promotion.StartAt,
		promotion.EndAt,
		promotion.Priority,
		promotion.IsActive,
		promotion.ID,
		promotion.CompanyID,
	).Scan(&promotion.UpdatedAt)
}

func (r *promotionRepository) Delete(ctx context.Context, companyID, id int64) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `DELETE FROM promotions WHERE id = $1 AND company_id = $2`, id, companyID)
	return err
}

func (r *promotionRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM promotions WHERE company_id = $1 AND code = $2 AND id != $3`
	err := conn(ctx, r.db).GetContext(ctx, &count, query, companyID, code, excludeID)
	return count > 0, err
}

func (r *promotionRepository) GetActivePromotions(ctx context.Context, companyID int64, now time.Time) ([]*model.Promotion, error) {
	var promotions []*model.Promotion
	query := `
		SELECT * FROM promotions
		WHERE company_id = $1
		  AND is_active = true
		  AND start_at <= $2
		  AND (end_at IS NULL OR end_at >= $2)
		ORDER BY priority DESC, created_at DESC
	`
	err := conn(ctx, r.db).SelectContext(ctx, &promotions, query, companyID, now)
	return promotions, err
}

func (r *promotionRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.Promotion, error) {
	var promo model.Promotion
	query := `SELECT * FROM promotions WHERE company_id = $1 AND code = $2`
	err := conn(ctx, r.db).GetContext(ctx, &promo, query, companyID, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &promo, nil
}
