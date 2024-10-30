// Code generated by mockery v2.43.2. DO NOT EDIT.

package rabbitmq_mocks

import (
	context "context"

	amqp091 "github.com/rabbitmq/amqp091-go"

	mock "github.com/stretchr/testify/mock"
)

// MockChannel is an autogenerated mock type for the Channel type
type MockChannel struct {
	mock.Mock
}

type MockChannel_Expecter struct {
	mock *mock.Mock
}

func (_m *MockChannel) EXPECT() *MockChannel_Expecter {
	return &MockChannel_Expecter{mock: &_m.Mock}
}

// Cancel provides a mock function with given fields: consumer, noWait
func (_m *MockChannel) Cancel(consumer string, noWait bool) error {
	ret := _m.Called(consumer, noWait)

	if len(ret) == 0 {
		panic("no return value specified for Cancel")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(consumer, noWait)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChannel_Cancel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cancel'
type MockChannel_Cancel_Call struct {
	*mock.Call
}

// Cancel is a helper method to define mock.On call
//   - consumer string
//   - noWait bool
func (_e *MockChannel_Expecter) Cancel(consumer interface{}, noWait interface{}) *MockChannel_Cancel_Call {
	return &MockChannel_Cancel_Call{Call: _e.mock.On("Cancel", consumer, noWait)}
}

func (_c *MockChannel_Cancel_Call) Run(run func(consumer string, noWait bool)) *MockChannel_Cancel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(bool))
	})
	return _c
}

func (_c *MockChannel_Cancel_Call) Return(_a0 error) *MockChannel_Cancel_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_Cancel_Call) RunAndReturn(run func(string, bool) error) *MockChannel_Cancel_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with given fields:
func (_m *MockChannel) Close() error {
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

// MockChannel_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockChannel_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockChannel_Expecter) Close() *MockChannel_Close_Call {
	return &MockChannel_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockChannel_Close_Call) Run(run func()) *MockChannel_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockChannel_Close_Call) Return(_a0 error) *MockChannel_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_Close_Call) RunAndReturn(run func() error) *MockChannel_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Consume provides a mock function with given fields: queue, consumer, autoAck, exclusive, noLocal, noWait, args
func (_m *MockChannel) Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	ret := _m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)

	if len(ret) == 0 {
		panic("no return value specified for Consume")
	}

	var r0 <-chan amqp091.Delivery
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, bool, bool, bool, bool, amqp091.Table) (<-chan amqp091.Delivery, error)); ok {
		return rf(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	}
	if rf, ok := ret.Get(0).(func(string, string, bool, bool, bool, bool, amqp091.Table) <-chan amqp091.Delivery); ok {
		r0 = rf(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan amqp091.Delivery)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, bool, bool, bool, bool, amqp091.Table) error); ok {
		r1 = rf(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockChannel_Consume_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Consume'
type MockChannel_Consume_Call struct {
	*mock.Call
}

// Consume is a helper method to define mock.On call
//   - queue string
//   - consumer string
//   - autoAck bool
//   - exclusive bool
//   - noLocal bool
//   - noWait bool
//   - args amqp091.Table
func (_e *MockChannel_Expecter) Consume(queue interface{}, consumer interface{}, autoAck interface{}, exclusive interface{}, noLocal interface{}, noWait interface{}, args interface{}) *MockChannel_Consume_Call {
	return &MockChannel_Consume_Call{Call: _e.mock.On("Consume", queue, consumer, autoAck, exclusive, noLocal, noWait, args)}
}

func (_c *MockChannel_Consume_Call) Run(run func(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table)) *MockChannel_Consume_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(bool), args[3].(bool), args[4].(bool), args[5].(bool), args[6].(amqp091.Table))
	})
	return _c
}

func (_c *MockChannel_Consume_Call) Return(_a0 <-chan amqp091.Delivery, _a1 error) *MockChannel_Consume_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockChannel_Consume_Call) RunAndReturn(run func(string, string, bool, bool, bool, bool, amqp091.Table) (<-chan amqp091.Delivery, error)) *MockChannel_Consume_Call {
	_c.Call.Return(run)
	return _c
}

// ExchangeDeclare provides a mock function with given fields: name, kind, durable, autoDelete, internal, noWait, args
func (_m *MockChannel) ExchangeDeclare(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp091.Table) error {
	ret := _m.Called(name, kind, durable, autoDelete, internal, noWait, args)

	if len(ret) == 0 {
		panic("no return value specified for ExchangeDeclare")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, bool, bool, bool, bool, amqp091.Table) error); ok {
		r0 = rf(name, kind, durable, autoDelete, internal, noWait, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChannel_ExchangeDeclare_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExchangeDeclare'
type MockChannel_ExchangeDeclare_Call struct {
	*mock.Call
}

// ExchangeDeclare is a helper method to define mock.On call
//   - name string
//   - kind string
//   - durable bool
//   - autoDelete bool
//   - internal bool
//   - noWait bool
//   - args amqp091.Table
func (_e *MockChannel_Expecter) ExchangeDeclare(name interface{}, kind interface{}, durable interface{}, autoDelete interface{}, internal interface{}, noWait interface{}, args interface{}) *MockChannel_ExchangeDeclare_Call {
	return &MockChannel_ExchangeDeclare_Call{Call: _e.mock.On("ExchangeDeclare", name, kind, durable, autoDelete, internal, noWait, args)}
}

func (_c *MockChannel_ExchangeDeclare_Call) Run(run func(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp091.Table)) *MockChannel_ExchangeDeclare_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(bool), args[3].(bool), args[4].(bool), args[5].(bool), args[6].(amqp091.Table))
	})
	return _c
}

func (_c *MockChannel_ExchangeDeclare_Call) Return(_a0 error) *MockChannel_ExchangeDeclare_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_ExchangeDeclare_Call) RunAndReturn(run func(string, string, bool, bool, bool, bool, amqp091.Table) error) *MockChannel_ExchangeDeclare_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: queue, autoAck
func (_m *MockChannel) Get(queue string, autoAck bool) (amqp091.Delivery, bool, error) {
	ret := _m.Called(queue, autoAck)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 amqp091.Delivery
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(string, bool) (amqp091.Delivery, bool, error)); ok {
		return rf(queue, autoAck)
	}
	if rf, ok := ret.Get(0).(func(string, bool) amqp091.Delivery); ok {
		r0 = rf(queue, autoAck)
	} else {
		r0 = ret.Get(0).(amqp091.Delivery)
	}

	if rf, ok := ret.Get(1).(func(string, bool) bool); ok {
		r1 = rf(queue, autoAck)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(string, bool) error); ok {
		r2 = rf(queue, autoAck)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockChannel_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockChannel_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - queue string
//   - autoAck bool
func (_e *MockChannel_Expecter) Get(queue interface{}, autoAck interface{}) *MockChannel_Get_Call {
	return &MockChannel_Get_Call{Call: _e.mock.On("Get", queue, autoAck)}
}

func (_c *MockChannel_Get_Call) Run(run func(queue string, autoAck bool)) *MockChannel_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(bool))
	})
	return _c
}

func (_c *MockChannel_Get_Call) Return(msg amqp091.Delivery, ok bool, err error) *MockChannel_Get_Call {
	_c.Call.Return(msg, ok, err)
	return _c
}

func (_c *MockChannel_Get_Call) RunAndReturn(run func(string, bool) (amqp091.Delivery, bool, error)) *MockChannel_Get_Call {
	_c.Call.Return(run)
	return _c
}

// IsClosed provides a mock function with given fields:
func (_m *MockChannel) IsClosed() bool {
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

// MockChannel_IsClosed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsClosed'
type MockChannel_IsClosed_Call struct {
	*mock.Call
}

// IsClosed is a helper method to define mock.On call
func (_e *MockChannel_Expecter) IsClosed() *MockChannel_IsClosed_Call {
	return &MockChannel_IsClosed_Call{Call: _e.mock.On("IsClosed")}
}

func (_c *MockChannel_IsClosed_Call) Run(run func()) *MockChannel_IsClosed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockChannel_IsClosed_Call) Return(_a0 bool) *MockChannel_IsClosed_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_IsClosed_Call) RunAndReturn(run func() bool) *MockChannel_IsClosed_Call {
	_c.Call.Return(run)
	return _c
}

// PublishWithContext provides a mock function with given fields: ctx, exchange, key, mandatory, immediate, msg
func (_m *MockChannel) PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp091.Publishing) error {
	ret := _m.Called(ctx, exchange, key, mandatory, immediate, msg)

	if len(ret) == 0 {
		panic("no return value specified for PublishWithContext")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, bool, bool, amqp091.Publishing) error); ok {
		r0 = rf(ctx, exchange, key, mandatory, immediate, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChannel_PublishWithContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PublishWithContext'
type MockChannel_PublishWithContext_Call struct {
	*mock.Call
}

// PublishWithContext is a helper method to define mock.On call
//   - ctx context.Context
//   - exchange string
//   - key string
//   - mandatory bool
//   - immediate bool
//   - msg amqp091.Publishing
func (_e *MockChannel_Expecter) PublishWithContext(ctx interface{}, exchange interface{}, key interface{}, mandatory interface{}, immediate interface{}, msg interface{}) *MockChannel_PublishWithContext_Call {
	return &MockChannel_PublishWithContext_Call{Call: _e.mock.On("PublishWithContext", ctx, exchange, key, mandatory, immediate, msg)}
}

func (_c *MockChannel_PublishWithContext_Call) Run(run func(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp091.Publishing)) *MockChannel_PublishWithContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(bool), args[4].(bool), args[5].(amqp091.Publishing))
	})
	return _c
}

func (_c *MockChannel_PublishWithContext_Call) Return(_a0 error) *MockChannel_PublishWithContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_PublishWithContext_Call) RunAndReturn(run func(context.Context, string, string, bool, bool, amqp091.Publishing) error) *MockChannel_PublishWithContext_Call {
	_c.Call.Return(run)
	return _c
}

// Qos provides a mock function with given fields: prefetchCount, prefetchSize, global
func (_m *MockChannel) Qos(prefetchCount int, prefetchSize int, global bool) error {
	ret := _m.Called(prefetchCount, prefetchSize, global)

	if len(ret) == 0 {
		panic("no return value specified for Qos")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int, int, bool) error); ok {
		r0 = rf(prefetchCount, prefetchSize, global)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChannel_Qos_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Qos'
type MockChannel_Qos_Call struct {
	*mock.Call
}

// Qos is a helper method to define mock.On call
//   - prefetchCount int
//   - prefetchSize int
//   - global bool
func (_e *MockChannel_Expecter) Qos(prefetchCount interface{}, prefetchSize interface{}, global interface{}) *MockChannel_Qos_Call {
	return &MockChannel_Qos_Call{Call: _e.mock.On("Qos", prefetchCount, prefetchSize, global)}
}

func (_c *MockChannel_Qos_Call) Run(run func(prefetchCount int, prefetchSize int, global bool)) *MockChannel_Qos_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int), args[1].(int), args[2].(bool))
	})
	return _c
}

func (_c *MockChannel_Qos_Call) Return(_a0 error) *MockChannel_Qos_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_Qos_Call) RunAndReturn(run func(int, int, bool) error) *MockChannel_Qos_Call {
	_c.Call.Return(run)
	return _c
}

// QueueBind provides a mock function with given fields: name, key, exchange, noWait, args
func (_m *MockChannel) QueueBind(name string, key string, exchange string, noWait bool, args amqp091.Table) error {
	ret := _m.Called(name, key, exchange, noWait, args)

	if len(ret) == 0 {
		panic("no return value specified for QueueBind")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, bool, amqp091.Table) error); ok {
		r0 = rf(name, key, exchange, noWait, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockChannel_QueueBind_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QueueBind'
type MockChannel_QueueBind_Call struct {
	*mock.Call
}

// QueueBind is a helper method to define mock.On call
//   - name string
//   - key string
//   - exchange string
//   - noWait bool
//   - args amqp091.Table
func (_e *MockChannel_Expecter) QueueBind(name interface{}, key interface{}, exchange interface{}, noWait interface{}, args interface{}) *MockChannel_QueueBind_Call {
	return &MockChannel_QueueBind_Call{Call: _e.mock.On("QueueBind", name, key, exchange, noWait, args)}
}

func (_c *MockChannel_QueueBind_Call) Run(run func(name string, key string, exchange string, noWait bool, args amqp091.Table)) *MockChannel_QueueBind_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(bool), args[4].(amqp091.Table))
	})
	return _c
}

func (_c *MockChannel_QueueBind_Call) Return(_a0 error) *MockChannel_QueueBind_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChannel_QueueBind_Call) RunAndReturn(run func(string, string, string, bool, amqp091.Table) error) *MockChannel_QueueBind_Call {
	_c.Call.Return(run)
	return _c
}

// QueueDeclare provides a mock function with given fields: name, durable, autoDelete, exclusive, noWait, args
func (_m *MockChannel) QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	ret := _m.Called(name, durable, autoDelete, exclusive, noWait, args)

	if len(ret) == 0 {
		panic("no return value specified for QueueDeclare")
	}

	var r0 amqp091.Queue
	var r1 error
	if rf, ok := ret.Get(0).(func(string, bool, bool, bool, bool, amqp091.Table) (amqp091.Queue, error)); ok {
		return rf(name, durable, autoDelete, exclusive, noWait, args)
	}
	if rf, ok := ret.Get(0).(func(string, bool, bool, bool, bool, amqp091.Table) amqp091.Queue); ok {
		r0 = rf(name, durable, autoDelete, exclusive, noWait, args)
	} else {
		r0 = ret.Get(0).(amqp091.Queue)
	}

	if rf, ok := ret.Get(1).(func(string, bool, bool, bool, bool, amqp091.Table) error); ok {
		r1 = rf(name, durable, autoDelete, exclusive, noWait, args)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockChannel_QueueDeclare_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QueueDeclare'
type MockChannel_QueueDeclare_Call struct {
	*mock.Call
}

// QueueDeclare is a helper method to define mock.On call
//   - name string
//   - durable bool
//   - autoDelete bool
//   - exclusive bool
//   - noWait bool
//   - args amqp091.Table
func (_e *MockChannel_Expecter) QueueDeclare(name interface{}, durable interface{}, autoDelete interface{}, exclusive interface{}, noWait interface{}, args interface{}) *MockChannel_QueueDeclare_Call {
	return &MockChannel_QueueDeclare_Call{Call: _e.mock.On("QueueDeclare", name, durable, autoDelete, exclusive, noWait, args)}
}

func (_c *MockChannel_QueueDeclare_Call) Run(run func(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp091.Table)) *MockChannel_QueueDeclare_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(bool), args[2].(bool), args[3].(bool), args[4].(bool), args[5].(amqp091.Table))
	})
	return _c
}

func (_c *MockChannel_QueueDeclare_Call) Return(_a0 amqp091.Queue, _a1 error) *MockChannel_QueueDeclare_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockChannel_QueueDeclare_Call) RunAndReturn(run func(string, bool, bool, bool, bool, amqp091.Table) (amqp091.Queue, error)) *MockChannel_QueueDeclare_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockChannel creates a new instance of MockChannel. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockChannel(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockChannel {
	mock := &MockChannel{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}