package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockProductCategoryRepository struct {
	mock.Mock
}

func NewMockProductCategoryRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProductCategoryRepository {
	m := &MockProductCategoryRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockProductCategoryRepository) EXPECT() *MockProductCategoryRepositoryExpectation {
	return &MockProductCategoryRepositoryExpectation{mock: m}
}

func (m *MockProductCategoryRepository) Create(ctx context.Context, category *model.ProductCategory) error {
	ret := m.Called(ctx, category)
	return ret.Error(0)
}

func (m *MockProductCategoryRepository) GetByID(ctx context.Context, companyID, id int64) (*model.ProductCategory, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.ProductCategory
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.ProductCategory)
	}
	return r0, ret.Error(1)
}

func (m *MockProductCategoryRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.ProductCategory, error) {
	ret := m.Called(ctx, companyID, code)
	var r0 *model.ProductCategory
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.ProductCategory)
	}
	return r0, ret.Error(1)
}

func (m *MockProductCategoryRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.ProductCategory, int, error) {
	ret := m.Called(ctx, companyID, search, isActive, limit, offset)
	var r0 []*model.ProductCategory
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.ProductCategory)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockProductCategoryRepository) ListAll(ctx context.Context, companyID int64) ([]*model.ProductCategory, error) {
	ret := m.Called(ctx, companyID)
	var r0 []*model.ProductCategory
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.ProductCategory)
	}
	return r0, ret.Error(1)
}

func (m *MockProductCategoryRepository) Update(ctx context.Context, category *model.ProductCategory) error {
	ret := m.Called(ctx, category)
	return ret.Error(0)
}

func (m *MockProductCategoryRepository) Delete(ctx context.Context, companyID, id int64) error {
	ret := m.Called(ctx, companyID, id)
	return ret.Error(0)
}

func (m *MockProductCategoryRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	ret := m.Called(ctx, companyID, code, excludeID)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockProductCategoryRepository) HasChildren(ctx context.Context, companyID, id int64) (bool, error) {
	ret := m.Called(ctx, companyID, id)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockProductCategoryRepository) HasProducts(ctx context.Context, companyID, id int64) (bool, error) {
	ret := m.Called(ctx, companyID, id)
	return ret.Bool(0), ret.Error(1)
}

type MockProductCategoryRepositoryExpectation struct {
	mock *MockProductCategoryRepository
}

func (e *MockProductCategoryRepositoryExpectation) Create(ctx context.Context, category interface{}) *mock.Call {
	return e.mock.On("Create", ctx, category)
}

func (e *MockProductCategoryRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockProductCategoryRepositoryExpectation) GetByCode(ctx context.Context, companyID int64, code string) *mock.Call {
	return e.mock.On("GetByCode", ctx, companyID, code)
}

func (e *MockProductCategoryRepositoryExpectation) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, search, isActive, limit, offset)
}

func (e *MockProductCategoryRepositoryExpectation) ListAll(ctx context.Context, companyID int64) *mock.Call {
	return e.mock.On("ListAll", ctx, companyID)
}

func (e *MockProductCategoryRepositoryExpectation) Update(ctx context.Context, category interface{}) *mock.Call {
	return e.mock.On("Update", ctx, category)
}

func (e *MockProductCategoryRepositoryExpectation) Delete(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, companyID, id)
}

func (e *MockProductCategoryRepositoryExpectation) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) *mock.Call {
	return e.mock.On("CodeExists", ctx, companyID, code, excludeID)
}

func (e *MockProductCategoryRepositoryExpectation) HasChildren(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("HasChildren", ctx, companyID, id)
}

func (e *MockProductCategoryRepositoryExpectation) HasProducts(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("HasProducts", ctx, companyID, id)
}
