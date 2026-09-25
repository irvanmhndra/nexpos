package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type CompanySettingsRepository struct {
	db *sqlx.DB
}

func NewCompanySettingsRepository(db *sqlx.DB) *CompanySettingsRepository {
	return &CompanySettingsRepository{db: db}
}

func (r *CompanySettingsRepository) GetByCompanyID(ctx context.Context, companyID int64) (*model.CompanySettings, error) {
	var settings model.CompanySettings
	query := `
		SELECT id, company_id, tax_enabled, tax_rate, tax_inclusive,
			rounding_enabled, rounding_amount, auto_complete_counter_orders,
			require_customer_for_delivery, receipt_header, receipt_footer,
			show_tax_on_receipt, offline_mode_enabled, max_offline_days,
			created_at, updated_at
		FROM company_settings
		WHERE company_id = $1
	`
	err := conn(ctx, r.db).GetContext(ctx, &settings, query, companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default settings if not found
			return &model.CompanySettings{
				CompanyID:                  companyID,
				TaxEnabled:                 false,
				TaxRate:                    0,
				TaxInclusive:               true,
				RoundingEnabled:            false,
				RoundingAmount:             0,
				AutoCompleteCounterOrders:  true,
				RequireCustomerForDelivery: true,
				ShowTaxOnReceipt:           true,
				OfflineModeEnabled:         false,
				MaxOfflineDays:             7,
			}, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *CompanySettingsRepository) Upsert(ctx context.Context, settings *model.CompanySettings) error {
	query := `
		INSERT INTO company_settings (
			company_id, tax_enabled, tax_rate, tax_inclusive,
			rounding_enabled, rounding_amount, auto_complete_counter_orders,
			require_customer_for_delivery, receipt_header, receipt_footer,
			show_tax_on_receipt, offline_mode_enabled, max_offline_days
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (company_id) DO UPDATE SET
			tax_enabled = EXCLUDED.tax_enabled,
			tax_rate = EXCLUDED.tax_rate,
			tax_inclusive = EXCLUDED.tax_inclusive,
			rounding_enabled = EXCLUDED.rounding_enabled,
			rounding_amount = EXCLUDED.rounding_amount,
			auto_complete_counter_orders = EXCLUDED.auto_complete_counter_orders,
			require_customer_for_delivery = EXCLUDED.require_customer_for_delivery,
			receipt_header = EXCLUDED.receipt_header,
			receipt_footer = EXCLUDED.receipt_footer,
			show_tax_on_receipt = EXCLUDED.show_tax_on_receipt,
			offline_mode_enabled = EXCLUDED.offline_mode_enabled,
			max_offline_days = EXCLUDED.max_offline_days,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		settings.CompanyID,
		settings.TaxEnabled,
		settings.TaxRate,
		settings.TaxInclusive,
		settings.RoundingEnabled,
		settings.RoundingAmount,
		settings.AutoCompleteCounterOrders,
		settings.RequireCustomerForDelivery,
		settings.ReceiptHeader,
		settings.ReceiptFooter,
		settings.ShowTaxOnReceipt,
		settings.OfflineModeEnabled,
		settings.MaxOfflineDays,
	).Scan(&settings.ID, &settings.CreatedAt, &settings.UpdatedAt)
}
