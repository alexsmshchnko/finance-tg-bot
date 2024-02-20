// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/repository/reports.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	entity "finance-tg-bot/internal/entity"
	slog "log/slog"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockReporter is a mock of Reporter interface.
type MockReporter struct {
	ctrl     *gomock.Controller
	recorder *MockReporterMockRecorder
}

// MockReporterMockRecorder is the mock recorder for MockReporter.
type MockReporterMockRecorder struct {
	mock *MockReporter
}

// NewMockReporter creates a new mock instance.
func NewMockReporter(ctrl *gomock.Controller) *MockReporter {
	mock := &MockReporter{ctrl: ctrl}
	mock.recorder = &MockReporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReporter) EXPECT() *MockReporterMockRecorder {
	return m.recorder
}

// GetStatementTotals mocks base method.
func (m *MockReporter) GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) ([]entity.ReportResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatementTotals", ctx, log, p)
	ret0, _ := ret[0].([]entity.ReportResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatementTotals indicates an expected call of GetStatementTotals.
func (mr *MockReporterMockRecorder) GetStatementTotals(ctx, log, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatementTotals", reflect.TypeOf((*MockReporter)(nil).GetStatementTotals), ctx, log, p)
}
