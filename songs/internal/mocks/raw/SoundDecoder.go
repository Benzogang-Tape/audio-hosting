// Code generated by mockery v2.48.0. DO NOT EDIT.

package rawmocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// SoundDecoder is an autogenerated mock type for the SoundDecoder type
type SoundDecoder struct {
	mock.Mock
}

type SoundDecoder_Expecter struct {
	mock *mock.Mock
}

func (_m *SoundDecoder) EXPECT() *SoundDecoder_Expecter {
	return &SoundDecoder_Expecter{mock: &_m.Mock}
}

// GetMp3Duration provides a mock function with given fields: _a0, _a1
func (_m *SoundDecoder) GetMp3Duration(_a0 context.Context, _a1 io.Reader) (time.Duration, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetMp3Duration")
	}

	var r0 time.Duration
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) (time.Duration, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) time.Duration); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	if rf, ok := ret.Get(1).(func(context.Context, io.Reader) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SoundDecoder_GetMp3Duration_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMp3Duration'
type SoundDecoder_GetMp3Duration_Call struct {
	*mock.Call
}

// GetMp3Duration is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 io.Reader
func (_e *SoundDecoder_Expecter) GetMp3Duration(_a0 interface{}, _a1 interface{}) *SoundDecoder_GetMp3Duration_Call {
	return &SoundDecoder_GetMp3Duration_Call{Call: _e.mock.On("GetMp3Duration", _a0, _a1)}
}

func (_c *SoundDecoder_GetMp3Duration_Call) Run(run func(_a0 context.Context, _a1 io.Reader)) *SoundDecoder_GetMp3Duration_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(io.Reader))
	})
	return _c
}

func (_c *SoundDecoder_GetMp3Duration_Call) Return(_a0 time.Duration, _a1 error) *SoundDecoder_GetMp3Duration_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SoundDecoder_GetMp3Duration_Call) RunAndReturn(run func(context.Context, io.Reader) (time.Duration, error)) *SoundDecoder_GetMp3Duration_Call {
	_c.Call.Return(run)
	return _c
}

// NewSoundDecoder creates a new instance of SoundDecoder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSoundDecoder(t interface {
	mock.TestingT
	Cleanup(func())
}) *SoundDecoder {
	mock := &SoundDecoder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
