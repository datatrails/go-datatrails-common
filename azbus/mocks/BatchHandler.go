// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	azservicebus "github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	azbus "github.com/datatrails/go-datatrails-common/azbus"

	context "context"

	mock "github.com/stretchr/testify/mock"
)

// BatchHandler is an autogenerated mock type for the BatchHandler type
type BatchHandler struct {
	mock.Mock
}

type BatchHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *BatchHandler) EXPECT() *BatchHandler_Expecter {
	return &BatchHandler_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *BatchHandler) Close() {
	_m.Called()
}

// BatchHandler_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type BatchHandler_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *BatchHandler_Expecter) Close() *BatchHandler_Close_Call {
	return &BatchHandler_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *BatchHandler_Close_Call) Run(run func()) *BatchHandler_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BatchHandler_Close_Call) Return() *BatchHandler_Close_Call {
	_c.Call.Return()
	return _c
}

func (_c *BatchHandler_Close_Call) RunAndReturn(run func()) *BatchHandler_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Handle provides a mock function with given fields: _a0, _a1, _a2
func (_m *BatchHandler) Handle(_a0 context.Context, _a1 azbus.Disposer, _a2 []*azservicebus.ReceivedMessage) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Handle")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, azbus.Disposer, []*azservicebus.ReceivedMessage) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BatchHandler_Handle_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Handle'
type BatchHandler_Handle_Call struct {
	*mock.Call
}

// Handle is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 azbus.Disposer
//   - _a2 []*azservicebus.ReceivedMessage
func (_e *BatchHandler_Expecter) Handle(_a0 interface{}, _a1 interface{}, _a2 interface{}) *BatchHandler_Handle_Call {
	return &BatchHandler_Handle_Call{Call: _e.mock.On("Handle", _a0, _a1, _a2)}
}

func (_c *BatchHandler_Handle_Call) Run(run func(_a0 context.Context, _a1 azbus.Disposer, _a2 []*azservicebus.ReceivedMessage)) *BatchHandler_Handle_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(azbus.Disposer), args[2].([]*azservicebus.ReceivedMessage))
	})
	return _c
}

func (_c *BatchHandler_Handle_Call) Return(_a0 error) *BatchHandler_Handle_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BatchHandler_Handle_Call) RunAndReturn(run func(context.Context, azbus.Disposer, []*azservicebus.ReceivedMessage) error) *BatchHandler_Handle_Call {
	_c.Call.Return(run)
	return _c
}

// Open provides a mock function with given fields:
func (_m *BatchHandler) Open() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Open")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BatchHandler_Open_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Open'
type BatchHandler_Open_Call struct {
	*mock.Call
}

// Open is a helper method to define mock.On call
func (_e *BatchHandler_Expecter) Open() *BatchHandler_Open_Call {
	return &BatchHandler_Open_Call{Call: _e.mock.On("Open")}
}

func (_c *BatchHandler_Open_Call) Run(run func()) *BatchHandler_Open_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BatchHandler_Open_Call) Return(_a0 error) *BatchHandler_Open_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BatchHandler_Open_Call) RunAndReturn(run func() error) *BatchHandler_Open_Call {
	_c.Call.Return(run)
	return _c
}

// NewBatchHandler creates a new instance of BatchHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBatchHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *BatchHandler {
	mock := &BatchHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
