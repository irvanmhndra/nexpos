package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockCompanySettingsRepository struct {
	mock.Mock
}

func NewMockCompanySettingsRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCompanySettingsRepository {
	m := &MockCompanySettingsRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockCompanySettingsRepository) EXPECT() *MockCompanySettingsRepositoryExpectation {
	return &MockCompanySettingsRepositoryExpectation{mock: m}
}

func (m *MockCompanySettingsRepository) GetByCompanyID(ctx context.Context, companyID int64) (*model.CompanySettings, error) {
	ret := m.Called(ctx, companyID)
	var r0 *model.CompanySettings
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.CompanySettings)
	}
	return r0, ret.Error(1)
}

func (m *MockCompanySettingsRepository) Upsert(ctx context.Context, settings *model.CompanySettings) error {
	ret := m.Called(ctx, settings)
	return ret.Error(0)
}

type MockCompanySettingsRepositoryExpectation struct {
	mock *MockCompanySettingsRepository
}

func (e *MockCompanySettingsRepositoryExpectation) GetByCompanyID(ctx context.Context, companyID int64) *mock.Call {
	return e.mock.On("GetByCompanyID", ctx, companyID)
}

func (e *MockCompanySettingsRepositoryExpectation) Upsert(ctx context.Context, settings interface{}) *mock.Call {
	return e.mock.On("Upsert", ctx, settings)
}
