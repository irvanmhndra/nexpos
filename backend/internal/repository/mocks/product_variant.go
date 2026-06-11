package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockProductVariantRepository struct {
	mock.Mock
}

func NewMockProductVariantRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProductVariantRepository {
	m := &MockProductVariantRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockProductVariantRepository) EXPECT() *MockProductVariantRepositoryExpectation {
	return &MockProductVariantRepositoryExpectation{mock: m}
}

func (m *MockProductVariantRepository) Create(ctx context.Context, variant *model.ProductVariant) error {
	ret := m.Called(ctx, variant)
	return ret.Error(0)
}

func (m *MockProductVariantRepository) GetByID(ctx context.Context, id int64) (*model.ProductVariant, error) {
	ret := m.Called(ctx, id)
	var r0 *model.ProductVariant
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.ProductVariant)
	}
	return r0, ret.Error(1)
}

func (m *MockProductVariantRepository) GetByProductID(ctx context.Context, productID int64) ([]*model.ProductVariant, error) {
	ret := m.Called(ctx, productID)
	var r0 []*model.ProductVariant
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.ProductVariant)
	}
	return r0, ret.Error(1)
}

func (m *MockProductVariantRepository) GetBySKU(ctx context.Context, sku string) (*model.ProductVariant, error) {
	ret := m.Called(ctx, sku)
	var r0 *model.ProductVariant
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.ProductVariant)
	}
	return r0, ret.Error(1)
}

func (m *MockProductVariantRepository) Update(ctx context.Context, variant *model.ProductVariant) error {
	ret := m.Called(ctx, variant)
	return ret.Error(0)
}

func (m *MockProductVariantRepository) Delete(ctx context.Context, id int64) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockProductVariantRepository) DeleteByProductID(ctx context.Context, productID int64) error {
	ret := m.Called(ctx, productID)
	return ret.Error(0)
}

func (m *MockProductVariantRepository) SKUExists(ctx context.Context, sku string, excludeID int64) (bool, error) {
	ret := m.Called(ctx, sku, excludeID)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockProductVariantRepository) SKUExistsInOtherProduct(ctx context.Context, sku string, productID int64) (bool, error) {
	ret := m.Called(ctx, sku, productID)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockProductVariantRepository) FindByCodeInCompany(ctx context.Context, companyID int64, code string) (*model.ProductVariant, error) {
	ret := m.Called(ctx, companyID, code)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*model.ProductVariant), ret.Error(1)
}

func (m *MockProductVariantRepository) BarcodeExistsInOtherProduct(ctx context.Context, companyID int64, barcode string, productID int64) (bool, error) {
	ret := m.Called(ctx, companyID, barcode, productID)
	return ret.Bool(0), ret.Error(1)
}

type MockProductVariantRepositoryExpectation struct {
	mock *MockProductVariantRepository
}

func (e *MockProductVariantRepositoryExpectation) Create(ctx context.Context, variant interface{}) *mock.Call {
	return e.mock.On("Create", ctx, variant)
}

func (e *MockProductVariantRepositoryExpectation) GetByID(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, id)
}

func (e *MockProductVariantRepositoryExpectation) GetByProductID(ctx context.Context, productID int64) *mock.Call {
	return e.mock.On("GetByProductID", ctx, productID)
}

func (e *MockProductVariantRepositoryExpectation) GetBySKU(ctx context.Context, sku string) *mock.Call {
	return e.mock.On("GetBySKU", ctx, sku)
}

func (e *MockProductVariantRepositoryExpectation) Update(ctx context.Context, variant interface{}) *mock.Call {
	return e.mock.On("Update", ctx, variant)
}

func (e *MockProductVariantRepositoryExpectation) Delete(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, id)
}

func (e *MockProductVariantRepositoryExpectation) DeleteByProductID(ctx context.Context, productID int64) *mock.Call {
	return e.mock.On("DeleteByProductID", ctx, productID)
}

func (e *MockProductVariantRepositoryExpectation) SKUExists(ctx context.Context, sku string, excludeID int64) *mock.Call {
	return e.mock.On("SKUExists", ctx, sku, excludeID)
}

func (e *MockProductVariantRepositoryExpectation) SKUExistsInOtherProduct(ctx context.Context, sku string, productID int64) *mock.Call {
	return e.mock.On("SKUExistsInOtherProduct", ctx, sku, productID)
}

func (e *MockProductVariantRepositoryExpectation) FindByCodeInCompany(ctx context.Context, companyID int64, code string) *mock.Call {
	return e.mock.On("FindByCodeInCompany", ctx, companyID, code)
}

func (e *MockProductVariantRepositoryExpectation) BarcodeExistsInOtherProduct(ctx context.Context, companyID int64, barcode string, productID int64) *mock.Call {
	return e.mock.On("BarcodeExistsInOtherProduct", ctx, companyID, barcode, productID)
}
