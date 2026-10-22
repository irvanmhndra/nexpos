package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/internal/storage"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/lib/pq"
)

type ProductService struct {
	productRepo  repository.ProductRepository
	variantRepo  repository.ProductVariantRepository
	categoryRepo repository.ProductCategoryRepository
	store        *storage.Client // may be nil (uploads disabled)
}

func NewProductService(
	productRepo repository.ProductRepository,
	variantRepo repository.ProductVariantRepository,
	categoryRepo repository.ProductCategoryRepository,
	store *storage.Client,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		variantRepo:  variantRepo,
		categoryRepo: categoryRepo,
		store:        store,
	}
}

// deleteImage best-effort removes an image object from R2 by its stored URL. It
// never fails the caller: a leftover object is far less bad than a failed
// update/delete, and it's a no-op for empty/foreign URLs or when R2 is off.
func (s *ProductService) deleteImage(ctx context.Context, url string) {
	if s.store == nil {
		return
	}
	key := s.store.KeyFromURL(url)
	if key == "" {
		return
	}
	if err := s.store.Delete(ctx, key); err != nil {
		slog.WarnContext(ctx, "failed to delete orphaned product image from R2", "key", key, "error", err)
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

	// Validate at least one variant
	if len(req.Variants) == 0 {
		return nil, apperror.BadRequest("product must have at least one variant")
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

	// Validate barcodes (per-company, optional) don't collide
	if err := s.validateBarcodes(ctx, companyID, 0, req.Variants); err != nil {
		return nil, err
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
		ImageURL:          req.ImageURL,
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
			Barcode:          normalizeBarcode(v.Barcode),
			Name:             v.Name,
			Attributes:       model.JSONMap(v.Attributes),
			Price:            v.Price,
			StandardCost:     v.StandardCost,
			LastPurchaseCost: v.LastPurchaseCost,
			IsDefault:        isDefault,
			IsActive:         variantIsActive,
			SalePrice:        v.SalePrice,
			SaleStart:        parseOptionalTime(v.SaleStart),
			SaleEnd:          parseOptionalTime(v.SaleEnd),
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

	// Validate at least one variant
	if len(req.Variants) == 0 {
		return nil, apperror.BadRequest("product must have at least one variant")
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
		exists, err := s.variantRepo.SKUExistsInOtherProduct(ctx, v.SKU, id)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if exists {
			return nil, apperror.BadRequest("SKU already exists: " + v.SKU)
		}
	}

	// Validate barcodes (per-company, optional) don't collide with other products
	if err := s.validateBarcodes(ctx, companyID, id, req.Variants); err != nil {
		return nil, err
	}

	// Capture the current image before overwriting so we can clean up R2 after.
	var oldImageURL string
	if product.ImageURL != nil {
		oldImageURL = *product.ImageURL
	}

	// Update product
	product.ProductCategoryID = req.CategoryID
	product.Name = req.Name
	product.Description = req.Description
	product.ImageData = req.ImageData
	product.ImageURL = req.ImageURL
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Image changed (replaced or cleared) → clean up the old R2 object.
	newImageURL := ""
	if product.ImageURL != nil {
		newImageURL = *product.ImageURL
	}
	if oldImageURL != "" && oldImageURL != newImageURL {
		s.deleteImage(ctx, oldImageURL)
	}

	// Separate incoming variants: those with a valid existing ID vs brand-new ones
	var toExplicitUpdate []dto.ProductVariantInput
	var toAssign []dto.ProductVariantInput
	for _, v := range req.Variants {
		if v.ID != nil && existingIDs[*v.ID] {
			toExplicitUpdate = append(toExplicitUpdate, v)
		} else {
			toAssign = append(toAssign, v)
		}
	}

	// Collect orphan existing rows (not explicitly kept by ID)
	keptIDs := make(map[int64]bool)
	for _, v := range toExplicitUpdate {
		keptIDs[*v.ID] = true
	}
	var orphans []*model.ProductVariant
	for _, ev := range existingVariants {
		if !keptIDs[ev.ID] {
			orphans = append(orphans, ev)
		}
	}

	buildVariant := func(v dto.ProductVariantInput, pos int) *model.ProductVariant {
		variantIsActive := true
		if v.IsActive != nil {
			variantIsActive = *v.IsActive
		}
		return &model.ProductVariant{
			ProductID:        id,
			SKU:              v.SKU,
			Barcode:          normalizeBarcode(v.Barcode),
			Name:             v.Name,
			Attributes:       model.JSONMap(v.Attributes),
			Price:            v.Price,
			StandardCost:     v.StandardCost,
			LastPurchaseCost: v.LastPurchaseCost,
			IsDefault:        v.IsDefault || pos == 0,
			IsActive:         variantIsActive,
			SalePrice:        v.SalePrice,
			SaleStart:        parseOptionalTime(v.SaleStart),
			SaleEnd:          parseOptionalTime(v.SaleEnd),
		}
	}

	// Process explicit updates first
	for i, v := range toExplicitUpdate {
		variant := buildVariant(v, i)
		variant.ID = *v.ID
		if err := s.variantRepo.Update(ctx, variant); err != nil {
			return nil, apperror.InternalError(err)
		}
	}

	// For brand-new variants: reuse orphan rows (UPDATE) before inserting.
	// This avoids duplicate-key errors and minimises row churn.
	offset := len(toExplicitUpdate)
	orphanIdx := 0
	for i, v := range toAssign {
		variant := buildVariant(v, offset+i)
		if orphanIdx < len(orphans) {
			variant.ID = orphans[orphanIdx].ID
			orphanIdx++
			if err := s.variantRepo.Update(ctx, variant); err != nil {
				return nil, apperror.InternalError(err)
			}
		} else {
			if err := s.variantRepo.Create(ctx, variant); err != nil {
				return nil, apperror.InternalError(err)
			}
		}
	}

	// Delete leftover orphans only when variant count decreases
	for ; orphanIdx < len(orphans); orphanIdx++ {
		orphan := orphans[orphanIdx]
		if err := s.variantRepo.Delete(ctx, orphan.ID); err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23503" {
				return nil, apperror.BadRequest(fmt.Sprintf(
					"Variant '%s' (SKU: %s) tidak dapat dihapus karena sudah memiliki riwayat transaksi",
					orphan.Name, orphan.SKU,
				))
			}
			return nil, apperror.InternalError(err)
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

	if product.ImageURL != nil {
		s.deleteImage(ctx, *product.ImageURL)
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
		ImageURL:    p.ImageURL,
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
			Barcode:          v.Barcode,
			Name:             v.Name,
			Attributes:       v.Attributes,
			Price:            v.Price,
			StandardCost:     v.StandardCost,
			LastPurchaseCost: v.LastPurchaseCost,
			IsDefault:        v.IsDefault,
			IsActive:         v.IsActive,
			SalePrice:        v.SalePrice,
			SaleStart:        v.SaleStart,
			SaleEnd:          v.SaleEnd,
			CreatedAt:        v.CreatedAt,
			UpdatedAt:        v.UpdatedAt,
		}
	}

	return resp
}

func parseOptionalTime(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

// LookupByCode resolves a scanned code to its product + matched variant within
// the company, matching barcode first and falling back to SKU.
func (s *ProductService) LookupByCode(ctx context.Context, companyID int64, code string) (*dto.ProductLookupResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, apperror.BadRequest("code is required")
	}

	variant, err := s.variantRepo.FindByCodeInCompany(ctx, companyID, code)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if variant == nil {
		return nil, apperror.NotFound("No product found for code: " + code)
	}

	product, err := s.GetByID(ctx, companyID, variant.ProductID)
	if err != nil {
		return nil, err
	}

	matchedBy := "sku"
	if variant.Barcode != nil && *variant.Barcode == code {
		matchedBy = "barcode"
	}

	return &dto.ProductLookupResponse{
		Product:          product,
		MatchedVariantID: variant.ID,
		MatchedBy:        matchedBy,
	}, nil
}

// validateBarcodes ensures each provided (non-empty) barcode is unique within
// the request and not already used by another product in the company. Pass
// productID 0 on create.
func (s *ProductService) validateBarcodes(ctx context.Context, companyID, productID int64, variants []dto.ProductVariantInput) error {
	seen := make(map[string]bool)
	for _, v := range variants {
		bc := normalizeBarcode(v.Barcode)
		if bc == nil {
			continue
		}
		if seen[*bc] {
			return apperror.BadRequest("Duplicate barcode in request: " + *bc)
		}
		seen[*bc] = true

		exists, err := s.variantRepo.BarcodeExistsInOtherProduct(ctx, companyID, *bc, productID)
		if err != nil {
			return apperror.InternalError(err)
		}
		if exists {
			return apperror.BadRequest("Barcode already registered: " + *bc)
		}
	}
	return nil
}

// normalizeBarcode trims a barcode and treats blank as absent (nil).
func normalizeBarcode(b *string) *string {
	if b == nil {
		return nil
	}
	t := strings.TrimSpace(*b)
	if t == "" {
		return nil
	}
	return &t
}
