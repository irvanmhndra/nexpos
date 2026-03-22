package mocks

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockPromotionRepository struct {
	mock.Mock
}

func NewMockPromotionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPromotionRepository {
	m := &MockPromotionRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockPromotionRepository) EXPECT() *MockPromotionRepositoryExpectation {
	return &MockPromotionRepositoryExpectation{mock: m}
}

func (m *MockPromotionRepository) Create(ctx context.Context, promotion *model.Promotion) error {
	ret := m.Called(ctx, promotion)
	return ret.Error(0)
}

func (m *MockPromotionRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Promotion, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.Promotion
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Promotion)
	}
	return r0, ret.Error(1)
}

func (m *MockPromotionRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, promoType string, limit, offset int) ([]*model.Promotion, int, error) {
	ret := m.Called(ctx, companyID, search, isActive, promoType, limit, offset)
	var r0 []*model.Promotion
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Promotion)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockPromotionRepository) Update(ctx context.Context, promotion *model.Promotion) error {
	ret := m.Called(ctx, promotion)
	return ret.Error(0)
}

func (m *MockPromotionRepository) Delete(ctx context.Context, companyID, id int64) error {
	ret := m.Called(ctx, companyID, id)
	return ret.Error(0)
}

func (m *MockPromotionRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	ret := m.Called(ctx, companyID, code, excludeID)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockPromotionRepository) GetActivePromotions(ctx context.Context, companyID int64, now time.Time) ([]*model.Promotion, error) {
	ret := m.Called(ctx, companyID, now)
	var r0 []*model.Promotion
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Promotion)
	}
	return r0, ret.Error(1)
}

func (m *MockPromotionRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.Promotion, error) {
	ret := m.Called(ctx, companyID, code)
	var r0 *model.Promotion
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Promotion)
	}
	return r0, ret.Error(1)
}

type MockPromotionRepositoryExpectation struct {
	mock *MockPromotionRepository
}

func (e *MockPromotionRepositoryExpectation) Create(ctx context.Context, promotion interface{}) *mock.Call {
	return e.mock.On("Create", ctx, promotion)
}

func (e *MockPromotionRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockPromotionRepositoryExpectation) List(ctx context.Context, companyID int64, search string, isActive *bool, promoType string, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, search, isActive, promoType, limit, offset)
}

func (e *MockPromotionRepositoryExpectation) Update(ctx context.Context, promotion interface{}) *mock.Call {
	return e.mock.On("Update", ctx, promotion)
}

func (e *MockPromotionRepositoryExpectation) Delete(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, companyID, id)
}

func (e *MockPromotionRepositoryExpectation) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) *mock.Call {
	return e.mock.On("CodeExists", ctx, companyID, code, excludeID)
}

func (e *MockPromotionRepositoryExpectation) GetActivePromotions(ctx context.Context, companyID int64, now time.Time) *mock.Call {
	return e.mock.On("GetActivePromotions", ctx, companyID, now)
}

func (e *MockPromotionRepositoryExpectation) GetByCode(ctx context.Context, companyID int64, code string) *mock.Call {
	return e.mock.On("GetByCode", ctx, companyID, code)
}
