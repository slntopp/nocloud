// Code generated by mockery v2.43.2. DO NOT EDIT.

package graph_mocks

import (
	context "context"

	access "github.com/slntopp/nocloud-proto/access"

	driver "github.com/arangodb/go-driver"

	mock "github.com/stretchr/testify/mock"
)

// MockCommonActionsController is an autogenerated mock type for the CommonActionsController type
type MockCommonActionsController struct {
	mock.Mock
}

type MockCommonActionsController_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCommonActionsController) EXPECT() *MockCommonActionsController_Expecter {
	return &MockCommonActionsController_Expecter{mock: &_m.Mock}
}

// AccessLevel provides a mock function with given fields: ctx, account, node
func (_m *MockCommonActionsController) AccessLevel(ctx context.Context, account string, node driver.DocumentID) (bool, access.Level) {
	ret := _m.Called(ctx, account, node)

	if len(ret) == 0 {
		panic("no return value specified for AccessLevel")
	}

	var r0 bool
	var r1 access.Level
	if rf, ok := ret.Get(0).(func(context.Context, string, driver.DocumentID) (bool, access.Level)); ok {
		return rf(ctx, account, node)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, driver.DocumentID) bool); ok {
		r0 = rf(ctx, account, node)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, driver.DocumentID) access.Level); ok {
		r1 = rf(ctx, account, node)
	} else {
		r1 = ret.Get(1).(access.Level)
	}

	return r0, r1
}

// MockCommonActionsController_AccessLevel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AccessLevel'
type MockCommonActionsController_AccessLevel_Call struct {
	*mock.Call
}

// AccessLevel is a helper method to define mock.On call
//   - ctx context.Context
//   - account string
//   - node driver.DocumentID
func (_e *MockCommonActionsController_Expecter) AccessLevel(ctx interface{}, account interface{}, node interface{}) *MockCommonActionsController_AccessLevel_Call {
	return &MockCommonActionsController_AccessLevel_Call{Call: _e.mock.On("AccessLevel", ctx, account, node)}
}

func (_c *MockCommonActionsController_AccessLevel_Call) Run(run func(ctx context.Context, account string, node driver.DocumentID)) *MockCommonActionsController_AccessLevel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(driver.DocumentID))
	})
	return _c
}

func (_c *MockCommonActionsController_AccessLevel_Call) Return(_a0 bool, _a1 access.Level) *MockCommonActionsController_AccessLevel_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCommonActionsController_AccessLevel_Call) RunAndReturn(run func(context.Context, string, driver.DocumentID) (bool, access.Level)) *MockCommonActionsController_AccessLevel_Call {
	_c.Call.Return(run)
	return _c
}

// HasAccess provides a mock function with given fields: ctx, account, node, level
func (_m *MockCommonActionsController) HasAccess(ctx context.Context, account string, node driver.DocumentID, level access.Level) bool {
	ret := _m.Called(ctx, account, node, level)

	if len(ret) == 0 {
		panic("no return value specified for HasAccess")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, driver.DocumentID, access.Level) bool); ok {
		r0 = rf(ctx, account, node, level)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockCommonActionsController_HasAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasAccess'
type MockCommonActionsController_HasAccess_Call struct {
	*mock.Call
}

// HasAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - account string
//   - node driver.DocumentID
//   - level access.Level
func (_e *MockCommonActionsController_Expecter) HasAccess(ctx interface{}, account interface{}, node interface{}, level interface{}) *MockCommonActionsController_HasAccess_Call {
	return &MockCommonActionsController_HasAccess_Call{Call: _e.mock.On("HasAccess", ctx, account, node, level)}
}

func (_c *MockCommonActionsController_HasAccess_Call) Run(run func(ctx context.Context, account string, node driver.DocumentID, level access.Level)) *MockCommonActionsController_HasAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(driver.DocumentID), args[3].(access.Level))
	})
	return _c
}

func (_c *MockCommonActionsController_HasAccess_Call) Return(_a0 bool) *MockCommonActionsController_HasAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCommonActionsController_HasAccess_Call) RunAndReturn(run func(context.Context, string, driver.DocumentID, access.Level) bool) *MockCommonActionsController_HasAccess_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCommonActionsController creates a new instance of MockCommonActionsController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCommonActionsController(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCommonActionsController {
	mock := &MockCommonActionsController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}