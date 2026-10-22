package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockStockOpnameRepository struct {
	mock.Mock
}

func NewMockStockOpnameRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStockOpnameRepository {
	m := &MockStockOpnameRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockStockOpnameRepository) EXPECT() *MockStockOpnameRepositoryExpectation {
	return &MockStockOpnameRepositoryExpectation{mock: m}
}

func (m *MockStockOpnameRepository) GetByIDForUpdate(ctx context.Context, companyID, id int64) (*model.StockOpname, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.StockOpname
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.StockOpname)
	}
	return r0, ret.Error(1)
}

func (m *MockStockOpnameRepository) Create(ctx context.Context, opname *model.StockOpname) error {
	return m.Called(ctx, opname).Error(0)
}

func (m *MockStockOpnameRepository) GetByID(ctx context.Context, companyID, id int64) (*model.StockOpname, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.StockOpname
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.StockOpname)
	}
	return r0, ret.Error(1)
}

func (m *MockStockOpnameRepository) List(ctx context.Context, companyID int64, params *repository.StockOpnameListParams) ([]*model.StockOpname, int, error) {
	ret := m.Called(ctx, companyID, params)
	var r0 []*model.StockOpname
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.StockOpname)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockStockOpnameRepository) Update(ctx context.Context, opname *model.StockOpname) error {
	return m.Called(ctx, opname).Error(0)
}

func (m *MockStockOpnameRepository) GenerateOpnameNumber(ctx context.Context, companyID int64) (string, error) {
	ret := m.Called(ctx, companyID)
	return ret.String(0), ret.Error(1)
}

func (m *MockStockOpnameRepository) SnapshotItems(ctx context.Context, opnameID, companyID, branchID int64, categoryID *int64) (int, error) {
	ret := m.Called(ctx, opnameID, companyID, branchID, categoryID)
	return ret.Int(0), ret.Error(1)
}

func (m *MockStockOpnameRepository) GetItems(ctx context.Context, opnameID int64) ([]*model.StockOpnameItem, error) {
	ret := m.Called(ctx, opnameID)
	var r0 []*model.StockOpnameItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.StockOpnameItem)
	}
	return r0, ret.Error(1)
}

func (m *MockStockOpnameRepository) GetItem(ctx context.Context, opnameID, itemID int64) (*model.StockOpnameItem, error) {
	ret := m.Called(ctx, opnameID, itemID)
	var r0 *model.StockOpnameItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.StockOpnameItem)
	}
	return r0, ret.Error(1)
}

func (m *MockStockOpnameRepository) UpdateItem(ctx context.Context, item *model.StockOpnameItem) error {
	return m.Called(ctx, item).Error(0)
}

func (m *MockStockOpnameRepository) GetItemStats(ctx context.Context, opnameID int64) (int, int, error) {
	ret := m.Called(ctx, opnameID)
	return ret.Int(0), ret.Int(1), ret.Error(2)
}

type MockStockOpnameRepositoryExpectation struct {
	mock *MockStockOpnameRepository
}

func (e *MockStockOpnameRepositoryExpectation) Create(ctx context.Context, opname interface{}) *mock.Call {
	return e.mock.On("Create", ctx, opname)
}

func (e *MockStockOpnameRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockStockOpnameRepositoryExpectation) List(ctx context.Context, companyID int64, params interface{}) *mock.Call {
	return e.mock.On("List", ctx, companyID, params)
}

func (e *MockStockOpnameRepositoryExpectation) Update(ctx context.Context, opname interface{}) *mock.Call {
	return e.mock.On("Update", ctx, opname)
}

func (e *MockStockOpnameRepositoryExpectation) GenerateOpnameNumber(ctx context.Context, companyID int64) *mock.Call {
	return e.mock.On("GenerateOpnameNumber", ctx, companyID)
}

func (e *MockStockOpnameRepositoryExpectation) SnapshotItems(ctx context.Context, opnameID, companyID, branchID int64, categoryID interface{}) *mock.Call {
	return e.mock.On("SnapshotItems", ctx, opnameID, companyID, branchID, categoryID)
}

func (e *MockStockOpnameRepositoryExpectation) GetItems(ctx context.Context, opnameID int64) *mock.Call {
	return e.mock.On("GetItems", ctx, opnameID)
}

func (e *MockStockOpnameRepositoryExpectation) GetItem(ctx context.Context, opnameID, itemID int64) *mock.Call {
	return e.mock.On("GetItem", ctx, opnameID, itemID)
}

func (e *MockStockOpnameRepositoryExpectation) UpdateItem(ctx context.Context, item interface{}) *mock.Call {
	return e.mock.On("UpdateItem", ctx, item)
}

func (e *MockStockOpnameRepositoryExpectation) GetItemStats(ctx context.Context, opnameID int64) *mock.Call {
	return e.mock.On("GetItemStats", ctx, opnameID)
}

func (e *MockStockOpnameRepositoryExpectation) GetByIDForUpdate(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByIDForUpdate", ctx, companyID, id)
}
