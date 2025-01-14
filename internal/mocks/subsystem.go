// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/base-org/pessimism/internal/subsystem (interfaces: Subsystem)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/base-org/pessimism/internal/api/models"
	core "github.com/base-org/pessimism/internal/core"
	heuristic "github.com/base-org/pessimism/internal/engine/heuristic"
	gomock "github.com/golang/mock/gomock"
)

// SubManager is a mock of Subsystem interface.
type SubManager struct {
	ctrl     *gomock.Controller
	recorder *SubManagerMockRecorder
}

// SubManagerMockRecorder is the mock recorder for SubManager.
type SubManagerMockRecorder struct {
	mock *SubManager
}

// NewSubManager creates a new mock instance.
func NewSubManager(ctrl *gomock.Controller) *SubManager {
	mock := &SubManager{ctrl: ctrl}
	mock.recorder = &SubManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *SubManager) EXPECT() *SubManagerMockRecorder {
	return m.recorder
}

// BuildDeployCfg mocks base method.
func (m *SubManager) BuildDeployCfg(arg0 *core.PipelineConfig, arg1 *core.SessionConfig) (*heuristic.DeployConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildDeployCfg", arg0, arg1)
	ret0, _ := ret[0].(*heuristic.DeployConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildDeployCfg indicates an expected call of BuildDeployCfg.
func (mr *SubManagerMockRecorder) BuildDeployCfg(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildDeployCfg", reflect.TypeOf((*SubManager)(nil).BuildDeployCfg), arg0, arg1)
}

// BuildPipelineCfg mocks base method.
func (m *SubManager) BuildPipelineCfg(arg0 *models.SessionRequestParams) (*core.PipelineConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildPipelineCfg", arg0)
	ret0, _ := ret[0].(*core.PipelineConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildPipelineCfg indicates an expected call of BuildPipelineCfg.
func (mr *SubManagerMockRecorder) BuildPipelineCfg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildPipelineCfg", reflect.TypeOf((*SubManager)(nil).BuildPipelineCfg), arg0)
}

// RunSession mocks base method.
func (m *SubManager) RunSession(arg0 *heuristic.DeployConfig) (core.SUUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunSession", arg0)
	ret0, _ := ret[0].(core.SUUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunSession indicates an expected call of RunSession.
func (mr *SubManagerMockRecorder) RunSession(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunSession", reflect.TypeOf((*SubManager)(nil).RunSession), arg0)
}

// Shutdown mocks base method.
func (m *SubManager) Shutdown() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown")
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *SubManagerMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*SubManager)(nil).Shutdown))
}

// StartEventRoutines mocks base method.
func (m *SubManager) StartEventRoutines(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartEventRoutines", arg0)
}

// StartEventRoutines indicates an expected call of StartEventRoutines.
func (mr *SubManagerMockRecorder) StartEventRoutines(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartEventRoutines", reflect.TypeOf((*SubManager)(nil).StartEventRoutines), arg0)
}
