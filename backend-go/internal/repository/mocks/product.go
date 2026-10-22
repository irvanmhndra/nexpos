package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockProductRepository struct {
	mock.Mock
}

func NewMockProductRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProductRepository {
	m := &MockProductRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockProductRepository) EXPECT() *MockProductRepositoryExpectation {
	return &MockProductRepositoryExpectation{mock: m}
}

func (m *MockProductRepository) Create(ctx context.Context, product *model.Product) error {
	ret := m.Called(ctx, product)
	return ret.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Product, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.Product
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Product)
	}
	return r0, ret.Error(1)
}

func (m *MockProductRepository) List(ctx context.Context, companyID int64, search string, categoryID *int64, isActive *bool, limit, offset int) ([]*model.Product, int, error) {
	ret := m.Called(ctx, companyID, search, categoryID, isActive, limit, offset)
	var r0 []*model.Product
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Product)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockProductRepository) Update(ctx context.Context, product *model.Product) error {
	ret := m.Called(ctx, product)
	return ret.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, companyID, id int64) error {
	ret := m.Called(ctx, companyID, id)
	return ret.Error(0)
}

type MockProductRepositoryExpectation struct {
	mock *MockProductRepository
}

func (e *MockProductRepositoryExpectation) Create(ctx context.Context, product interface{}) *mock.Call {
	return e.mock.On("Create", ctx, product)
}

func (e *MockProductRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockProductRepositoryExpectation) List(ctx context.Context, companyID int64, search string, categoryID *int64, isActive *bool, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, search, categoryID, isActive, limit, offset)
}

func (e *MockProductRepositoryExpectation) Update(ctx context.Context, product interface{}) *mock.Call {
	return e.mock.On("Update", ctx, product)
}

func (e *MockProductRepositoryExpectation) Delete(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, companyID, id)
}
