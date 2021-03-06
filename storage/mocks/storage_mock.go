// Code generated by MockGen. DO NOT EDIT.
// Source: storage/storage.go

// Package mock_storage is a generated GoMock package.
package mock_storage

import (
	reflect "reflect"
	api "stock_ticker/api"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// AddPrices mocks base method.
func (m *MockStorage) AddPrices(prices *api.JSONResponse) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddPrices", prices)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddPrices indicates an expected call of AddPrices.
func (mr *MockStorageMockRecorder) AddPrices(prices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPrices", reflect.TypeOf((*MockStorage)(nil).AddPrices), prices)
}

// GetPriceInfo mocks base method.
func (m *MockStorage) GetPriceInfo(days int) ([]*api.DailyPrice, float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPriceInfo", days)
	ret0, _ := ret[0].([]*api.DailyPrice)
	ret1, _ := ret[1].(float64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetPriceInfo indicates an expected call of GetPriceInfo.
func (mr *MockStorageMockRecorder) GetPriceInfo(days interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPriceInfo", reflect.TypeOf((*MockStorage)(nil).GetPriceInfo), days)
}
