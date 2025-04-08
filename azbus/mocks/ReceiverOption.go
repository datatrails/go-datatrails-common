// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	azbus "github.com/datatrails/go-datatrails-common/azbus"
	mock "github.com/stretchr/testify/mock"
)

// ReceiverOption is an autogenerated mock type for the ReceiverOption type
type ReceiverOption struct {
	mock.Mock
}

type ReceiverOption_Expecter struct {
	mock *mock.Mock
}

func (_m *ReceiverOption) EXPECT() *ReceiverOption_Expecter {
	return &ReceiverOption_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *ReceiverOption) Execute(_a0 *azbus.Receiver) {
	_m.Called(_a0)
}

// ReceiverOption_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type ReceiverOption_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 *azbus.Receiver
func (_e *ReceiverOption_Expecter) Execute(_a0 interface{}) *ReceiverOption_Execute_Call {
	return &ReceiverOption_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *ReceiverOption_Execute_Call) Run(run func(_a0 *azbus.Receiver)) *ReceiverOption_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*azbus.Receiver))
	})
	return _c
}

func (_c *ReceiverOption_Execute_Call) Return() *ReceiverOption_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *ReceiverOption_Execute_Call) RunAndReturn(run func(*azbus.Receiver)) *ReceiverOption_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewReceiverOption creates a new instance of ReceiverOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReceiverOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReceiverOption {
	mock := &ReceiverOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
