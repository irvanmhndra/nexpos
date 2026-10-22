package service

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type BranchService struct {
	branchRepo repository.BranchRepository
}

func NewBranchService(branchRepo repository.BranchRepository) *BranchService {
	return &BranchService{branchRepo: branchRepo}
}

func (s *BranchService) Create(ctx context.Context, companyID int64, req dto.CreateBranchRequest) (*dto.BranchResponse, error) {
	exists, err := s.branchRepo.CodeExists(ctx, companyID, req.Code, 0)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Branch code already exists")
	}

	branch := &model.Branch{
		CompanyID: companyID,
		Code:      req.Code,
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		IsActive:  req.IsActive,
	}

	if err := s.branchRepo.Create(ctx, branch); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(branch), nil
}

func (s *BranchService) GetByID(ctx context.Context, companyID, id int64) (*dto.BranchResponse, error) {
	branch, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if branch == nil || branch.CompanyID != companyID {
		return nil, apperror.NotFound("Branch not found")
	}

	return s.toResponse(branch), nil
}

func (s *BranchService) List(ctx context.Context, companyID int64, req dto.ListBranchRequest) (*dto.BranchListResponse, error) {
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

	branches, total, err := s.branchRepo.List(ctx, companyID, req.Search, req.IsActive, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.BranchResponse, len(branches))
	for i, b := range branches {
		responses[i] = s.toResponse(b)
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

	return &dto.BranchListResponse{
		Branches: responses,
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

func (s *BranchService) Update(ctx context.Context, companyID, id int64, req dto.UpdateBranchRequest) (*dto.BranchResponse, error) {
	branch, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if branch == nil || branch.CompanyID != companyID {
		return nil, apperror.NotFound("Branch not found")
	}

	exists, err := s.branchRepo.CodeExists(ctx, companyID, req.Code, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Branch code already exists")
	}

	branch.Code = req.Code
	branch.Name = req.Name
	branch.Address = req.Address
	branch.Phone = req.Phone
	branch.IsActive = req.IsActive

	if err := s.branchRepo.Update(ctx, branch); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(branch), nil
}

func (s *BranchService) Delete(ctx context.Context, companyID, id int64) error {
	branch, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if branch == nil || branch.CompanyID != companyID {
		return apperror.NotFound("Branch not found")
	}

	if err := s.branchRepo.Delete(ctx, id); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

func (s *BranchService) toResponse(b *model.Branch) *dto.BranchResponse {
	return &dto.BranchResponse{
		ID:        b.ID,
		Code:      b.Code,
		Name:      b.Name,
		Address:   b.Address,
		Phone:     b.Phone,
		IsActive:  b.IsActive,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}
