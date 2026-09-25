package service

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

type InventoryService struct {
	stockRepo    repository.StockRepository
	movementRepo repository.StockMovementRepository
	variantRepo  repository.ProductVariantRepository
	branchRepo   repository.BranchRepository
	tx           repository.Transactor
}

func NewInventoryService(
	stockRepo repository.StockRepository,
	movementRepo repository.StockMovementRepository,
	variantRepo repository.ProductVariantRepository,
	branchRepo repository.BranchRepository,
	tx repository.Transactor,
) *InventoryService {
	return &InventoryService{
		stockRepo:    stockRepo,
		movementRepo: movementRepo,
		variantRepo:  variantRepo,
		branchRepo:   branchRepo,
		tx:           tx,
	}
}

func (s *InventoryService) AdjustStock(ctx context.Context, companyID int64, createdBy *int64, req dto.AdjustStockRequest) error {
	// Validate variant exists
	variant, err := s.variantRepo.GetByID(ctx, req.VariantID)
	if err != nil {
		return apperror.InternalError(err)
	}
	if variant == nil {
		return apperror.NotFound("Product variant not found")
	}

	// Validate branch exists and belongs to company
	branch, err := s.branchRepo.GetByID(ctx, req.BranchID)
	if err != nil {
		return apperror.InternalError(err)
	}
	if branch == nil {
		return apperror.NotFound("Branch not found")
	}
	if branch.CompanyID != companyID {
		return apperror.Forbidden("Branch does not belong to your company")
	}

	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		// Lock the current stock: the new quantity is computed from it, so a
		// concurrent change must wait rather than be overwritten
		current, err := s.stockRepo.LockForUpdate(ctx, req.VariantID, req.BranchID)
		if err != nil {
			return apperror.InternalError(err)
		}

		currentQty := 0
		minQty := 0
		if current != nil {
			currentQty = current.Quantity
			minQty = current.MinQuantity
		}

		// Compute new quantity
		var newQty int
		switch req.Type {
		case model.StockMovementIn:
			newQty = currentQty + req.Quantity
		case model.StockMovementOut:
			newQty = currentQty - req.Quantity
			if newQty < 0 {
				return apperror.BadRequest("Insufficient stock")
			}
		case model.StockMovementAdjust:
			newQty = req.Quantity
		default:
			return apperror.BadRequest("Invalid movement type")
		}

		// Create movement record
		movement := &model.StockMovement{
			ProductVariantID: req.VariantID,
			BranchID:         req.BranchID,
			Type:             req.Type,
			Quantity:         req.Quantity,
			StockBefore:      currentQty,
			StockAfter:       newQty,
			UnitCost:         req.UnitCost,
			Note:             req.Note,
			CreatedBy:        createdBy,
		}
		if err := s.movementRepo.Create(ctx, movement); err != nil {
			return apperror.InternalError(err)
		}

		// Upsert stock
		stock := &model.Stock{
			ProductVariantID: req.VariantID,
			BranchID:         req.BranchID,
			Quantity:         newQty,
			MinQuantity:      minQty,
		}
		if err := s.stockRepo.Upsert(ctx, stock); err != nil {
			return apperror.InternalError(err)
		}
		return nil
	})
}

func (s *InventoryService) UpdateMinStock(ctx context.Context, companyID, variantID, branchID int64, req dto.UpdateMinStockRequest) error {
	variant, err := s.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return apperror.InternalError(err)
	}
	if variant == nil {
		return apperror.NotFound("Product variant not found")
	}

	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		return apperror.InternalError(err)
	}
	if branch == nil {
		return apperror.NotFound("Branch not found")
	}
	if branch.CompanyID != companyID {
		return apperror.Forbidden("Branch does not belong to your company")
	}

	if err := s.stockRepo.UpdateMinQuantity(ctx, variantID, branchID, req.MinQuantity); err != nil {
		return apperror.InternalError(err)
	}
	return nil
}

func (s *InventoryService) ListInventory(ctx context.Context, companyID int64, req dto.ListInventoryRequest) (*dto.InventoryListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 20
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	offset := (req.Page - 1) * req.PerPage

	rows, total, err := s.stockRepo.ListInventory(ctx, companyID, req.BranchID, req.Search, req.Category, req.Status, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	items := make([]*dto.InventoryItemResponse, len(rows))
	for i, r := range rows {
		items[i] = toInventoryItemResponse(r)
	}

	return &dto.InventoryListResponse{
		Items:      items,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *InventoryService) GetInventoryStats(ctx context.Context, companyID, branchID int64) (*dto.InventoryStatsResponse, error) {
	stats, err := s.stockRepo.GetInventoryStats(ctx, companyID, branchID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	return &dto.InventoryStatsResponse{
		TotalSKU:        stats.TotalSKU,
		LowStock:        stats.LowStock,
		OutOfStock:      stats.OutOfStock,
		TotalStockValue: stats.TotalStockValue,
	}, nil
}

func (s *InventoryService) ListMovements(ctx context.Context, companyID int64, req dto.ListMovementsRequest) (*dto.MovementListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 20
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	offset := (req.Page - 1) * req.PerPage

	rows, total, err := s.movementRepo.List(ctx, companyID, req.BranchID, req.Type, req.Search, req.StartDate, req.EndDate, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	movements := make([]*dto.StockMovementResponse, len(rows))
	for i, r := range rows {
		movements[i] = toMovementResponse(r)
	}

	return &dto.MovementListResponse{
		Movements:  movements,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *InventoryService) GetMovementStats(ctx context.Context, companyID, branchID int64) (*dto.MovementStatsResponse, error) {
	stats, err := s.movementRepo.GetMonthlyStats(ctx, companyID, branchID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	return &dto.MovementStatsResponse{
		TotalMovements: stats.TotalMovements,
		StockIn:        stats.StockIn,
		StockOut:       stats.StockOut,
		Adjustments:    stats.Adjustments,
	}, nil
}

func buildPagination(total, page, perPage int) *httputil.Pagination {
	totalPages := (total + perPage - 1) / perPage
	var nextPage, prevPage *int
	if page < totalPages {
		next := page + 1
		nextPage = &next
	}
	if page > 1 {
		prev := page - 1
		prevPage = &prev
	}
	return &httputil.Pagination{
		TotalRecords: total,
		TotalPages:   totalPages,
		CurrentPage:  page,
		PerPage:      perPage,
		NextPage:     nextPage,
		PrevPage:     prevPage,
	}
}

func stockStatus(current, min int) dto.StockStatus {
	if current <= 0 {
		return dto.StockStatusOutOfStock
	}
	if current <= min {
		return dto.StockStatusLow
	}
	return dto.StockStatusNormal
}

func toInventoryItemResponse(r *repository.InventoryRow) *dto.InventoryItemResponse {
	return &dto.InventoryItemResponse{
		ProductVariantID: r.ProductVariantID,
		ProductID:        r.ProductID,
		ProductName:      r.ProductName,
		VariantName:      r.VariantName,
		SKU:              r.SKU,
		CategoryName:     r.CategoryName,
		ImageData:        r.ImageData,
		BranchID:         r.BranchID,
		BranchName:       r.BranchName,
		CurrentStock:     r.CurrentStock,
		MinStock:         r.MinStock,
		StockValue:       r.StockValue,
		Status:           stockStatus(r.CurrentStock, r.MinStock),
		HasVariants:      r.HasVariants,
	}
}

func toMovementResponse(r *repository.MovementRow) *dto.StockMovementResponse {
	return &dto.StockMovementResponse{
		ID:               r.ID,
		ProductVariantID: r.ProductVariantID,
		ProductName:      r.ProductName,
		VariantName:      r.VariantName,
		SKU:              r.SKU,
		BranchID:         r.BranchID,
		BranchName:       r.BranchName,
		Type:             r.Type,
		Quantity:         r.Quantity,
		StockBefore:      r.StockBefore,
		StockAfter:       r.StockAfter,
		UnitCost:         r.UnitCost,
		ReferenceType:    r.ReferenceType,
		ReferenceID:      r.ReferenceID,
		Note:             r.Note,
		CreatedBy:        r.CreatedBy,
		CreatedByName:    r.CreatedByName,
		CreatedAt:        r.CreatedAt,
	}
}
