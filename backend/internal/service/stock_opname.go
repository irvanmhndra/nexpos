package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type StockOpnameService struct {
	opnameRepo   repository.StockOpnameRepository
	stockRepo    repository.StockRepository
	movementRepo repository.StockMovementRepository
	branchRepo   repository.BranchRepository
}

func NewStockOpnameService(
	opnameRepo repository.StockOpnameRepository,
	stockRepo repository.StockRepository,
	movementRepo repository.StockMovementRepository,
	branchRepo repository.BranchRepository,
) *StockOpnameService {
	return &StockOpnameService{
		opnameRepo:   opnameRepo,
		stockRepo:    stockRepo,
		movementRepo: movementRepo,
		branchRepo:   branchRepo,
	}
}

func (s *StockOpnameService) Create(ctx context.Context, companyID, userID int64, req dto.CreateStockOpnameRequest) (*dto.StockOpnameResponse, error) {
	branch, err := s.branchRepo.GetByID(ctx, req.BranchID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if branch == nil {
		return nil, apperror.NotFound("Branch not found")
	}
	if branch.CompanyID != companyID {
		return nil, apperror.Forbidden("Branch does not belong to your company")
	}

	number, err := s.opnameRepo.GenerateOpnameNumber(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	startedBy := &userID
	opname := &model.StockOpname{
		CompanyID:    companyID,
		BranchID:     req.BranchID,
		OpnameNumber: number,
		Status:       model.StockOpnameStatusInProgress,
		Notes:        req.Notes,
		StartedBy:    startedBy,
		StartedAt:    time.Now(),
	}
	if err := s.opnameRepo.Create(ctx, opname); err != nil {
		return nil, apperror.InternalError(err)
	}

	count, err := s.opnameRepo.SnapshotItems(ctx, opname.ID, companyID, req.BranchID, req.CategoryID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if count == 0 {
		return nil, apperror.BadRequest("No active products found to snapshot for this branch")
	}

	return s.GetByID(ctx, companyID, opname.ID)
}

func (s *StockOpnameService) GetByID(ctx context.Context, companyID, id int64) (*dto.StockOpnameResponse, error) {
	opname, err := s.opnameRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if opname == nil {
		return nil, apperror.NotFound("Stock opname not found")
	}

	items, err := s.opnameRepo.GetItems(ctx, opname.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	opname.Items = items

	total, counted, err := s.opnameRepo.GetItemStats(ctx, opname.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	return toOpnameResponse(opname, total, counted), nil
}

func (s *StockOpnameService) List(ctx context.Context, companyID int64, req dto.ListStockOpnameRequest) (*dto.StockOpnameListResponse, error) {
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

	params := &repository.StockOpnameListParams{
		Status:   req.Status,
		BranchID: req.BranchID,
		Limit:    req.PerPage,
		Offset:   offset,
	}
	opnames, total, err := s.opnameRepo.List(ctx, companyID, params)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.StockOpnameResponse, len(opnames))
	for i, op := range opnames {
		tot, counted, err := s.opnameRepo.GetItemStats(ctx, op.ID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		responses[i] = toOpnameResponse(op, tot, counted)
	}

	return &dto.StockOpnameListResponse{
		Opnames:    responses,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *StockOpnameService) UpdateItem(ctx context.Context, companyID, opnameID, itemID, userID int64, req dto.UpdateOpnameItemRequest) (*dto.StockOpnameItemResponse, error) {
	opname, err := s.opnameRepo.GetByID(ctx, companyID, opnameID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if opname == nil {
		return nil, apperror.NotFound("Stock opname not found")
	}
	if opname.Status != model.StockOpnameStatusInProgress {
		return nil, apperror.BadRequest("Cannot edit items on a completed or cancelled opname")
	}

	item, err := s.opnameRepo.GetItem(ctx, opnameID, itemID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if item == nil {
		return nil, apperror.NotFound("Stock opname item not found")
	}

	counted := req.CountedStock
	now := time.Now()
	item.CountedStock = &counted
	item.VarianceQty = counted - item.SystemStock
	item.VarianceValue = float64(item.VarianceQty) * item.UnitCost
	item.Notes = req.Notes
	item.CountedBy = &userID
	item.CountedAt = &now

	if err := s.opnameRepo.UpdateItem(ctx, item); err != nil {
		return nil, apperror.InternalError(err)
	}

	return toOpnameItemResponse(item), nil
}

func (s *StockOpnameService) BulkUpdateItems(ctx context.Context, companyID, opnameID, userID int64, req dto.BulkUpdateOpnameItemsRequest) (*dto.StockOpnameResponse, error) {
	opname, err := s.opnameRepo.GetByID(ctx, companyID, opnameID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if opname == nil {
		return nil, apperror.NotFound("Stock opname not found")
	}
	if opname.Status != model.StockOpnameStatusInProgress {
		return nil, apperror.BadRequest("Cannot edit items on a completed or cancelled opname")
	}

	now := time.Now()
	for _, line := range req.Items {
		item, err := s.opnameRepo.GetItem(ctx, opnameID, line.ItemID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if item == nil {
			return nil, apperror.BadRequest("Stock opname item not found")
		}

		counted := line.CountedStock
		item.CountedStock = &counted
		item.VarianceQty = counted - item.SystemStock
		item.VarianceValue = float64(item.VarianceQty) * item.UnitCost
		item.Notes = line.Notes
		item.CountedBy = &userID
		item.CountedAt = &now

		if err := s.opnameRepo.UpdateItem(ctx, item); err != nil {
			return nil, apperror.InternalError(err)
		}
	}

	return s.GetByID(ctx, companyID, opnameID)
}

func (s *StockOpnameService) Complete(ctx context.Context, companyID, opnameID, userID int64, req dto.CompleteStockOpnameRequest) (*dto.StockOpnameResponse, error) {
	opname, err := s.opnameRepo.GetByID(ctx, companyID, opnameID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if opname == nil {
		return nil, apperror.NotFound("Stock opname not found")
	}
	if opname.Status != model.StockOpnameStatusInProgress {
		return nil, apperror.BadRequest("Stock opname is not in progress")
	}

	items, err := s.opnameRepo.GetItems(ctx, opname.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	var totalVarianceQty int
	var totalVarianceValue float64
	refType := "stock_opname"

	for _, item := range items {
		if item.CountedStock == nil {
			continue
		}
		counted := *item.CountedStock

		current, err := s.stockRepo.GetByVariantAndBranch(ctx, item.ProductVariantID, opname.BranchID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		currentQty := 0
		minQty := 0
		if current != nil {
			currentQty = current.Quantity
			minQty = current.MinQuantity
		}

		if counted != currentQty {
			delta := counted - currentQty
			unitCost := item.UnitCost
			movement := &model.StockMovement{
				ProductVariantID: item.ProductVariantID,
				BranchID:         opname.BranchID,
				Type:             model.StockMovementAdjust,
				Quantity:         absInt(delta),
				StockBefore:      currentQty,
				StockAfter:       counted,
				UnitCost:         &unitCost,
				ReferenceType:    &refType,
				ReferenceID:      &opname.ID,
				CreatedBy:        &userID,
			}
			if err := s.movementRepo.Create(ctx, movement); err != nil {
				return nil, apperror.InternalError(err)
			}

			stock := &model.Stock{
				ProductVariantID: item.ProductVariantID,
				BranchID:         opname.BranchID,
				Quantity:         counted,
				MinQuantity:      minQty,
			}
			if err := s.stockRepo.Upsert(ctx, stock); err != nil {
				return nil, apperror.InternalError(err)
			}
		}

		totalVarianceQty += item.VarianceQty
		totalVarianceValue += item.VarianceValue
	}

	now := time.Now()
	opname.Status = model.StockOpnameStatusCompleted
	opname.CompletedBy = &userID
	opname.CompletedAt = &now
	opname.TotalVarianceQty = totalVarianceQty
	opname.TotalVarianceValue = totalVarianceValue
	if req.Notes != nil {
		opname.Notes = req.Notes
	}

	if err := s.opnameRepo.Update(ctx, opname); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.GetByID(ctx, companyID, opname.ID)
}

func (s *StockOpnameService) Cancel(ctx context.Context, companyID, opnameID int64) (*dto.StockOpnameResponse, error) {
	opname, err := s.opnameRepo.GetByID(ctx, companyID, opnameID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if opname == nil {
		return nil, apperror.NotFound("Stock opname not found")
	}
	if opname.Status != model.StockOpnameStatusInProgress {
		return nil, apperror.BadRequest("Only in-progress opnames can be cancelled")
	}

	now := time.Now()
	opname.Status = model.StockOpnameStatusCancelled
	opname.CancelledAt = &now

	if err := s.opnameRepo.Update(ctx, opname); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.GetByID(ctx, companyID, opname.ID)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func toOpnameResponse(op *model.StockOpname, totalItems, countedItems int) *dto.StockOpnameResponse {
	resp := &dto.StockOpnameResponse{
		ID:                 op.ID,
		BranchID:           op.BranchID,
		OpnameNumber:       op.OpnameNumber,
		Status:             op.Status,
		Notes:              op.Notes,
		TotalItems:         totalItems,
		CountedItems:       countedItems,
		TotalVarianceQty:   op.TotalVarianceQty,
		TotalVarianceValue: op.TotalVarianceValue,
		StartedBy:          op.StartedBy,
		CompletedBy:        op.CompletedBy,
		StartedAt:          op.StartedAt,
		CompletedAt:        op.CompletedAt,
		CancelledAt:        op.CancelledAt,
		CreatedAt:          op.CreatedAt,
		UpdatedAt:          op.UpdatedAt,
	}
	if op.Items != nil {
		resp.Items = make([]*dto.StockOpnameItemResponse, len(op.Items))
		for i, it := range op.Items {
			resp.Items[i] = toOpnameItemResponse(it)
		}
	}
	return resp
}

func toOpnameItemResponse(it *model.StockOpnameItem) *dto.StockOpnameItemResponse {
	return &dto.StockOpnameItemResponse{
		ID:               it.ID,
		StockOpnameID:    it.StockOpnameID,
		ProductVariantID: it.ProductVariantID,
		SKU:              it.SKU,
		ProductName:      it.ProductName,
		VariantName:      it.VariantName,
		SystemStock:      it.SystemStock,
		CountedStock:     it.CountedStock,
		VarianceQty:      it.VarianceQty,
		UnitCost:         it.UnitCost,
		VarianceValue:    it.VarianceValue,
		Notes:            it.Notes,
		CountedAt:        it.CountedAt,
	}
}
