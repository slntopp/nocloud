// Code generated by mockery v2.43.2. DO NOT EDIT.

package graph_mocks

import (
	context "context"

	driver "github.com/arangodb/go-driver"
	graph "github.com/slntopp/nocloud/pkg/graph"

	instances "github.com/slntopp/nocloud-proto/instances"

	mock "github.com/stretchr/testify/mock"

	states "github.com/slntopp/nocloud-proto/states"

	statuses "github.com/slntopp/nocloud-proto/statuses"
)

// MockInstancesController is an autogenerated mock type for the InstancesController type
type MockInstancesController struct {
	mock.Mock
}

type MockInstancesController_Expecter struct {
	mock *mock.Mock
}

func (_m *MockInstancesController) EXPECT() *MockInstancesController_Expecter {
	return &MockInstancesController_Expecter{mock: &_m.Mock}
}

// CalculateInstanceEstimatePrice provides a mock function with given fields: i, includeOneTimePayments
func (_m *MockInstancesController) CalculateInstanceEstimatePrice(i *instances.Instance, includeOneTimePayments bool) (float64, error) {
	ret := _m.Called(i, includeOneTimePayments)

	if len(ret) == 0 {
		panic("no return value specified for CalculateInstanceEstimatePrice")
	}

	var r0 float64
	var r1 error
	if rf, ok := ret.Get(0).(func(*instances.Instance, bool) (float64, error)); ok {
		return rf(i, includeOneTimePayments)
	}
	if rf, ok := ret.Get(0).(func(*instances.Instance, bool) float64); ok {
		r0 = rf(i, includeOneTimePayments)
	} else {
		r0 = ret.Get(0).(float64)
	}

	if rf, ok := ret.Get(1).(func(*instances.Instance, bool) error); ok {
		r1 = rf(i, includeOneTimePayments)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_CalculateInstanceEstimatePrice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CalculateInstanceEstimatePrice'
type MockInstancesController_CalculateInstanceEstimatePrice_Call struct {
	*mock.Call
}

// CalculateInstanceEstimatePrice is a helper method to define mock.On call
//   - i *instances.Instance
//   - includeOneTimePayments bool
func (_e *MockInstancesController_Expecter) CalculateInstanceEstimatePrice(i interface{}, includeOneTimePayments interface{}) *MockInstancesController_CalculateInstanceEstimatePrice_Call {
	return &MockInstancesController_CalculateInstanceEstimatePrice_Call{Call: _e.mock.On("CalculateInstanceEstimatePrice", i, includeOneTimePayments)}
}

func (_c *MockInstancesController_CalculateInstanceEstimatePrice_Call) Run(run func(i *instances.Instance, includeOneTimePayments bool)) *MockInstancesController_CalculateInstanceEstimatePrice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*instances.Instance), args[1].(bool))
	})
	return _c
}

func (_c *MockInstancesController_CalculateInstanceEstimatePrice_Call) Return(_a0 float64, _a1 error) *MockInstancesController_CalculateInstanceEstimatePrice_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_CalculateInstanceEstimatePrice_Call) RunAndReturn(run func(*instances.Instance, bool) (float64, error)) *MockInstancesController_CalculateInstanceEstimatePrice_Call {
	_c.Call.Return(run)
	return _c
}

// CheckEdgeExist provides a mock function with given fields: ctx, spUuid, i
func (_m *MockInstancesController) CheckEdgeExist(ctx context.Context, spUuid string, i *instances.Instance) error {
	ret := _m.Called(ctx, spUuid, i)

	if len(ret) == 0 {
		panic("no return value specified for CheckEdgeExist")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *instances.Instance) error); ok {
		r0 = rf(ctx, spUuid, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_CheckEdgeExist_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckEdgeExist'
type MockInstancesController_CheckEdgeExist_Call struct {
	*mock.Call
}

// CheckEdgeExist is a helper method to define mock.On call
//   - ctx context.Context
//   - spUuid string
//   - i *instances.Instance
func (_e *MockInstancesController_Expecter) CheckEdgeExist(ctx interface{}, spUuid interface{}, i interface{}) *MockInstancesController_CheckEdgeExist_Call {
	return &MockInstancesController_CheckEdgeExist_Call{Call: _e.mock.On("CheckEdgeExist", ctx, spUuid, i)}
}

func (_c *MockInstancesController_CheckEdgeExist_Call) Run(run func(ctx context.Context, spUuid string, i *instances.Instance)) *MockInstancesController_CheckEdgeExist_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_CheckEdgeExist_Call) Return(_a0 error) *MockInstancesController_CheckEdgeExist_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstancesController_CheckEdgeExist_Call) RunAndReturn(run func(context.Context, string, *instances.Instance) error) *MockInstancesController_CheckEdgeExist_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: ctx, group, sp, i
func (_m *MockInstancesController) Create(ctx context.Context, group driver.DocumentID, sp string, i *instances.Instance) (string, error) {
	ret := _m.Called(ctx, group, sp, i)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, driver.DocumentID, string, *instances.Instance) (string, error)); ok {
		return rf(ctx, group, sp, i)
	}
	if rf, ok := ret.Get(0).(func(context.Context, driver.DocumentID, string, *instances.Instance) string); ok {
		r0 = rf(ctx, group, sp, i)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, driver.DocumentID, string, *instances.Instance) error); ok {
		r1 = rf(ctx, group, sp, i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockInstancesController_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - group driver.DocumentID
//   - sp string
//   - i *instances.Instance
func (_e *MockInstancesController_Expecter) Create(ctx interface{}, group interface{}, sp interface{}, i interface{}) *MockInstancesController_Create_Call {
	return &MockInstancesController_Create_Call{Call: _e.mock.On("Create", ctx, group, sp, i)}
}

func (_c *MockInstancesController_Create_Call) Run(run func(ctx context.Context, group driver.DocumentID, sp string, i *instances.Instance)) *MockInstancesController_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(driver.DocumentID), args[2].(string), args[3].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_Create_Call) Return(_a0 string, _a1 error) *MockInstancesController_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_Create_Call) RunAndReturn(run func(context.Context, driver.DocumentID, string, *instances.Instance) (string, error)) *MockInstancesController_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, group, i
func (_m *MockInstancesController) Delete(ctx context.Context, group string, i *instances.Instance) error {
	ret := _m.Called(ctx, group, i)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *instances.Instance) error); ok {
		r0 = rf(ctx, group, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockInstancesController_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - group string
//   - i *instances.Instance
func (_e *MockInstancesController_Expecter) Delete(ctx interface{}, group interface{}, i interface{}) *MockInstancesController_Delete_Call {
	return &MockInstancesController_Delete_Call{Call: _e.mock.On("Delete", ctx, group, i)}
}

func (_c *MockInstancesController_Delete_Call) Run(run func(ctx context.Context, group string, i *instances.Instance)) *MockInstancesController_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_Delete_Call) Return(_a0 error) *MockInstancesController_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstancesController_Delete_Call) RunAndReturn(run func(context.Context, string, *instances.Instance) error) *MockInstancesController_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, uuid
func (_m *MockInstancesController) Get(ctx context.Context, uuid string) (*graph.Instance, error) {
	ret := _m.Called(ctx, uuid)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *graph.Instance
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*graph.Instance, error)); ok {
		return rf(ctx, uuid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *graph.Instance); ok {
		r0 = rf(ctx, uuid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graph.Instance)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uuid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockInstancesController_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - uuid string
func (_e *MockInstancesController_Expecter) Get(ctx interface{}, uuid interface{}) *MockInstancesController_Get_Call {
	return &MockInstancesController_Get_Call{Call: _e.mock.On("Get", ctx, uuid)}
}

func (_c *MockInstancesController_Get_Call) Run(run func(ctx context.Context, uuid string)) *MockInstancesController_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockInstancesController_Get_Call) Return(_a0 *graph.Instance, _a1 error) *MockInstancesController_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_Get_Call) RunAndReturn(run func(context.Context, string) (*graph.Instance, error)) *MockInstancesController_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetEdge provides a mock function with given fields: ctx, inboundNode, collection
func (_m *MockInstancesController) GetEdge(ctx context.Context, inboundNode string, collection string) (string, error) {
	ret := _m.Called(ctx, inboundNode, collection)

	if len(ret) == 0 {
		panic("no return value specified for GetEdge")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, error)); ok {
		return rf(ctx, inboundNode, collection)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, inboundNode, collection)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, inboundNode, collection)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_GetEdge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEdge'
type MockInstancesController_GetEdge_Call struct {
	*mock.Call
}

// GetEdge is a helper method to define mock.On call
//   - ctx context.Context
//   - inboundNode string
//   - collection string
func (_e *MockInstancesController_Expecter) GetEdge(ctx interface{}, inboundNode interface{}, collection interface{}) *MockInstancesController_GetEdge_Call {
	return &MockInstancesController_GetEdge_Call{Call: _e.mock.On("GetEdge", ctx, inboundNode, collection)}
}

func (_c *MockInstancesController_GetEdge_Call) Run(run func(ctx context.Context, inboundNode string, collection string)) *MockInstancesController_GetEdge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockInstancesController_GetEdge_Call) Return(_a0 string, _a1 error) *MockInstancesController_GetEdge_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_GetEdge_Call) RunAndReturn(run func(context.Context, string, string) (string, error)) *MockInstancesController_GetEdge_Call {
	_c.Call.Return(run)
	return _c
}

// GetGroup provides a mock function with given fields: ctx, i
func (_m *MockInstancesController) GetGroup(ctx context.Context, i string) (*graph.GroupWithSP, error) {
	ret := _m.Called(ctx, i)

	if len(ret) == 0 {
		panic("no return value specified for GetGroup")
	}

	var r0 *graph.GroupWithSP
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*graph.GroupWithSP, error)); ok {
		return rf(ctx, i)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *graph.GroupWithSP); ok {
		r0 = rf(ctx, i)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graph.GroupWithSP)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_GetGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGroup'
type MockInstancesController_GetGroup_Call struct {
	*mock.Call
}

// GetGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - i string
func (_e *MockInstancesController_Expecter) GetGroup(ctx interface{}, i interface{}) *MockInstancesController_GetGroup_Call {
	return &MockInstancesController_GetGroup_Call{Call: _e.mock.On("GetGroup", ctx, i)}
}

func (_c *MockInstancesController_GetGroup_Call) Run(run func(ctx context.Context, i string)) *MockInstancesController_GetGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockInstancesController_GetGroup_Call) Return(_a0 *graph.GroupWithSP, _a1 error) *MockInstancesController_GetGroup_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_GetGroup_Call) RunAndReturn(run func(context.Context, string) (*graph.GroupWithSP, error)) *MockInstancesController_GetGroup_Call {
	_c.Call.Return(run)
	return _c
}

// GetInstancePeriod provides a mock function with given fields: i
func (_m *MockInstancesController) GetInstancePeriod(i *instances.Instance) (*int64, error) {
	ret := _m.Called(i)

	if len(ret) == 0 {
		panic("no return value specified for GetInstancePeriod")
	}

	var r0 *int64
	var r1 error
	if rf, ok := ret.Get(0).(func(*instances.Instance) (*int64, error)); ok {
		return rf(i)
	}
	if rf, ok := ret.Get(0).(func(*instances.Instance) *int64); ok {
		r0 = rf(i)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*int64)
		}
	}

	if rf, ok := ret.Get(1).(func(*instances.Instance) error); ok {
		r1 = rf(i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_GetInstancePeriod_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInstancePeriod'
type MockInstancesController_GetInstancePeriod_Call struct {
	*mock.Call
}

// GetInstancePeriod is a helper method to define mock.On call
//   - i *instances.Instance
func (_e *MockInstancesController_Expecter) GetInstancePeriod(i interface{}) *MockInstancesController_GetInstancePeriod_Call {
	return &MockInstancesController_GetInstancePeriod_Call{Call: _e.mock.On("GetInstancePeriod", i)}
}

func (_c *MockInstancesController_GetInstancePeriod_Call) Run(run func(i *instances.Instance)) *MockInstancesController_GetInstancePeriod_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_GetInstancePeriod_Call) Return(_a0 *int64, _a1 error) *MockInstancesController_GetInstancePeriod_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_GetInstancePeriod_Call) RunAndReturn(run func(*instances.Instance) (*int64, error)) *MockInstancesController_GetInstancePeriod_Call {
	_c.Call.Return(run)
	return _c
}

// GetWithAccess provides a mock function with given fields: ctx, from, id
func (_m *MockInstancesController) GetWithAccess(ctx context.Context, from driver.DocumentID, id string) (graph.Instance, error) {
	ret := _m.Called(ctx, from, id)

	if len(ret) == 0 {
		panic("no return value specified for GetWithAccess")
	}

	var r0 graph.Instance
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, driver.DocumentID, string) (graph.Instance, error)); ok {
		return rf(ctx, from, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, driver.DocumentID, string) graph.Instance); ok {
		r0 = rf(ctx, from, id)
	} else {
		r0 = ret.Get(0).(graph.Instance)
	}

	if rf, ok := ret.Get(1).(func(context.Context, driver.DocumentID, string) error); ok {
		r1 = rf(ctx, from, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_GetWithAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetWithAccess'
type MockInstancesController_GetWithAccess_Call struct {
	*mock.Call
}

// GetWithAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - from driver.DocumentID
//   - id string
func (_e *MockInstancesController_Expecter) GetWithAccess(ctx interface{}, from interface{}, id interface{}) *MockInstancesController_GetWithAccess_Call {
	return &MockInstancesController_GetWithAccess_Call{Call: _e.mock.On("GetWithAccess", ctx, from, id)}
}

func (_c *MockInstancesController_GetWithAccess_Call) Run(run func(ctx context.Context, from driver.DocumentID, id string)) *MockInstancesController_GetWithAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(driver.DocumentID), args[2].(string))
	})
	return _c
}

func (_c *MockInstancesController_GetWithAccess_Call) Return(_a0 graph.Instance, _a1 error) *MockInstancesController_GetWithAccess_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_GetWithAccess_Call) RunAndReturn(run func(context.Context, driver.DocumentID, string) (graph.Instance, error)) *MockInstancesController_GetWithAccess_Call {
	_c.Call.Return(run)
	return _c
}

// SetState provides a mock function with given fields: ctx, inst, state
func (_m *MockInstancesController) SetState(ctx context.Context, inst *instances.Instance, state states.NoCloudState) error {
	ret := _m.Called(ctx, inst, state)

	if len(ret) == 0 {
		panic("no return value specified for SetState")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *instances.Instance, states.NoCloudState) error); ok {
		r0 = rf(ctx, inst, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_SetState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetState'
type MockInstancesController_SetState_Call struct {
	*mock.Call
}

// SetState is a helper method to define mock.On call
//   - ctx context.Context
//   - inst *instances.Instance
//   - state states.NoCloudState
func (_e *MockInstancesController_Expecter) SetState(ctx interface{}, inst interface{}, state interface{}) *MockInstancesController_SetState_Call {
	return &MockInstancesController_SetState_Call{Call: _e.mock.On("SetState", ctx, inst, state)}
}

func (_c *MockInstancesController_SetState_Call) Run(run func(ctx context.Context, inst *instances.Instance, state states.NoCloudState)) *MockInstancesController_SetState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*instances.Instance), args[2].(states.NoCloudState))
	})
	return _c
}

func (_c *MockInstancesController_SetState_Call) Return(err error) *MockInstancesController_SetState_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockInstancesController_SetState_Call) RunAndReturn(run func(context.Context, *instances.Instance, states.NoCloudState) error) *MockInstancesController_SetState_Call {
	_c.Call.Return(run)
	return _c
}

// SetStatus provides a mock function with given fields: ctx, inst, status
func (_m *MockInstancesController) SetStatus(ctx context.Context, inst *instances.Instance, status statuses.NoCloudStatus) error {
	ret := _m.Called(ctx, inst, status)

	if len(ret) == 0 {
		panic("no return value specified for SetStatus")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *instances.Instance, statuses.NoCloudStatus) error); ok {
		r0 = rf(ctx, inst, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_SetStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetStatus'
type MockInstancesController_SetStatus_Call struct {
	*mock.Call
}

// SetStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - inst *instances.Instance
//   - status statuses.NoCloudStatus
func (_e *MockInstancesController_Expecter) SetStatus(ctx interface{}, inst interface{}, status interface{}) *MockInstancesController_SetStatus_Call {
	return &MockInstancesController_SetStatus_Call{Call: _e.mock.On("SetStatus", ctx, inst, status)}
}

func (_c *MockInstancesController_SetStatus_Call) Run(run func(ctx context.Context, inst *instances.Instance, status statuses.NoCloudStatus)) *MockInstancesController_SetStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*instances.Instance), args[2].(statuses.NoCloudStatus))
	})
	return _c
}

func (_c *MockInstancesController_SetStatus_Call) Return(err error) *MockInstancesController_SetStatus_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockInstancesController_SetStatus_Call) RunAndReturn(run func(context.Context, *instances.Instance, statuses.NoCloudStatus) error) *MockInstancesController_SetStatus_Call {
	_c.Call.Return(run)
	return _c
}

// TransferInst provides a mock function with given fields: ctx, oldIGEdge, newIG, inst
func (_m *MockInstancesController) TransferInst(ctx context.Context, oldIGEdge string, newIG driver.DocumentID, inst driver.DocumentID) error {
	ret := _m.Called(ctx, oldIGEdge, newIG, inst)

	if len(ret) == 0 {
		panic("no return value specified for TransferInst")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, driver.DocumentID, driver.DocumentID) error); ok {
		r0 = rf(ctx, oldIGEdge, newIG, inst)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_TransferInst_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TransferInst'
type MockInstancesController_TransferInst_Call struct {
	*mock.Call
}

// TransferInst is a helper method to define mock.On call
//   - ctx context.Context
//   - oldIGEdge string
//   - newIG driver.DocumentID
//   - inst driver.DocumentID
func (_e *MockInstancesController_Expecter) TransferInst(ctx interface{}, oldIGEdge interface{}, newIG interface{}, inst interface{}) *MockInstancesController_TransferInst_Call {
	return &MockInstancesController_TransferInst_Call{Call: _e.mock.On("TransferInst", ctx, oldIGEdge, newIG, inst)}
}

func (_c *MockInstancesController_TransferInst_Call) Run(run func(ctx context.Context, oldIGEdge string, newIG driver.DocumentID, inst driver.DocumentID)) *MockInstancesController_TransferInst_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(driver.DocumentID), args[3].(driver.DocumentID))
	})
	return _c
}

func (_c *MockInstancesController_TransferInst_Call) Return(_a0 error) *MockInstancesController_TransferInst_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstancesController_TransferInst_Call) RunAndReturn(run func(context.Context, string, driver.DocumentID, driver.DocumentID) error) *MockInstancesController_TransferInst_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, sp, inst, oldInst
func (_m *MockInstancesController) Update(ctx context.Context, sp string, inst *instances.Instance, oldInst *instances.Instance) error {
	ret := _m.Called(ctx, sp, inst, oldInst)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *instances.Instance, *instances.Instance) error); ok {
		r0 = rf(ctx, sp, inst, oldInst)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockInstancesController_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - sp string
//   - inst *instances.Instance
//   - oldInst *instances.Instance
func (_e *MockInstancesController_Expecter) Update(ctx interface{}, sp interface{}, inst interface{}, oldInst interface{}) *MockInstancesController_Update_Call {
	return &MockInstancesController_Update_Call{Call: _e.mock.On("Update", ctx, sp, inst, oldInst)}
}

func (_c *MockInstancesController_Update_Call) Run(run func(ctx context.Context, sp string, inst *instances.Instance, oldInst *instances.Instance)) *MockInstancesController_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*instances.Instance), args[3].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_Update_Call) Return(_a0 error) *MockInstancesController_Update_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstancesController_Update_Call) RunAndReturn(run func(context.Context, string, *instances.Instance, *instances.Instance) error) *MockInstancesController_Update_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateNotes provides a mock function with given fields: ctx, inst
func (_m *MockInstancesController) UpdateNotes(ctx context.Context, inst *instances.Instance) error {
	ret := _m.Called(ctx, inst)

	if len(ret) == 0 {
		panic("no return value specified for UpdateNotes")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *instances.Instance) error); ok {
		r0 = rf(ctx, inst)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_UpdateNotes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateNotes'
type MockInstancesController_UpdateNotes_Call struct {
	*mock.Call
}

// UpdateNotes is a helper method to define mock.On call
//   - ctx context.Context
//   - inst *instances.Instance
func (_e *MockInstancesController_Expecter) UpdateNotes(ctx interface{}, inst interface{}) *MockInstancesController_UpdateNotes_Call {
	return &MockInstancesController_UpdateNotes_Call{Call: _e.mock.On("UpdateNotes", ctx, inst)}
}

func (_c *MockInstancesController_UpdateNotes_Call) Run(run func(ctx context.Context, inst *instances.Instance)) *MockInstancesController_UpdateNotes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_UpdateNotes_Call) Return(_a0 error) *MockInstancesController_UpdateNotes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstancesController_UpdateNotes_Call) RunAndReturn(run func(context.Context, *instances.Instance) error) *MockInstancesController_UpdateNotes_Call {
	_c.Call.Return(run)
	return _c
}

// ValidateBillingPlan provides a mock function with given fields: ctx, spUuid, i
func (_m *MockInstancesController) ValidateBillingPlan(ctx context.Context, spUuid string, i *instances.Instance) error {
	ret := _m.Called(ctx, spUuid, i)

	if len(ret) == 0 {
		panic("no return value specified for ValidateBillingPlan")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *instances.Instance) error); ok {
		r0 = rf(ctx, spUuid, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockInstancesController_ValidateBillingPlan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ValidateBillingPlan'
type MockInstancesController_ValidateBillingPlan_Call struct {
	*mock.Call
}

// ValidateBillingPlan is a helper method to define mock.On call
//   - ctx context.Context
//   - spUuid string
//   - i *instances.Instance
func (_e *MockInstancesController_Expecter) ValidateBillingPlan(ctx interface{}, spUuid interface{}, i interface{}) *MockInstancesController_ValidateBillingPlan_Call {
	return &MockInstancesController_ValidateBillingPlan_Call{Call: _e.mock.On("ValidateBillingPlan", ctx, spUuid, i)}
}

func (_c *MockInstancesController_ValidateBillingPlan_Call) Run(run func(ctx context.Context, spUuid string, i *instances.Instance)) *MockInstancesController_ValidateBillingPlan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*instances.Instance))
	})
	return _c
}

func (_c *MockInstancesController_ValidateBillingPlan_Call) Return(_a0 error) *MockInstancesController_ValidateBillingPlan_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockInstancesController_ValidateBillingPlan_Call) RunAndReturn(run func(context.Context, string, *instances.Instance) error) *MockInstancesController_ValidateBillingPlan_Call {
	_c.Call.Return(run)
	return _c
}

// getSp provides a mock function with given fields: ctx, uuid
func (_m *MockInstancesController) getSp(ctx context.Context, uuid string) (string, error) {
	ret := _m.Called(ctx, uuid)

	if len(ret) == 0 {
		panic("no return value specified for getSp")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, uuid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, uuid)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uuid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockInstancesController_getSp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'getSp'
type MockInstancesController_getSp_Call struct {
	*mock.Call
}

// getSp is a helper method to define mock.On call
//   - ctx context.Context
//   - uuid string
func (_e *MockInstancesController_Expecter) getSp(ctx interface{}, uuid interface{}) *MockInstancesController_getSp_Call {
	return &MockInstancesController_getSp_Call{Call: _e.mock.On("getSp", ctx, uuid)}
}

func (_c *MockInstancesController_getSp_Call) Run(run func(ctx context.Context, uuid string)) *MockInstancesController_getSp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockInstancesController_getSp_Call) Return(_a0 string, _a1 error) *MockInstancesController_getSp_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockInstancesController_getSp_Call) RunAndReturn(run func(context.Context, string) (string, error)) *MockInstancesController_getSp_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockInstancesController creates a new instance of MockInstancesController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockInstancesController(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockInstancesController {
	mock := &MockInstancesController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}