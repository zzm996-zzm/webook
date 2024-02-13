// Code generated by MockGen. DO NOT EDIT.
// Source: ./code.go
//
// Generated by this command:
//
//	mockgen -source=./code.go -destination=./mock/code.mock.go -package=svcmocks
//

// Package svcmocks is a generated GoMock package.
package svcmocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockCodeService is a mock of CodeService interface.
type MockCodeService struct {
	ctrl     *gomock.Controller
	recorder *MockCodeServiceMockRecorder
}

// MockCodeServiceMockRecorder is the mock recorder for MockCodeService.
type MockCodeServiceMockRecorder struct {
	mock *MockCodeService
}

// NewMockCodeService creates a new mock instance.
func NewMockCodeService(ctrl *gomock.Controller) *MockCodeService {
	mock := &MockCodeService{ctrl: ctrl}
	mock.recorder = &MockCodeServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodeService) EXPECT() *MockCodeServiceMockRecorder {
	return m.recorder
}

// Send mocks base method.
func (m *MockCodeService) Send(ctx context.Context, biz, phone string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", ctx, biz, phone)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockCodeServiceMockRecorder) Send(ctx, biz, phone any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockCodeService)(nil).Send), ctx, biz, phone)
}

// Verify mocks base method.
func (m *MockCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", ctx, biz, phone, inputCode)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockCodeServiceMockRecorder) Verify(ctx, biz, phone, inputCode any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockCodeService)(nil).Verify), ctx, biz, phone, inputCode)
}
