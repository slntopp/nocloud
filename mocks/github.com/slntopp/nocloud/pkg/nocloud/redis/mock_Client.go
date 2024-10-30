// Code generated by mockery v2.43.2. DO NOT EDIT.

package redisdb_mocks

import (
	context "context"

	redis "github.com/go-redis/redis/v8"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

type MockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClient) EXPECT() *MockClient_Expecter {
	return &MockClient_Expecter{mock: &_m.Mock}
}

// Del provides a mock function with given fields: ctx, keys
func (_m *MockClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	_va := make([]interface{}, len(keys))
	for _i := range keys {
		_va[_i] = keys[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Del")
	}

	var r0 *redis.IntCmd
	if rf, ok := ret.Get(0).(func(context.Context, ...string) *redis.IntCmd); ok {
		r0 = rf(ctx, keys...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.IntCmd)
		}
	}

	return r0
}

// MockClient_Del_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Del'
type MockClient_Del_Call struct {
	*mock.Call
}

// Del is a helper method to define mock.On call
//   - ctx context.Context
//   - keys ...string
func (_e *MockClient_Expecter) Del(ctx interface{}, keys ...interface{}) *MockClient_Del_Call {
	return &MockClient_Del_Call{Call: _e.mock.On("Del",
		append([]interface{}{ctx}, keys...)...)}
}

func (_c *MockClient_Del_Call) Run(run func(ctx context.Context, keys ...string)) *MockClient_Del_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_Del_Call) Return(_a0 *redis.IntCmd) *MockClient_Del_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_Del_Call) RunAndReturn(run func(context.Context, ...string) *redis.IntCmd) *MockClient_Del_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *MockClient) Get(ctx context.Context, key string) *redis.StringCmd {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *redis.StringCmd
	if rf, ok := ret.Get(0).(func(context.Context, string) *redis.StringCmd); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StringCmd)
		}
	}

	return r0
}

// MockClient_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockClient_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockClient_Expecter) Get(ctx interface{}, key interface{}) *MockClient_Get_Call {
	return &MockClient_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *MockClient_Get_Call) Run(run func(ctx context.Context, key string)) *MockClient_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockClient_Get_Call) Return(_a0 *redis.StringCmd) *MockClient_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_Get_Call) RunAndReturn(run func(context.Context, string) *redis.StringCmd) *MockClient_Get_Call {
	_c.Call.Return(run)
	return _c
}

// HGetAll provides a mock function with given fields: ctx, key
func (_m *MockClient) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for HGetAll")
	}

	var r0 *redis.StringStringMapCmd
	if rf, ok := ret.Get(0).(func(context.Context, string) *redis.StringStringMapCmd); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StringStringMapCmd)
		}
	}

	return r0
}

// MockClient_HGetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HGetAll'
type MockClient_HGetAll_Call struct {
	*mock.Call
}

// HGetAll is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockClient_Expecter) HGetAll(ctx interface{}, key interface{}) *MockClient_HGetAll_Call {
	return &MockClient_HGetAll_Call{Call: _e.mock.On("HGetAll", ctx, key)}
}

func (_c *MockClient_HGetAll_Call) Run(run func(ctx context.Context, key string)) *MockClient_HGetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockClient_HGetAll_Call) Return(_a0 *redis.StringStringMapCmd) *MockClient_HGetAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_HGetAll_Call) RunAndReturn(run func(context.Context, string) *redis.StringStringMapCmd) *MockClient_HGetAll_Call {
	_c.Call.Return(run)
	return _c
}

// HSet provides a mock function with given fields: ctx, key, values
func (_m *MockClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	var _ca []interface{}
	_ca = append(_ca, ctx, key)
	_ca = append(_ca, values...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for HSet")
	}

	var r0 *redis.IntCmd
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) *redis.IntCmd); ok {
		r0 = rf(ctx, key, values...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.IntCmd)
		}
	}

	return r0
}

// MockClient_HSet_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HSet'
type MockClient_HSet_Call struct {
	*mock.Call
}

// HSet is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - values ...interface{}
func (_e *MockClient_Expecter) HSet(ctx interface{}, key interface{}, values ...interface{}) *MockClient_HSet_Call {
	return &MockClient_HSet_Call{Call: _e.mock.On("HSet",
		append([]interface{}{ctx, key}, values...)...)}
}

func (_c *MockClient_HSet_Call) Run(run func(ctx context.Context, key string, values ...interface{})) *MockClient_HSet_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_HSet_Call) Return(_a0 *redis.IntCmd) *MockClient_HSet_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_HSet_Call) RunAndReturn(run func(context.Context, string, ...interface{}) *redis.IntCmd) *MockClient_HSet_Call {
	_c.Call.Return(run)
	return _c
}

// Keys provides a mock function with given fields: ctx, pattern
func (_m *MockClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	ret := _m.Called(ctx, pattern)

	if len(ret) == 0 {
		panic("no return value specified for Keys")
	}

	var r0 *redis.StringSliceCmd
	if rf, ok := ret.Get(0).(func(context.Context, string) *redis.StringSliceCmd); ok {
		r0 = rf(ctx, pattern)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StringSliceCmd)
		}
	}

	return r0
}

// MockClient_Keys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Keys'
type MockClient_Keys_Call struct {
	*mock.Call
}

// Keys is a helper method to define mock.On call
//   - ctx context.Context
//   - pattern string
func (_e *MockClient_Expecter) Keys(ctx interface{}, pattern interface{}) *MockClient_Keys_Call {
	return &MockClient_Keys_Call{Call: _e.mock.On("Keys", ctx, pattern)}
}

func (_c *MockClient_Keys_Call) Run(run func(ctx context.Context, pattern string)) *MockClient_Keys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockClient_Keys_Call) Return(_a0 *redis.StringSliceCmd) *MockClient_Keys_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_Keys_Call) RunAndReturn(run func(context.Context, string) *redis.StringSliceCmd) *MockClient_Keys_Call {
	_c.Call.Return(run)
	return _c
}

// MGet provides a mock function with given fields: ctx, keys
func (_m *MockClient) MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	_va := make([]interface{}, len(keys))
	for _i := range keys {
		_va[_i] = keys[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for MGet")
	}

	var r0 *redis.SliceCmd
	if rf, ok := ret.Get(0).(func(context.Context, ...string) *redis.SliceCmd); ok {
		r0 = rf(ctx, keys...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.SliceCmd)
		}
	}

	return r0
}

// MockClient_MGet_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MGet'
type MockClient_MGet_Call struct {
	*mock.Call
}

// MGet is a helper method to define mock.On call
//   - ctx context.Context
//   - keys ...string
func (_e *MockClient_Expecter) MGet(ctx interface{}, keys ...interface{}) *MockClient_MGet_Call {
	return &MockClient_MGet_Call{Call: _e.mock.On("MGet",
		append([]interface{}{ctx}, keys...)...)}
}

func (_c *MockClient_MGet_Call) Run(run func(ctx context.Context, keys ...string)) *MockClient_MGet_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_MGet_Call) Return(_a0 *redis.SliceCmd) *MockClient_MGet_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_MGet_Call) RunAndReturn(run func(context.Context, ...string) *redis.SliceCmd) *MockClient_MGet_Call {
	_c.Call.Return(run)
	return _c
}

// Options provides a mock function with given fields:
func (_m *MockClient) Options() *redis.Options {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Options")
	}

	var r0 *redis.Options
	if rf, ok := ret.Get(0).(func() *redis.Options); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.Options)
		}
	}

	return r0
}

// MockClient_Options_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Options'
type MockClient_Options_Call struct {
	*mock.Call
}

// Options is a helper method to define mock.On call
func (_e *MockClient_Expecter) Options() *MockClient_Options_Call {
	return &MockClient_Options_Call{Call: _e.mock.On("Options")}
}

func (_c *MockClient_Options_Call) Run(run func()) *MockClient_Options_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockClient_Options_Call) Return(_a0 *redis.Options) *MockClient_Options_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_Options_Call) RunAndReturn(run func() *redis.Options) *MockClient_Options_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: ctx, key, value, expiration
func (_m *MockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	ret := _m.Called(ctx, key, value, expiration)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 *redis.StatusCmd
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}, time.Duration) *redis.StatusCmd); ok {
		r0 = rf(ctx, key, value, expiration)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StatusCmd)
		}
	}

	return r0
}

// MockClient_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type MockClient_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value interface{}
//   - expiration time.Duration
func (_e *MockClient_Expecter) Set(ctx interface{}, key interface{}, value interface{}, expiration interface{}) *MockClient_Set_Call {
	return &MockClient_Set_Call{Call: _e.mock.On("Set", ctx, key, value, expiration)}
}

func (_c *MockClient_Set_Call) Run(run func(ctx context.Context, key string, value interface{}, expiration time.Duration)) *MockClient_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(interface{}), args[3].(time.Duration))
	})
	return _c
}

func (_c *MockClient_Set_Call) Return(_a0 *redis.StatusCmd) *MockClient_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_Set_Call) RunAndReturn(run func(context.Context, string, interface{}, time.Duration) *redis.StatusCmd) *MockClient_Set_Call {
	_c.Call.Return(run)
	return _c
}

// Subscribe provides a mock function with given fields: ctx, channels
func (_m *MockClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	_va := make([]interface{}, len(channels))
	for _i := range channels {
		_va[_i] = channels[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Subscribe")
	}

	var r0 *redis.PubSub
	if rf, ok := ret.Get(0).(func(context.Context, ...string) *redis.PubSub); ok {
		r0 = rf(ctx, channels...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.PubSub)
		}
	}

	return r0
}

// MockClient_Subscribe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Subscribe'
type MockClient_Subscribe_Call struct {
	*mock.Call
}

// Subscribe is a helper method to define mock.On call
//   - ctx context.Context
//   - channels ...string
func (_e *MockClient_Expecter) Subscribe(ctx interface{}, channels ...interface{}) *MockClient_Subscribe_Call {
	return &MockClient_Subscribe_Call{Call: _e.mock.On("Subscribe",
		append([]interface{}{ctx}, channels...)...)}
}

func (_c *MockClient_Subscribe_Call) Run(run func(ctx context.Context, channels ...string)) *MockClient_Subscribe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_Subscribe_Call) Return(_a0 *redis.PubSub) *MockClient_Subscribe_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_Subscribe_Call) RunAndReturn(run func(context.Context, ...string) *redis.PubSub) *MockClient_Subscribe_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
