package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockCompanyRepository struct {
	mock.Mock
}

func NewMockCompanyRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCompanyRepository {
	m := &MockCompanyRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockCompanyRepository) EXPECT() *MockCompanyRepositoryExpectation {
	return &MockCompanyRepositoryExpectation{mock: m}
}

func (m *MockCompanyRepository) Create(ctx context.Context, company *model.Company) error {
	ret := m.Called(ctx, company)
	return ret.Error(0)
}

func (m *MockCompanyRepository) GetByID(ctx context.Context, id int64) (*model.Company, error) {
	ret := m.Called(ctx, id)
	var r0 *model.Company
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Company)
	}
	return r0, ret.Error(1)
}

func (m *MockCompanyRepository) GetByCode(ctx context.Context, code string) (*model.Company, error) {
	ret := m.Called(ctx, code)
	var r0 *model.Company
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Company)
	}
	return r0, ret.Error(1)
}

func (m *MockCompanyRepository) Update(ctx context.Context, company *model.Company) error {
	ret := m.Called(ctx, company)
	return ret.Error(0)
}

type MockCompanyRepositoryExpectation struct {
	mock *MockCompanyRepository
}

func (e *MockCompanyRepositoryExpectation) Create(ctx context.Context, company interface{}) *mock.Call {
	return e.mock.On("Create", ctx, company)
}

func (e *MockCompanyRepositoryExpectation) GetByID(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, id)
}

func (e *MockCompanyRepositoryExpectation) GetByCode(ctx context.Context, code string) *mock.Call {
	return e.mock.On("GetByCode", ctx, code)
}

func (e *MockCompanyRepositoryExpectation) Update(ctx context.Context, company interface{}) *mock.Call {
	return e.mock.On("Update", ctx, company)
}
