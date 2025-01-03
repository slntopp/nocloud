// Code generated by mockery v2.43.2. DO NOT EDIT.

package rabbitmq_mocks

import (
	rabbitmq "github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	mock "github.com/stretchr/testify/mock"
)

// MockConnection is an autogenerated mock type for the Connection type
type MockConnection struct {
	mock.Mock
}

type MockConnection_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConnection) EXPECT() *MockConnection_Expecter {
	return &MockConnection_Expecter{mock: &_m.Mock}
}

// Channel provides a mock function with given fields:
func (_m *MockConnection) Channel() (rabbitmq.Channel, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Channel")
	}

	var r0 rabbitmq.Channel
	var r1 error
	if rf, ok := ret.Get(0).(func() (rabbitmq.Channel, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() rabbitmq.Channel); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rabbitmq.Channel)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConnection_Channel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Channel'
type MockConnection_Channel_Call struct {
	*mock.Call
}

// Channel is a helper method to define mock.On call
func (_e *MockConnection_Expecter) Channel() *MockConnection_Channel_Call {
	return &MockConnection_Channel_Call{Call: _e.mock.On("Channel")}
}

func (_c *MockConnection_Channel_Call) Run(run func()) *MockConnection_Channel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConnection_Channel_Call) Return(_a0 rabbitmq.Channel, _a1 error) *MockConnection_Channel_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConnection_Channel_Call) RunAndReturn(run func() (rabbitmq.Channel, error)) *MockConnection_Channel_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with given fields:
func (_m *MockConnection) Close() error {
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

// MockConnection_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockConnection_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockConnection_Expecter) Close() *MockConnection_Close_Call {
	return &MockConnection_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockConnection_Close_Call) Run(run func()) *MockConnection_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConnection_Close_Call) Return(_a0 error) *MockConnection_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConnection_Close_Call) RunAndReturn(run func() error) *MockConnection_Close_Call {
	_c.Call.Return(run)
	return _c
}

// IsClosed provides a mock function with given fields:
func (_m *MockConnection) IsClosed() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsClosed")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockConnection_IsClosed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsClosed'
type MockConnection_IsClosed_Call struct {
	*mock.Call
}

// IsClosed is a helper method to define mock.On call
func (_e *MockConnection_Expecter) IsClosed() *MockConnection_IsClosed_Call {
	return &MockConnection_IsClosed_Call{Call: _e.mock.On("IsClosed")}
}

func (_c *MockConnection_IsClosed_Call) Run(run func()) *MockConnection_IsClosed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConnection_IsClosed_Call) Return(_a0 bool) *MockConnection_IsClosed_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConnection_IsClosed_Call) RunAndReturn(run func() bool) *MockConnection_IsClosed_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConnection creates a new instance of MockConnection. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConnection(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConnection {
	mock := &MockConnection{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
