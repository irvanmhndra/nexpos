package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockCustomerRepository struct {
	mock.Mock
}

func NewMockCustomerRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCustomerRepository {
	m := &MockCustomerRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockCustomerRepository) EXPECT() *MockCustomerRepositoryExpectation {
	return &MockCustomerRepositoryExpectation{mock: m}
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	ret := m.Called(ctx, customer)
	return ret.Error(0)
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Customer, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.Customer
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Customer)
	}
	return r0, ret.Error(1)
}

func (m *MockCustomerRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.Customer, error) {
	ret := m.Called(ctx, companyID, code)
	var r0 *model.Customer
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Customer)
	}
	return r0, ret.Error(1)
}

func (m *MockCustomerRepository) List(ctx context.Context, companyID int64, search string, isMember *bool, limit, offset int) ([]*model.Customer, int, error) {
	ret := m.Called(ctx, companyID, search, isMember, limit, offset)
	var r0 []*model.Customer
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Customer)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *model.Customer) error {
	ret := m.Called(ctx, customer)
	return ret.Error(0)
}

func (m *MockCustomerRepository) Delete(ctx context.Context, companyID, id int64) error {
	ret := m.Called(ctx, companyID, id)
	return ret.Error(0)
}

func (m *MockCustomerRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	ret := m.Called(ctx, companyID, code, excludeID)
	return ret.Bool(0), ret.Error(1)
}

type MockCustomerRepositoryExpectation struct {
	mock *MockCustomerRepository
}

func (e *MockCustomerRepositoryExpectation) Create(ctx context.Context, customer interface{}) *mock.Call {
	return e.mock.On("Create", ctx, customer)
}

func (e *MockCustomerRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockCustomerRepositoryExpectation) GetByCode(ctx context.Context, companyID int64, code string) *mock.Call {
	return e.mock.On("GetByCode", ctx, companyID, code)
}

func (e *MockCustomerRepositoryExpectation) List(ctx context.Context, companyID int64, search string, isMember *bool, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, search, isMember, limit, offset)
}

func (e *MockCustomerRepositoryExpectation) Update(ctx context.Context, customer interface{}) *mock.Call {
	return e.mock.On("Update", ctx, customer)
}

func (e *MockCustomerRepositoryExpectation) Delete(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, companyID, id)
}

func (e *MockCustomerRepositoryExpectation) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) *mock.Call {
	return e.mock.On("CodeExists", ctx, companyID, code, excludeID)
}
