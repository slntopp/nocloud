// Code generated by mockery v2.43.2. DO NOT EDIT.

package graph_mocks

import (
	context "context"

	billing "github.com/slntopp/nocloud-proto/billing"

	mock "github.com/stretchr/testify/mock"
)

// MockCurrencyController is an autogenerated mock type for the CurrencyController type
type MockCurrencyController struct {
	mock.Mock
}

type MockCurrencyController_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCurrencyController) EXPECT() *MockCurrencyController_Expecter {
	return &MockCurrencyController_Expecter{mock: &_m.Mock}
}

// Convert provides a mock function with given fields: ctx, from, to, amount
func (_m *MockCurrencyController) Convert(ctx context.Context, from *billing.Currency, to *billing.Currency, amount float64) (float64, error) {
	ret := _m.Called(ctx, from, to, amount)

	if len(ret) == 0 {
		panic("no return value specified for Convert")
	}

	var r0 float64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency, *billing.Currency, float64) (float64, error)); ok {
		return rf(ctx, from, to, amount)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency, *billing.Currency, float64) float64); ok {
		r0 = rf(ctx, from, to, amount)
	} else {
		r0 = ret.Get(0).(float64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *billing.Currency, *billing.Currency, float64) error); ok {
		r1 = rf(ctx, from, to, amount)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCurrencyController_Convert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Convert'
type MockCurrencyController_Convert_Call struct {
	*mock.Call
}

// Convert is a helper method to define mock.On call
//   - ctx context.Context
//   - from *billing.Currency
//   - to *billing.Currency
//   - amount float64
func (_e *MockCurrencyController_Expecter) Convert(ctx interface{}, from interface{}, to interface{}, amount interface{}) *MockCurrencyController_Convert_Call {
	return &MockCurrencyController_Convert_Call{Call: _e.mock.On("Convert", ctx, from, to, amount)}
}

func (_c *MockCurrencyController_Convert_Call) Run(run func(ctx context.Context, from *billing.Currency, to *billing.Currency, amount float64)) *MockCurrencyController_Convert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*billing.Currency), args[2].(*billing.Currency), args[3].(float64))
	})
	return _c
}

func (_c *MockCurrencyController_Convert_Call) Return(_a0 float64, _a1 error) *MockCurrencyController_Convert_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCurrencyController_Convert_Call) RunAndReturn(run func(context.Context, *billing.Currency, *billing.Currency, float64) (float64, error)) *MockCurrencyController_Convert_Call {
	_c.Call.Return(run)
	return _c
}

// CreateCurrency provides a mock function with given fields: ctx, currency
func (_m *MockCurrencyController) CreateCurrency(ctx context.Context, currency *billing.Currency) error {
	ret := _m.Called(ctx, currency)

	if len(ret) == 0 {
		panic("no return value specified for CreateCurrency")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency) error); ok {
		r0 = rf(ctx, currency)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCurrencyController_CreateCurrency_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateCurrency'
type MockCurrencyController_CreateCurrency_Call struct {
	*mock.Call
}

// CreateCurrency is a helper method to define mock.On call
//   - ctx context.Context
//   - currency *billing.Currency
func (_e *MockCurrencyController_Expecter) CreateCurrency(ctx interface{}, currency interface{}) *MockCurrencyController_CreateCurrency_Call {
	return &MockCurrencyController_CreateCurrency_Call{Call: _e.mock.On("CreateCurrency", ctx, currency)}
}

func (_c *MockCurrencyController_CreateCurrency_Call) Run(run func(ctx context.Context, currency *billing.Currency)) *MockCurrencyController_CreateCurrency_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*billing.Currency))
	})
	return _c
}

func (_c *MockCurrencyController_CreateCurrency_Call) Return(_a0 error) *MockCurrencyController_CreateCurrency_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCurrencyController_CreateCurrency_Call) RunAndReturn(run func(context.Context, *billing.Currency) error) *MockCurrencyController_CreateCurrency_Call {
	_c.Call.Return(run)
	return _c
}

// CreateExchangeRate provides a mock function with given fields: ctx, from, to, rate, commission
func (_m *MockCurrencyController) CreateExchangeRate(ctx context.Context, from billing.Currency, to billing.Currency, rate float64, commission float64) error {
	ret := _m.Called(ctx, from, to, rate, commission)

	if len(ret) == 0 {
		panic("no return value specified for CreateExchangeRate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, billing.Currency, billing.Currency, float64, float64) error); ok {
		r0 = rf(ctx, from, to, rate, commission)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCurrencyController_CreateExchangeRate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateExchangeRate'
type MockCurrencyController_CreateExchangeRate_Call struct {
	*mock.Call
}

// CreateExchangeRate is a helper method to define mock.On call
//   - ctx context.Context
//   - from billing.Currency
//   - to billing.Currency
//   - rate float64
//   - commission float64
func (_e *MockCurrencyController_Expecter) CreateExchangeRate(ctx interface{}, from interface{}, to interface{}, rate interface{}, commission interface{}) *MockCurrencyController_CreateExchangeRate_Call {
	return &MockCurrencyController_CreateExchangeRate_Call{Call: _e.mock.On("CreateExchangeRate", ctx, from, to, rate, commission)}
}

func (_c *MockCurrencyController_CreateExchangeRate_Call) Run(run func(ctx context.Context, from billing.Currency, to billing.Currency, rate float64, commission float64)) *MockCurrencyController_CreateExchangeRate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(billing.Currency), args[2].(billing.Currency), args[3].(float64), args[4].(float64))
	})
	return _c
}

func (_c *MockCurrencyController_CreateExchangeRate_Call) Return(_a0 error) *MockCurrencyController_CreateExchangeRate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCurrencyController_CreateExchangeRate_Call) RunAndReturn(run func(context.Context, billing.Currency, billing.Currency, float64, float64) error) *MockCurrencyController_CreateExchangeRate_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteExchangeRate provides a mock function with given fields: ctx, from, to
func (_m *MockCurrencyController) DeleteExchangeRate(ctx context.Context, from *billing.Currency, to *billing.Currency) error {
	ret := _m.Called(ctx, from, to)

	if len(ret) == 0 {
		panic("no return value specified for DeleteExchangeRate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency, *billing.Currency) error); ok {
		r0 = rf(ctx, from, to)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCurrencyController_DeleteExchangeRate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteExchangeRate'
type MockCurrencyController_DeleteExchangeRate_Call struct {
	*mock.Call
}

// DeleteExchangeRate is a helper method to define mock.On call
//   - ctx context.Context
//   - from *billing.Currency
//   - to *billing.Currency
func (_e *MockCurrencyController_Expecter) DeleteExchangeRate(ctx interface{}, from interface{}, to interface{}) *MockCurrencyController_DeleteExchangeRate_Call {
	return &MockCurrencyController_DeleteExchangeRate_Call{Call: _e.mock.On("DeleteExchangeRate", ctx, from, to)}
}

func (_c *MockCurrencyController_DeleteExchangeRate_Call) Run(run func(ctx context.Context, from *billing.Currency, to *billing.Currency)) *MockCurrencyController_DeleteExchangeRate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*billing.Currency), args[2].(*billing.Currency))
	})
	return _c
}

func (_c *MockCurrencyController_DeleteExchangeRate_Call) Return(_a0 error) *MockCurrencyController_DeleteExchangeRate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCurrencyController_DeleteExchangeRate_Call) RunAndReturn(run func(context.Context, *billing.Currency, *billing.Currency) error) *MockCurrencyController_DeleteExchangeRate_Call {
	_c.Call.Return(run)
	return _c
}

// GetCurrencies provides a mock function with given fields: ctx, isAdmin
func (_m *MockCurrencyController) GetCurrencies(ctx context.Context, isAdmin bool) ([]*billing.Currency, error) {
	ret := _m.Called(ctx, isAdmin)

	if len(ret) == 0 {
		panic("no return value specified for GetCurrencies")
	}

	var r0 []*billing.Currency
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, bool) ([]*billing.Currency, error)); ok {
		return rf(ctx, isAdmin)
	}
	if rf, ok := ret.Get(0).(func(context.Context, bool) []*billing.Currency); ok {
		r0 = rf(ctx, isAdmin)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*billing.Currency)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, bool) error); ok {
		r1 = rf(ctx, isAdmin)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCurrencyController_GetCurrencies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCurrencies'
type MockCurrencyController_GetCurrencies_Call struct {
	*mock.Call
}

// GetCurrencies is a helper method to define mock.On call
//   - ctx context.Context
//   - isAdmin bool
func (_e *MockCurrencyController_Expecter) GetCurrencies(ctx interface{}, isAdmin interface{}) *MockCurrencyController_GetCurrencies_Call {
	return &MockCurrencyController_GetCurrencies_Call{Call: _e.mock.On("GetCurrencies", ctx, isAdmin)}
}

func (_c *MockCurrencyController_GetCurrencies_Call) Run(run func(ctx context.Context, isAdmin bool)) *MockCurrencyController_GetCurrencies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(bool))
	})
	return _c
}

func (_c *MockCurrencyController_GetCurrencies_Call) Return(_a0 []*billing.Currency, _a1 error) *MockCurrencyController_GetCurrencies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCurrencyController_GetCurrencies_Call) RunAndReturn(run func(context.Context, bool) ([]*billing.Currency, error)) *MockCurrencyController_GetCurrencies_Call {
	_c.Call.Return(run)
	return _c
}

// GetExchangeRate provides a mock function with given fields: ctx, from, to
func (_m *MockCurrencyController) GetExchangeRate(ctx context.Context, from *billing.Currency, to *billing.Currency) (float64, float64, error) {
	ret := _m.Called(ctx, from, to)

	if len(ret) == 0 {
		panic("no return value specified for GetExchangeRate")
	}

	var r0 float64
	var r1 float64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency, *billing.Currency) (float64, float64, error)); ok {
		return rf(ctx, from, to)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency, *billing.Currency) float64); ok {
		r0 = rf(ctx, from, to)
	} else {
		r0 = ret.Get(0).(float64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *billing.Currency, *billing.Currency) float64); ok {
		r1 = rf(ctx, from, to)
	} else {
		r1 = ret.Get(1).(float64)
	}

	if rf, ok := ret.Get(2).(func(context.Context, *billing.Currency, *billing.Currency) error); ok {
		r2 = rf(ctx, from, to)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockCurrencyController_GetExchangeRate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExchangeRate'
type MockCurrencyController_GetExchangeRate_Call struct {
	*mock.Call
}

// GetExchangeRate is a helper method to define mock.On call
//   - ctx context.Context
//   - from *billing.Currency
//   - to *billing.Currency
func (_e *MockCurrencyController_Expecter) GetExchangeRate(ctx interface{}, from interface{}, to interface{}) *MockCurrencyController_GetExchangeRate_Call {
	return &MockCurrencyController_GetExchangeRate_Call{Call: _e.mock.On("GetExchangeRate", ctx, from, to)}
}

func (_c *MockCurrencyController_GetExchangeRate_Call) Run(run func(ctx context.Context, from *billing.Currency, to *billing.Currency)) *MockCurrencyController_GetExchangeRate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*billing.Currency), args[2].(*billing.Currency))
	})
	return _c
}

func (_c *MockCurrencyController_GetExchangeRate_Call) Return(_a0 float64, _a1 float64, _a2 error) *MockCurrencyController_GetExchangeRate_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockCurrencyController_GetExchangeRate_Call) RunAndReturn(run func(context.Context, *billing.Currency, *billing.Currency) (float64, float64, error)) *MockCurrencyController_GetExchangeRate_Call {
	_c.Call.Return(run)
	return _c
}

// GetExchangeRateDirect provides a mock function with given fields: ctx, from, to
func (_m *MockCurrencyController) GetExchangeRateDirect(ctx context.Context, from billing.Currency, to billing.Currency) (float64, float64, error) {
	ret := _m.Called(ctx, from, to)

	if len(ret) == 0 {
		panic("no return value specified for GetExchangeRateDirect")
	}

	var r0 float64
	var r1 float64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, billing.Currency, billing.Currency) (float64, float64, error)); ok {
		return rf(ctx, from, to)
	}
	if rf, ok := ret.Get(0).(func(context.Context, billing.Currency, billing.Currency) float64); ok {
		r0 = rf(ctx, from, to)
	} else {
		r0 = ret.Get(0).(float64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, billing.Currency, billing.Currency) float64); ok {
		r1 = rf(ctx, from, to)
	} else {
		r1 = ret.Get(1).(float64)
	}

	if rf, ok := ret.Get(2).(func(context.Context, billing.Currency, billing.Currency) error); ok {
		r2 = rf(ctx, from, to)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockCurrencyController_GetExchangeRateDirect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExchangeRateDirect'
type MockCurrencyController_GetExchangeRateDirect_Call struct {
	*mock.Call
}

// GetExchangeRateDirect is a helper method to define mock.On call
//   - ctx context.Context
//   - from billing.Currency
//   - to billing.Currency
func (_e *MockCurrencyController_Expecter) GetExchangeRateDirect(ctx interface{}, from interface{}, to interface{}) *MockCurrencyController_GetExchangeRateDirect_Call {
	return &MockCurrencyController_GetExchangeRateDirect_Call{Call: _e.mock.On("GetExchangeRateDirect", ctx, from, to)}
}

func (_c *MockCurrencyController_GetExchangeRateDirect_Call) Run(run func(ctx context.Context, from billing.Currency, to billing.Currency)) *MockCurrencyController_GetExchangeRateDirect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(billing.Currency), args[2].(billing.Currency))
	})
	return _c
}

func (_c *MockCurrencyController_GetExchangeRateDirect_Call) Return(_a0 float64, _a1 float64, _a2 error) *MockCurrencyController_GetExchangeRateDirect_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockCurrencyController_GetExchangeRateDirect_Call) RunAndReturn(run func(context.Context, billing.Currency, billing.Currency) (float64, float64, error)) *MockCurrencyController_GetExchangeRateDirect_Call {
	_c.Call.Return(run)
	return _c
}

// GetExchangeRates provides a mock function with given fields: ctx
func (_m *MockCurrencyController) GetExchangeRates(ctx context.Context) ([]*billing.GetExchangeRateResponse, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetExchangeRates")
	}

	var r0 []*billing.GetExchangeRateResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*billing.GetExchangeRateResponse, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*billing.GetExchangeRateResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*billing.GetExchangeRateResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCurrencyController_GetExchangeRates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExchangeRates'
type MockCurrencyController_GetExchangeRates_Call struct {
	*mock.Call
}

// GetExchangeRates is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockCurrencyController_Expecter) GetExchangeRates(ctx interface{}) *MockCurrencyController_GetExchangeRates_Call {
	return &MockCurrencyController_GetExchangeRates_Call{Call: _e.mock.On("GetExchangeRates", ctx)}
}

func (_c *MockCurrencyController_GetExchangeRates_Call) Run(run func(ctx context.Context)) *MockCurrencyController_GetExchangeRates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockCurrencyController_GetExchangeRates_Call) Return(_a0 []*billing.GetExchangeRateResponse, _a1 error) *MockCurrencyController_GetExchangeRates_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCurrencyController_GetExchangeRates_Call) RunAndReturn(run func(context.Context) ([]*billing.GetExchangeRateResponse, error)) *MockCurrencyController_GetExchangeRates_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateCurrency provides a mock function with given fields: ctx, currency
func (_m *MockCurrencyController) UpdateCurrency(ctx context.Context, currency *billing.Currency) error {
	ret := _m.Called(ctx, currency)

	if len(ret) == 0 {
		panic("no return value specified for UpdateCurrency")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *billing.Currency) error); ok {
		r0 = rf(ctx, currency)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCurrencyController_UpdateCurrency_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateCurrency'
type MockCurrencyController_UpdateCurrency_Call struct {
	*mock.Call
}

// UpdateCurrency is a helper method to define mock.On call
//   - ctx context.Context
//   - currency *billing.Currency
func (_e *MockCurrencyController_Expecter) UpdateCurrency(ctx interface{}, currency interface{}) *MockCurrencyController_UpdateCurrency_Call {
	return &MockCurrencyController_UpdateCurrency_Call{Call: _e.mock.On("UpdateCurrency", ctx, currency)}
}

func (_c *MockCurrencyController_UpdateCurrency_Call) Run(run func(ctx context.Context, currency *billing.Currency)) *MockCurrencyController_UpdateCurrency_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*billing.Currency))
	})
	return _c
}

func (_c *MockCurrencyController_UpdateCurrency_Call) Return(_a0 error) *MockCurrencyController_UpdateCurrency_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCurrencyController_UpdateCurrency_Call) RunAndReturn(run func(context.Context, *billing.Currency) error) *MockCurrencyController_UpdateCurrency_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateExchangeRate provides a mock function with given fields: ctx, from, to, rate, commission
func (_m *MockCurrencyController) UpdateExchangeRate(ctx context.Context, from billing.Currency, to billing.Currency, rate float64, commission float64) error {
	ret := _m.Called(ctx, from, to, rate, commission)

	if len(ret) == 0 {
		panic("no return value specified for UpdateExchangeRate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, billing.Currency, billing.Currency, float64, float64) error); ok {
		r0 = rf(ctx, from, to, rate, commission)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCurrencyController_UpdateExchangeRate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateExchangeRate'
type MockCurrencyController_UpdateExchangeRate_Call struct {
	*mock.Call
}

// UpdateExchangeRate is a helper method to define mock.On call
//   - ctx context.Context
//   - from billing.Currency
//   - to billing.Currency
//   - rate float64
//   - commission float64
func (_e *MockCurrencyController_Expecter) UpdateExchangeRate(ctx interface{}, from interface{}, to interface{}, rate interface{}, commission interface{}) *MockCurrencyController_UpdateExchangeRate_Call {
	return &MockCurrencyController_UpdateExchangeRate_Call{Call: _e.mock.On("UpdateExchangeRate", ctx, from, to, rate, commission)}
}

func (_c *MockCurrencyController_UpdateExchangeRate_Call) Run(run func(ctx context.Context, from billing.Currency, to billing.Currency, rate float64, commission float64)) *MockCurrencyController_UpdateExchangeRate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(billing.Currency), args[2].(billing.Currency), args[3].(float64), args[4].(float64))
	})
	return _c
}

func (_c *MockCurrencyController_UpdateExchangeRate_Call) Return(_a0 error) *MockCurrencyController_UpdateExchangeRate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCurrencyController_UpdateExchangeRate_Call) RunAndReturn(run func(context.Context, billing.Currency, billing.Currency, float64, float64) error) *MockCurrencyController_UpdateExchangeRate_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCurrencyController creates a new instance of MockCurrencyController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCurrencyController(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCurrencyController {
	mock := &MockCurrencyController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
