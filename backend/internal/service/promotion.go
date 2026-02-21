package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/pos-core-api/internal/dto"
	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/irvanmhndra/pos-core-api/internal/repository"
	"github.com/irvanmhndra/pos-core-api/pkg/apperror"
)

type PromotionService struct {
	promotionRepo repository.PromotionRepository
}

func NewPromotionService(promotionRepo repository.PromotionRepository) *PromotionService {
	return &PromotionService{promotionRepo: promotionRepo}
}

func (s *PromotionService) Create(ctx context.Context, companyID int64, req dto.CreatePromotionRequest) (*dto.PromotionResponse, error) {
	exists, err := s.promotionRepo.CodeExists(ctx, companyID, req.Code, 0)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Promotion code already exists")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return nil, apperror.BadRequest("Invalid start_at format")
	}

	var endAt *time.Time
	if req.EndAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			return nil, apperror.BadRequest("Invalid end_at format")
		}
		endAt = &t
	}

	promotion := &model.Promotion{
		CompanyID:     companyID,
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		Type:          req.Type,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MinPurchase:   req.MinPurchase,
		MaxDiscount:   req.MaxDiscount,
		StartAt:       startAt,
		EndAt:         endAt,
		Priority:      req.Priority,
		IsActive:      req.IsActive,
	}

	if err := s.promotionRepo.Create(ctx, promotion); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(promotion), nil
}

func (s *PromotionService) GetByID(ctx context.Context, companyID, id int64) (*dto.PromotionResponse, error) {
	promotion, err := s.promotionRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if promotion == nil {
		return nil, apperror.NotFound("Promotion not found")
	}

	return s.toResponse(promotion), nil
}

func (s *PromotionService) List(ctx context.Context, companyID int64, req dto.ListPromotionRequest) (*dto.PromotionListResponse, error) {
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

	promotions, total, err := s.promotionRepo.List(ctx, companyID, req.Search, req.IsActive, req.Type, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.PromotionResponse, len(promotions))
	for i, p := range promotions {
		responses[i] = s.toResponse(p)
	}

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

	return &dto.PromotionListResponse{
		Promotions: responses,
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

func (s *PromotionService) Update(ctx context.Context, companyID, id int64, req dto.UpdatePromotionRequest) (*dto.PromotionResponse, error) {
	promotion, err := s.promotionRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if promotion == nil {
		return nil, apperror.NotFound("Promotion not found")
	}

	exists, err := s.promotionRepo.CodeExists(ctx, companyID, req.Code, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Promotion code already exists")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return nil, apperror.BadRequest("Invalid start_at format")
	}

	var endAt *time.Time
	if req.EndAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			return nil, apperror.BadRequest("Invalid end_at format")
		}
		endAt = &t
	}

	promotion.Code = req.Code
	promotion.Name = req.Name
	promotion.Description = req.Description
	promotion.Type = req.Type
	promotion.DiscountType = req.DiscountType
	promotion.DiscountValue = req.DiscountValue
	promotion.MinPurchase = req.MinPurchase
	promotion.MaxDiscount = req.MaxDiscount
	promotion.StartAt = startAt
	promotion.EndAt = endAt
	promotion.Priority = req.Priority
	promotion.IsActive = req.IsActive

	if err := s.promotionRepo.Update(ctx, promotion); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(promotion), nil
}

func (s *PromotionService) Delete(ctx context.Context, companyID, id int64) error {
	promotion, err := s.promotionRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if promotion == nil {
		return apperror.NotFound("Promotion not found")
	}

	if err := s.promotionRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

func (s *PromotionService) toResponse(p *model.Promotion) *dto.PromotionResponse {
	return &dto.PromotionResponse{
		ID:            p.ID,
		Code:          p.Code,
		Name:          p.Name,
		Description:   p.Description,
		Type:          p.Type,
		DiscountType:  p.DiscountType,
		DiscountValue: p.DiscountValue,
		MinPurchase:   p.MinPurchase,
		MaxDiscount:   p.MaxDiscount,
		StartAt:       p.StartAt,
		EndAt:         p.EndAt,
		Priority:      p.Priority,
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
