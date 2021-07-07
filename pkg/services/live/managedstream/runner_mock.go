// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/grafana/grafana/pkg/services/live/managedstream (interfaces: RuleCacheGetter)

// Package managedstream is a generated GoMock package.
package managedstream

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/grafana/grafana/pkg/models"
)

// MockRuleCacheGetter is a mock of RuleCacheGetter interface.
type MockRuleCacheGetter struct {
	ctrl     *gomock.Controller
	recorder *MockRuleCacheGetterMockRecorder
}

// MockRuleCacheGetterMockRecorder is the mock recorder for MockRuleCacheGetter.
type MockRuleCacheGetterMockRecorder struct {
	mock *MockRuleCacheGetter
}

// NewMockRuleCacheGetter creates a new mock instance.
func NewMockRuleCacheGetter(ctrl *gomock.Controller) *MockRuleCacheGetter {
	mock := &MockRuleCacheGetter{ctrl: ctrl}
	mock.recorder = &MockRuleCacheGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRuleCacheGetter) EXPECT() *MockRuleCacheGetterMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockRuleCacheGetter) Get(arg0 int64, arg1 string) (*models.LiveChannelRule, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*models.LiveChannelRule)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Get indicates an expected call of Get.
func (mr *MockRuleCacheGetterMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRuleCacheGetter)(nil).Get), arg0, arg1)
}
