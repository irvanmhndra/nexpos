package service

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type ProductCategoryService struct {
	categoryRepo repository.ProductCategoryRepository
}

func NewProductCategoryService(categoryRepo repository.ProductCategoryRepository) *ProductCategoryService {
	return &ProductCategoryService{categoryRepo: categoryRepo}
}

func (s *ProductCategoryService) Create(ctx context.Context, companyID int64, req dto.CreateProductCategoryRequest) (*dto.ProductCategoryResponse, error) {
	// Check if code already exists
	exists, err := s.categoryRepo.CodeExists(ctx, companyID, req.Code, 0)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Category code already exists")
	}

	// Validate parent if provided
	if req.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(ctx, companyID, *req.ParentID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if parent == nil {
			return nil, apperror.BadRequest("Parent category not found")
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	category := &model.ProductCategory{
		CompanyID: companyID,
		Code:      req.Code,
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		IsActive:  isActive,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(ctx, companyID, category), nil
}

func (s *ProductCategoryService) GetByID(ctx context.Context, companyID, id int64) (*dto.ProductCategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if category == nil {
		return nil, apperror.NotFound("Category not found")
	}

	return s.toResponse(ctx, companyID, category), nil
}

func (s *ProductCategoryService) List(ctx context.Context, companyID int64, req dto.ListProductCategoryRequest) (*dto.ProductCategoryListResponse, error) {
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

	categories, total, err := s.categoryRepo.List(ctx, companyID, req.Search, req.IsActive, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Build response
	responses := make([]*dto.ProductCategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = s.toResponse(ctx, companyID, c)
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

	return &dto.ProductCategoryListResponse{
		Categories: responses,
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

func (s *ProductCategoryService) ListAll(ctx context.Context, companyID int64) ([]*dto.ProductCategoryResponse, error) {
	categories, err := s.categoryRepo.ListAll(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.ProductCategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = s.toResponse(ctx, companyID, c)
	}

	return responses, nil
}

func (s *ProductCategoryService) Update(ctx context.Context, companyID, id int64, req dto.UpdateProductCategoryRequest) (*dto.ProductCategoryResponse, error) {
	// Check if category exists
	category, err := s.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if category == nil {
		return nil, apperror.NotFound("Category not found")
	}

	// Check if code already exists (excluding current category)
	exists, err := s.categoryRepo.CodeExists(ctx, companyID, req.Code, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Category code already exists")
	}

	// Validate parent if provided
	if req.ParentID != nil {
		// Cannot set self as parent
		if *req.ParentID == id {
			return nil, apperror.BadRequest("Category cannot be its own parent")
		}

		parent, err := s.categoryRepo.GetByID(ctx, companyID, *req.ParentID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if parent == nil {
			return nil, apperror.BadRequest("Parent category not found")
		}

		// Check for circular reference (parent's parent cannot be this category)
		if parent.ParentID != nil && *parent.ParentID == id {
			return nil, apperror.BadRequest("Circular reference detected")
		}
	}

	// Update fields
	category.Code = req.Code
	category.Name = req.Name
	category.ParentID = req.ParentID
	category.SortOrder = req.SortOrder
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Fetch updated category
	category, _ = s.categoryRepo.GetByID(ctx, companyID, id)
	return s.toResponse(ctx, companyID, category), nil
}

func (s *ProductCategoryService) Delete(ctx context.Context, companyID, id int64) error {
	// Check if category exists
	category, err := s.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if category == nil {
		return apperror.NotFound("Category not found")
	}

	// Check if category has children
	hasChildren, err := s.categoryRepo.HasChildren(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if hasChildren {
		return apperror.BadRequest("Cannot delete category with subcategories")
	}

	// Check if category has products
	hasProducts, err := s.categoryRepo.HasProducts(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if hasProducts {
		return apperror.BadRequest("Cannot delete category with products")
	}

	if err := s.categoryRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

func (s *ProductCategoryService) toResponse(ctx context.Context, companyID int64, c *model.ProductCategory) *dto.ProductCategoryResponse {
	resp := &dto.ProductCategoryResponse{
		ID:        c.ID,
		Code:      c.Code,
		Name:      c.Name,
		ParentID:  c.ParentID,
		SortOrder: c.SortOrder,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	// Get parent name if parent exists
	if c.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(ctx, companyID, *c.ParentID)
		if err == nil && parent != nil {
			resp.ParentName = &parent.Name
		}
	}

	return resp
}
