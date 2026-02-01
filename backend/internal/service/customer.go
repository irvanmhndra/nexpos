package service

import (
	"context"

	"github.com/irvanmhndra/pos-core-api/internal/dto"
	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/irvanmhndra/pos-core-api/internal/repository"
	"github.com/irvanmhndra/pos-core-api/pkg/apperror"
)

type CustomerService struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

func (s *CustomerService) Create(ctx context.Context, companyID int64, req dto.CreateCustomerRequest) (*dto.CustomerResponse, error) {
	// Check if code already exists
	exists, err := s.customerRepo.CodeExists(ctx, companyID, req.Code, 0)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Customer code already exists")
	}

	customer := &model.Customer{
		CompanyID: companyID,
		Code:      req.Code,
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		IsMember:  req.IsMember,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(customer), nil
}

func (s *CustomerService) GetByID(ctx context.Context, companyID, id int64) (*dto.CustomerResponse, error) {
	customer, err := s.customerRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if customer == nil {
		return nil, apperror.NotFound("Customer not found")
	}

	return s.toResponse(customer), nil
}

func (s *CustomerService) List(ctx context.Context, companyID int64, req dto.ListCustomerRequest) (*dto.CustomerListResponse, error) {
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

	customers, total, err := s.customerRepo.List(ctx, companyID, req.Search, req.IsMember, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Build response
	responses := make([]*dto.CustomerResponse, len(customers))
	for i, c := range customers {
		responses[i] = s.toResponse(c)
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

	return &dto.CustomerListResponse{
		Customers: responses,
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

func (s *CustomerService) Update(ctx context.Context, companyID, id int64, req dto.UpdateCustomerRequest) (*dto.CustomerResponse, error) {
	// Check if customer exists
	customer, err := s.customerRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if customer == nil {
		return nil, apperror.NotFound("Customer not found")
	}

	// Check if code already exists (excluding current customer)
	exists, err := s.customerRepo.CodeExists(ctx, companyID, req.Code, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if exists {
		return nil, apperror.BadRequest("Customer code already exists")
	}

	// Update fields
	customer.Code = req.Code
	customer.Name = req.Name
	customer.Phone = req.Phone
	customer.Email = req.Email
	customer.IsMember = req.IsMember

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Fetch updated customer
	customer, _ = s.customerRepo.GetByID(ctx, companyID, id)
	return s.toResponse(customer), nil
}

func (s *CustomerService) Delete(ctx context.Context, companyID, id int64) error {
	// Check if customer exists
	customer, err := s.customerRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if customer == nil {
		return apperror.NotFound("Customer not found")
	}

	if err := s.customerRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

func (s *CustomerService) toResponse(c *model.Customer) *dto.CustomerResponse {
	return &dto.CustomerResponse{
		ID:        c.ID,
		Code:      c.Code,
		Name:      c.Name,
		Phone:     c.Phone,
		Email:     c.Email,
		IsMember:  c.IsMember,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
