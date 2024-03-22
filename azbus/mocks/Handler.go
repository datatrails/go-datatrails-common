// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	azbus "github.com/datatrails/go-datatrails-common/azbus"

	mock "github.com/stretchr/testify/mock"
)

// Handler is an autogenerated mock type for the Handler type
type Handler struct {
	mock.Mock
}

type Handler_Expecter struct {
	mock *mock.Mock
}

func (_m *Handler) EXPECT() *Handler_Expecter {
	return &Handler_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *Handler) Close() {
	_m.Called()
}

// Handler_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Handler_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Handler_Expecter) Close() *Handler_Close_Call {
	return &Handler_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Handler_Close_Call) Run(run func()) *Handler_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Handler_Close_Call) Return() *Handler_Close_Call {
	_c.Call.Return()
	return _c
}

func (_c *Handler_Close_Call) RunAndReturn(run func()) *Handler_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Handle provides a mock function with given fields: _a0, _a1
func (_m *Handler) Handle(_a0 context.Context, _a1 *azbus.ReceivedMessage) (azbus.Disposition, context.Context, error) {
	ret := _m.Called(_a0, _a1)

	var r0 azbus.Disposition
	var r1 context.Context
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *azbus.ReceivedMessage) (azbus.Disposition, context.Context, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *azbus.ReceivedMessage) azbus.Disposition); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(azbus.Disposition)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *azbus.ReceivedMessage) context.Context); ok {
		r1 = rf(_a0, _a1)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(context.Context)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, *azbus.ReceivedMessage) error); ok {
		r2 = rf(_a0, _a1)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Handler_Handle_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Handle'
type Handler_Handle_Call struct {
	*mock.Call
}

// Handle is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *azbus.ReceivedMessage
func (_e *Handler_Expecter) Handle(_a0 interface{}, _a1 interface{}) *Handler_Handle_Call {
	return &Handler_Handle_Call{Call: _e.mock.On("Handle", _a0, _a1)}
}

func (_c *Handler_Handle_Call) Run(run func(_a0 context.Context, _a1 *azbus.ReceivedMessage)) *Handler_Handle_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*azbus.ReceivedMessage))
	})
	return _c
}

func (_c *Handler_Handle_Call) Return(_a0 azbus.Disposition, _a1 context.Context, _a2 error) *Handler_Handle_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Handler_Handle_Call) RunAndReturn(run func(context.Context, *azbus.ReceivedMessage) (azbus.Disposition, context.Context, error)) *Handler_Handle_Call {
	_c.Call.Return(run)
	return _c
}

// Open provides a mock function with given fields:
func (_m *Handler) Open() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Handler_Open_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Open'
type Handler_Open_Call struct {
	*mock.Call
}

// Open is a helper method to define mock.On call
func (_e *Handler_Expecter) Open() *Handler_Open_Call {
	return &Handler_Open_Call{Call: _e.mock.On("Open")}
}

func (_c *Handler_Open_Call) Run(run func()) *Handler_Open_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Handler_Open_Call) Return(_a0 error) *Handler_Open_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Handler_Open_Call) RunAndReturn(run func() error) *Handler_Open_Call {
	_c.Call.Return(run)
	return _c
}

// NewHandler creates a new instance of Handler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *Handler {
	mock := &Handler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
