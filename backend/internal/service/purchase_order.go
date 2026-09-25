package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type PurchaseOrderService struct {
	supplierRepo repository.SupplierRepository
	poRepo       repository.PurchaseOrderRepository
	stockRepo    repository.StockRepository
	movementRepo repository.StockMovementRepository
	variantRepo  repository.ProductVariantRepository
	tx           repository.Transactor
}

func NewPurchaseOrderService(
	supplierRepo repository.SupplierRepository,
	poRepo repository.PurchaseOrderRepository,
	stockRepo repository.StockRepository,
	movementRepo repository.StockMovementRepository,
	variantRepo repository.ProductVariantRepository,
	tx repository.Transactor,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		supplierRepo: supplierRepo,
		poRepo:       poRepo,
		stockRepo:    stockRepo,
		movementRepo: movementRepo,
		variantRepo:  variantRepo,
		tx:           tx,
	}
}

// ============== Supplier Methods ==============

func (s *PurchaseOrderService) CreateSupplier(ctx context.Context, companyID int64, req dto.CreateSupplierRequest) (*dto.SupplierResponse, error) {
	existing, err := s.supplierRepo.GetByCode(ctx, companyID, req.Code)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if existing != nil {
		return nil, apperror.BadRequest("Supplier code already exists")
	}

	supplier := &model.Supplier{
		CompanyID:   companyID,
		Name:        req.Name,
		Code:        req.Code,
		ContactName: req.ContactName,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		Notes:       req.Notes,
		IsActive:    req.IsActive,
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, apperror.InternalError(err)
	}

	return toSupplierResponse(supplier), nil
}

func (s *PurchaseOrderService) GetSupplier(ctx context.Context, companyID, id int64) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if supplier == nil {
		return nil, apperror.NotFound("Supplier not found")
	}
	return toSupplierResponse(supplier), nil
}

func (s *PurchaseOrderService) ListSuppliers(ctx context.Context, companyID int64, req dto.ListSupplierRequest) (*dto.SupplierListResponse, error) {
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

	suppliers, total, err := s.supplierRepo.List(ctx, companyID, req.Search, req.IsActive, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.SupplierResponse, len(suppliers))
	for i, sup := range suppliers {
		responses[i] = toSupplierResponse(sup)
	}

	return &dto.SupplierListResponse{
		Suppliers:  responses,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *PurchaseOrderService) UpdateSupplier(ctx context.Context, companyID, id int64, req dto.UpdateSupplierRequest) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if supplier == nil {
		return nil, apperror.NotFound("Supplier not found")
	}

	existing, err := s.supplierRepo.GetByCode(ctx, companyID, req.Code)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if existing != nil && existing.ID != id {
		return nil, apperror.BadRequest("Supplier code already exists")
	}

	supplier.Name = req.Name
	supplier.Code = req.Code
	supplier.ContactName = req.ContactName
	supplier.Phone = req.Phone
	supplier.Email = req.Email
	supplier.Address = req.Address
	supplier.Notes = req.Notes
	supplier.IsActive = req.IsActive

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, apperror.InternalError(err)
	}

	supplier, _ = s.supplierRepo.GetByID(ctx, companyID, id)
	return toSupplierResponse(supplier), nil
}

func (s *PurchaseOrderService) DeleteSupplier(ctx context.Context, companyID, id int64) error {
	supplier, err := s.supplierRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if supplier == nil {
		return apperror.NotFound("Supplier not found")
	}

	if err := s.supplierRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}
	return nil
}

// ============== Purchase Order Methods ==============

func (s *PurchaseOrderService) CreatePO(ctx context.Context, companyID, branchID, userID int64, req dto.CreatePurchaseOrderRequest) (*dto.PurchaseOrderResponse, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, companyID, req.SupplierID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if supplier == nil {
		return nil, apperror.NotFound("Supplier not found")
	}

	poNumber, err := s.poRepo.GeneratePONumber(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	now := time.Now()
	po := &model.PurchaseOrder{
		CompanyID:  companyID,
		BranchID:   branchID,
		SupplierID: req.SupplierID,
		PONumber:   poNumber,
		Status:     model.POStatusDraft,
		Notes:      req.Notes,
		OrderedAt:  &now,
	}

	var totalAmount float64
	items := make([]*model.PurchaseOrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		subtotal := float64(itemReq.Quantity) * itemReq.UnitCost
		totalAmount += subtotal
		items[i] = &model.PurchaseOrderItem{
			ProductVariantID: itemReq.ProductVariantID,
			SKU:              itemReq.SKU,
			VariantName:      itemReq.VariantName,
			Quantity:         itemReq.Quantity,
			UnitCost:         itemReq.UnitCost,
			Subtotal:         subtotal,
		}
	}
	po.TotalAmount = totalAmount

	if err := s.poRepo.Create(ctx, po); err != nil {
		return nil, apperror.InternalError(err)
	}

	for _, item := range items {
		item.PurchaseOrderID = po.ID
		if err := s.poRepo.CreateItem(ctx, item); err != nil {
			return nil, apperror.InternalError(err)
		}
	}

	po.Items = items
	return toPOResponse(po), nil
}

func (s *PurchaseOrderService) GetPO(ctx context.Context, companyID, id int64) (*dto.PurchaseOrderResponse, error) {
	po, err := s.poRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if po == nil {
		return nil, apperror.NotFound("Purchase order not found")
	}

	items, err := s.poRepo.GetItems(ctx, po.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	po.Items = items

	return toPOResponse(po), nil
}

func (s *PurchaseOrderService) ListPOs(ctx context.Context, companyID int64, req dto.ListPurchaseOrderRequest) (*dto.PurchaseOrderListResponse, error) {
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

	params := &repository.POListParams{
		Status:     req.Status,
		SupplierID: req.SupplierID,
		Limit:      req.PerPage,
		Offset:     offset,
	}

	pos, total, err := s.poRepo.List(ctx, companyID, params)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.PurchaseOrderResponse, len(pos))
	for i, po := range pos {
		responses[i] = toPOResponse(po)
	}

	return &dto.PurchaseOrderListResponse{
		PurchaseOrders: responses,
		Pagination:     buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *PurchaseOrderService) ReceivePO(ctx context.Context, companyID, poID int64, req dto.ReceivePurchaseOrderRequest) (*dto.PurchaseOrderResponse, error) {
	// Lock the PO and each stock row: a PO received twice concurrently, or a
	// receipt racing a sale, must not double-count or lose stock
	var po *model.PurchaseOrder
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		po, err = s.poRepo.GetByIDForUpdate(ctx, companyID, poID)
		if err != nil {
			return apperror.InternalError(err)
		}
		if po == nil {
			return apperror.NotFound("Purchase order not found")
		}
		if po.Status == model.POStatusCancelled || po.Status == model.POStatusReceived {
			return apperror.BadRequest("Purchase order cannot be received in its current status")
		}

		items, err := s.poRepo.GetItems(ctx, po.ID)
		if err != nil {
			return apperror.InternalError(err)
		}

		itemMap := make(map[int64]*model.PurchaseOrderItem)
		for _, item := range items {
			itemMap[item.ID] = item
		}

		allReceived := true
		anyReceived := false

		for _, receiveReq := range req.Items {
			item, ok := itemMap[receiveReq.ItemID]
			if !ok {
				return apperror.BadRequest("Purchase order item not found")
			}

			if receiveReq.ReceivedQuantity < 0 {
				return apperror.BadRequest("Received quantity cannot be negative")
			}

			totalReceived := item.ReceivedQuantity + receiveReq.ReceivedQuantity
			if totalReceived > item.Quantity {
				totalReceived = item.Quantity
			}

			if receiveReq.ReceivedQuantity > 0 && item.ProductVariantID != nil {
				current, err := s.stockRepo.LockForUpdate(ctx, *item.ProductVariantID, po.BranchID)
				if err != nil {
					return apperror.InternalError(err)
				}

				currentQty := 0
				minQty := 0
				if current != nil {
					currentQty = current.Quantity
					minQty = current.MinQuantity
				}

				newQty := currentQty + receiveReq.ReceivedQuantity
				refType := "purchase_order"
				movement := &model.StockMovement{
					ProductVariantID: *item.ProductVariantID,
					BranchID:         po.BranchID,
					Type:             model.StockMovementIn,
					Quantity:         receiveReq.ReceivedQuantity,
					StockBefore:      currentQty,
					StockAfter:       newQty,
					UnitCost:         &item.UnitCost,
					ReferenceType:    &refType,
					ReferenceID:      &po.ID,
				}
				if err := s.movementRepo.Create(ctx, movement); err != nil {
					return apperror.InternalError(err)
				}

				stock := &model.Stock{
					ProductVariantID: *item.ProductVariantID,
					BranchID:         po.BranchID,
					Quantity:         newQty,
					MinQuantity:      minQty,
				}
				if err := s.stockRepo.Upsert(ctx, stock); err != nil {
					return apperror.InternalError(err)
				}

				variant, err := s.variantRepo.GetByID(ctx, *item.ProductVariantID)
				if err != nil {
					return apperror.InternalError(err)
				}
				if variant != nil && item.UnitCost > 0 {
					variant.LastPurchaseCost = item.UnitCost
					if err := s.variantRepo.Update(ctx, variant); err != nil {
						return apperror.InternalError(err)
					}
				}
			}

			item.ReceivedQuantity = totalReceived
			if err := s.poRepo.UpdateItem(ctx, item); err != nil {
				return apperror.InternalError(err)
			}

			if totalReceived < item.Quantity {
				allReceived = false
			}
			if totalReceived > 0 {
				anyReceived = true
			}
		}

		for _, item := range items {
			if item.ReceivedQuantity < item.Quantity {
				allReceived = false
			}
			if item.ReceivedQuantity > 0 {
				anyReceived = true
			}
		}

		now := time.Now()
		if allReceived {
			po.Status = model.POStatusReceived
			po.ReceivedAt = &now
		} else if anyReceived {
			po.Status = model.POStatusPartial
		}

		if err := s.poRepo.Update(ctx, po); err != nil {
			return apperror.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	items, _ := s.poRepo.GetItems(ctx, po.ID)
	po.Items = items

	return toPOResponse(po), nil
}

// ============== Helper Methods ==============

func toSupplierResponse(s *model.Supplier) *dto.SupplierResponse {
	return &dto.SupplierResponse{
		ID:          s.ID,
		Name:        s.Name,
		Code:        s.Code,
		ContactName: s.ContactName,
		Phone:       s.Phone,
		Email:       s.Email,
		Address:     s.Address,
		Notes:       s.Notes,
		IsActive:    s.IsActive,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toPOResponse(po *model.PurchaseOrder) *dto.PurchaseOrderResponse {
	resp := &dto.PurchaseOrderResponse{
		ID:          po.ID,
		BranchID:    po.BranchID,
		SupplierID:  po.SupplierID,
		PONumber:    po.PONumber,
		Status:      po.Status,
		Notes:       po.Notes,
		TotalAmount: po.TotalAmount,
		OrderedAt:   po.OrderedAt,
		ReceivedAt:  po.ReceivedAt,
		CancelledAt: po.CancelledAt,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}

	if po.Items != nil {
		resp.Items = make([]*dto.POItemResponse, len(po.Items))
		for i, item := range po.Items {
			resp.Items[i] = &dto.POItemResponse{
				ID:               item.ID,
				ProductVariantID: item.ProductVariantID,
				SKU:              item.SKU,
				VariantName:      item.VariantName,
				Quantity:         item.Quantity,
				ReceivedQuantity: item.ReceivedQuantity,
				UnitCost:         item.UnitCost,
				Subtotal:         item.Subtotal,
			}
		}
	}

	return resp
}
