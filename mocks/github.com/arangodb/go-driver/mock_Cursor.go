// Code generated by mockery v2.43.2. DO NOT EDIT.

package driver_mocks

import (
	context "context"

	driver "github.com/arangodb/go-driver"
	mock "github.com/stretchr/testify/mock"
)

// MockCursor is an autogenerated mock type for the Cursor type
type MockCursor struct {
	mock.Mock
}

type MockCursor_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCursor) EXPECT() *MockCursor_Expecter {
	return &MockCursor_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *MockCursor) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCursor_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockCursor_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockCursor_Expecter) Close() *MockCursor_Close_Call {
	return &MockCursor_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockCursor_Close_Call) Run(run func()) *MockCursor_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCursor_Close_Call) Return(_a0 error) *MockCursor_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCursor_Close_Call) RunAndReturn(run func() error) *MockCursor_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Count provides a mock function with given fields:
func (_m *MockCursor) Count() int64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Count")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// MockCursor_Count_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Count'
type MockCursor_Count_Call struct {
	*mock.Call
}

// Count is a helper method to define mock.On call
func (_e *MockCursor_Expecter) Count() *MockCursor_Count_Call {
	return &MockCursor_Count_Call{Call: _e.mock.On("Count")}
}

func (_c *MockCursor_Count_Call) Run(run func()) *MockCursor_Count_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCursor_Count_Call) Return(_a0 int64) *MockCursor_Count_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCursor_Count_Call) RunAndReturn(run func() int64) *MockCursor_Count_Call {
	_c.Call.Return(run)
	return _c
}

// Extra provides a mock function with given fields:
func (_m *MockCursor) Extra() driver.QueryExtra {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Extra")
	}

	var r0 driver.QueryExtra
	if rf, ok := ret.Get(0).(func() driver.QueryExtra); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(driver.QueryExtra)
		}
	}

	return r0
}

// MockCursor_Extra_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Extra'
type MockCursor_Extra_Call struct {
	*mock.Call
}

// Extra is a helper method to define mock.On call
func (_e *MockCursor_Expecter) Extra() *MockCursor_Extra_Call {
	return &MockCursor_Extra_Call{Call: _e.mock.On("Extra")}
}

func (_c *MockCursor_Extra_Call) Run(run func()) *MockCursor_Extra_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCursor_Extra_Call) Return(_a0 driver.QueryExtra) *MockCursor_Extra_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCursor_Extra_Call) RunAndReturn(run func() driver.QueryExtra) *MockCursor_Extra_Call {
	_c.Call.Return(run)
	return _c
}

// HasMore provides a mock function with given fields:
func (_m *MockCursor) HasMore() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HasMore")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockCursor_HasMore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasMore'
type MockCursor_HasMore_Call struct {
	*mock.Call
}

// HasMore is a helper method to define mock.On call
func (_e *MockCursor_Expecter) HasMore() *MockCursor_HasMore_Call {
	return &MockCursor_HasMore_Call{Call: _e.mock.On("HasMore")}
}

func (_c *MockCursor_HasMore_Call) Run(run func()) *MockCursor_HasMore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCursor_HasMore_Call) Return(_a0 bool) *MockCursor_HasMore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCursor_HasMore_Call) RunAndReturn(run func() bool) *MockCursor_HasMore_Call {
	_c.Call.Return(run)
	return _c
}

// ReadDocument provides a mock function with given fields: ctx, result
func (_m *MockCursor) ReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error) {
	ret := _m.Called(ctx, result)

	if len(ret) == 0 {
		panic("no return value specified for ReadDocument")
	}

	var r0 driver.DocumentMeta
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) (driver.DocumentMeta, error)); ok {
		return rf(ctx, result)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) driver.DocumentMeta); ok {
		r0 = rf(ctx, result)
	} else {
		r0 = ret.Get(0).(driver.DocumentMeta)
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}) error); ok {
		r1 = rf(ctx, result)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCursor_ReadDocument_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadDocument'
type MockCursor_ReadDocument_Call struct {
	*mock.Call
}

// ReadDocument is a helper method to define mock.On call
//   - ctx context.Context
//   - result interface{}
func (_e *MockCursor_Expecter) ReadDocument(ctx interface{}, result interface{}) *MockCursor_ReadDocument_Call {
	return &MockCursor_ReadDocument_Call{Call: _e.mock.On("ReadDocument", ctx, result)}
}

func (_c *MockCursor_ReadDocument_Call) Run(run func(ctx context.Context, result interface{})) *MockCursor_ReadDocument_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *MockCursor_ReadDocument_Call) Return(_a0 driver.DocumentMeta, _a1 error) *MockCursor_ReadDocument_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCursor_ReadDocument_Call) RunAndReturn(run func(context.Context, interface{}) (driver.DocumentMeta, error)) *MockCursor_ReadDocument_Call {
	_c.Call.Return(run)
	return _c
}

// RetryReadDocument provides a mock function with given fields: ctx, result
func (_m *MockCursor) RetryReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error) {
	ret := _m.Called(ctx, result)

	if len(ret) == 0 {
		panic("no return value specified for RetryReadDocument")
	}

	var r0 driver.DocumentMeta
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) (driver.DocumentMeta, error)); ok {
		return rf(ctx, result)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) driver.DocumentMeta); ok {
		r0 = rf(ctx, result)
	} else {
		r0 = ret.Get(0).(driver.DocumentMeta)
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}) error); ok {
		r1 = rf(ctx, result)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCursor_RetryReadDocument_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetryReadDocument'
type MockCursor_RetryReadDocument_Call struct {
	*mock.Call
}

// RetryReadDocument is a helper method to define mock.On call
//   - ctx context.Context
//   - result interface{}
func (_e *MockCursor_Expecter) RetryReadDocument(ctx interface{}, result interface{}) *MockCursor_RetryReadDocument_Call {
	return &MockCursor_RetryReadDocument_Call{Call: _e.mock.On("RetryReadDocument", ctx, result)}
}

func (_c *MockCursor_RetryReadDocument_Call) Run(run func(ctx context.Context, result interface{})) *MockCursor_RetryReadDocument_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *MockCursor_RetryReadDocument_Call) Return(_a0 driver.DocumentMeta, _a1 error) *MockCursor_RetryReadDocument_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCursor_RetryReadDocument_Call) RunAndReturn(run func(context.Context, interface{}) (driver.DocumentMeta, error)) *MockCursor_RetryReadDocument_Call {
	_c.Call.Return(run)
	return _c
}

// Statistics provides a mock function with given fields:
func (_m *MockCursor) Statistics() driver.QueryStatistics {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Statistics")
	}

	var r0 driver.QueryStatistics
	if rf, ok := ret.Get(0).(func() driver.QueryStatistics); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(driver.QueryStatistics)
		}
	}

	return r0
}

// MockCursor_Statistics_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Statistics'
type MockCursor_Statistics_Call struct {
	*mock.Call
}

// Statistics is a helper method to define mock.On call
func (_e *MockCursor_Expecter) Statistics() *MockCursor_Statistics_Call {
	return &MockCursor_Statistics_Call{Call: _e.mock.On("Statistics")}
}

func (_c *MockCursor_Statistics_Call) Run(run func()) *MockCursor_Statistics_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCursor_Statistics_Call) Return(_a0 driver.QueryStatistics) *MockCursor_Statistics_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCursor_Statistics_Call) RunAndReturn(run func() driver.QueryStatistics) *MockCursor_Statistics_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCursor creates a new instance of MockCursor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCursor(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCursor {
	mock := &MockCursor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
