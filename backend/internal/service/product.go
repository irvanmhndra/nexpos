package service

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type ProductService struct {
	productRepo  repository.ProductRepository
	variantRepo  repository.ProductVariantRepository
	categoryRepo repository.ProductCategoryRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	variantRepo repository.ProductVariantRepository,
	categoryRepo repository.ProductCategoryRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		variantRepo:  variantRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, companyID int64, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Validate category if provided
	if req.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, companyID, *req.CategoryID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if category == nil {
			return nil, apperror.BadRequest("Category not found")
		}
	}

	// Validate SKUs don't exist
	for _, v := range req.Variants {
		exists, err := s.variantRepo.SKUExists(ctx, v.SKU, 0)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if exists {
			return nil, apperror.BadRequest("SKU already exists: " + v.SKU)
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	product := &model.Product{
		CompanyID:         companyID,
		ProductCategoryID: req.CategoryID,
		Name:              req.Name,
		Description:       req.Description,
		ImageData:         req.ImageData,
		IsActive:          isActive,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Create variants
	hasDefault := false
	for i, v := range req.Variants {
		variantIsActive := true
		if v.IsActive != nil {
			variantIsActive = *v.IsActive
		}

		// First variant is default if none specified
		isDefault := v.IsDefault
		if i == 0 && !hasDefault {
			isDefault = true
		}
		if v.IsDefault {
			hasDefault = true
		}

		variant := &model.ProductVariant{
			ProductID:        product.ID,
			SKU:              v.SKU,
			Name:             v.Name,
			Attributes:       model.JSONMap(v.Attributes),
			Price:            v.Price,
			StandardCost:     v.StandardCost,
			LastPurchaseCost: v.LastPurchaseCost,
			IsDefault:        isDefault,
			IsActive:         variantIsActive,
		}

		if err := s.variantRepo.Create(ctx, variant); err != nil {
			return nil, apperror.InternalError(err)
		}
	}

	return s.GetByID(ctx, companyID, product.ID)
}

func (s *ProductService) GetByID(ctx context.Context, companyID, id int64) (*dto.ProductResponse, error) {
	product, err := s.productRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if product == nil {
		return nil, apperror.NotFound("Product not found")
	}

	variants, err := s.variantRepo.GetByProductID(ctx, product.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(ctx, companyID, product, variants), nil
}

func (s *ProductService) List(ctx context.Context, companyID int64, req dto.ListProductRequest) (*dto.ProductListResponse, error) {
	// Set defaults
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

	products, total, err := s.productRepo.List(ctx, companyID, req.Search, req.CategoryID, req.IsActive, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Build response with variants
	responses := make([]*dto.ProductResponse, len(products))
	for i, p := range products {
		variants, err := s.variantRepo.GetByProductID(ctx, p.ID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		responses[i] = s.toResponse(ctx, companyID, p, variants)
	}

	// Calculate pagination
	totalPages := (total + req.PerPage - 1) / req.PerPage
	var nextPage, prevPage *int
	if req.Page < totalPages {
		next := req.Page + 1
		nextPage = &next
	}
	if req.Page > 1 {
		prev := req.Page - 1
		prevPage = &prev
	}

	return &dto.ProductListResponse{
		Products: responses,
		Pagination: &dto.PaginationMeta{
			TotalRecords: total,
			TotalPages:   totalPages,
			CurrentPage:  req.Page,
			PerPage:      req.PerPage,
			NextPage:     nextPage,
			PrevPage:     prevPage,
		},
	}, nil
}

func (s *ProductService) Update(ctx context.Context, companyID, id int64, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	// Check if product exists
	product, err := s.productRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if product == nil {
		return nil, apperror.NotFound("Product not found")
	}

	// Validate category if provided
	if req.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, companyID, *req.CategoryID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if category == nil {
			return nil, apperror.BadRequest("Category not found")
		}
	}

	// Validate SKUs don't exist (excluding current variants)
	existingVariants, err := s.variantRepo.GetByProductID(ctx, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	existingIDs := make(map[int64]bool)
	for _, v := range existingVariants {
		existingIDs[v.ID] = true
	}

	for _, v := range req.Variants {
		excludeID := int64(0)
		if v.ID != nil {
			excludeID = *v.ID
		}
		exists, err := s.variantRepo.SKUExists(ctx, v.SKU, excludeID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if exists {
			return nil, apperror.BadRequest("SKU already exists: " + v.SKU)
		}
	}

	// Update product
	product.ProductCategoryID = req.CategoryID
	product.Name = req.Name
	product.Description = req.Description
	product.ImageData = req.ImageData
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Track which variant IDs are being updated
	updatedIDs := make(map[int64]bool)

	// Update or create variants
	hasDefault := false
	for i, v := range req.Variants {
		variantIsActive := true
		if v.IsActive != nil {
			variantIsActive = *v.IsActive
		}

		isDefault := v.IsDefault
		if i == 0 && !hasDefault {
			isDefault = true
		}
		if v.IsDefault {
			hasDefault = true
		}

		if v.ID != nil && existingIDs[*v.ID] {
			// Update existing variant
			variant := &model.ProductVariant{
				ID:               *v.ID,
				ProductID:        id,
				SKU:              v.SKU,
				Name:             v.Name,
				Attributes:       model.JSONMap(v.Attributes),
				Price:            v.Price,
				StandardCost:     v.StandardCost,
				LastPurchaseCost: v.LastPurchaseCost,
				IsDefault:        isDefault,
				IsActive:         variantIsActive,
			}
			if err := s.variantRepo.Update(ctx, variant); err != nil {
				return nil, apperror.InternalError(err)
			}
			updatedIDs[*v.ID] = true
		} else {
			// Create new variant
			variant := &model.ProductVariant{
				ProductID:        id,
				SKU:              v.SKU,
				Name:             v.Name,
				Attributes:       model.JSONMap(v.Attributes),
				Price:            v.Price,
				StandardCost:     v.StandardCost,
				LastPurchaseCost: v.LastPurchaseCost,
				IsDefault:        isDefault,
				IsActive:         variantIsActive,
			}
			if err := s.variantRepo.Create(ctx, variant); err != nil {
				return nil, apperror.InternalError(err)
			}
		}
	}

	// Delete variants that were not in the update request
	for _, ev := range existingVariants {
		if !updatedIDs[ev.ID] {
			if err := s.variantRepo.Delete(ctx, ev.ID); err != nil {
				return nil, apperror.InternalError(err)
			}
		}
	}

	return s.GetByID(ctx, companyID, id)
}

func (s *ProductService) Delete(ctx context.Context, companyID, id int64) error {
	// Check if product exists
	product, err := s.productRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if product == nil {
		return apperror.NotFound("Product not found")
	}

	// Delete variants first (cascade should handle this, but be explicit)
	if err := s.variantRepo.DeleteByProductID(ctx, id); err != nil {
		return apperror.InternalError(err)
	}

	// Delete product
	if err := s.productRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

func (s *ProductService) toResponse(ctx context.Context, companyID int64, p *model.Product, variants []*model.ProductVariant) *dto.ProductResponse {
	resp := &dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		CategoryID:  p.ProductCategoryID,
		Description: p.Description,
		ImageData:   p.ImageData,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// Get category name if category exists
	if p.ProductCategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, companyID, *p.ProductCategoryID)
		if err == nil && category != nil {
			resp.CategoryName = &category.Name
		}
	}

	// Map variants
	resp.Variants = make([]*dto.ProductVariantResponse, len(variants))
	for i, v := range variants {
		resp.Variants[i] = &dto.ProductVariantResponse{
			ID:               v.ID,
			SKU:              v.SKU,
			Name:             v.Name,
			Attributes:       v.Attributes,
			Price:            v.Price,
			StandardCost:     v.StandardCost,
			LastPurchaseCost: v.LastPurchaseCost,
			IsDefault:        v.IsDefault,
			IsActive:         v.IsActive,
			CreatedAt:        v.CreatedAt,
			UpdatedAt:        v.UpdatedAt,
		}
	}

	return resp
}
