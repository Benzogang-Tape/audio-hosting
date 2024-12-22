// Code generated by mockery v2.48.0. DO NOT EDIT.

package rawmocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	postgres "github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"

	raw "github.com/Benzogang-Tape/audio-hosting/songs/internal/services/raw"
)

// SongRepo is an autogenerated mock type for the SongRepo type
type SongRepo struct {
	mock.Mock
}

type SongRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *SongRepo) EXPECT() *SongRepo_Expecter {
	return &SongRepo_Expecter{mock: &_m.Mock}
}

// Begin provides a mock function with given fields: _a0
func (_m *SongRepo) Begin(_a0 context.Context) (raw.SongRepo, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Begin")
	}

	var r0 raw.SongRepo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (raw.SongRepo, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) raw.SongRepo); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(raw.SongRepo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SongRepo_Begin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Begin'
type SongRepo_Begin_Call struct {
	*mock.Call
}

// Begin is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *SongRepo_Expecter) Begin(_a0 interface{}) *SongRepo_Begin_Call {
	return &SongRepo_Begin_Call{Call: _e.mock.On("Begin", _a0)}
}

func (_c *SongRepo_Begin_Call) Run(run func(_a0 context.Context)) *SongRepo_Begin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *SongRepo_Begin_Call) Return(_a0 raw.SongRepo, _a1 error) *SongRepo_Begin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SongRepo_Begin_Call) RunAndReturn(run func(context.Context) (raw.SongRepo, error)) *SongRepo_Begin_Call {
	_c.Call.Return(run)
	return _c
}

// Commit provides a mock function with given fields: _a0
func (_m *SongRepo) Commit(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SongRepo_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type SongRepo_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *SongRepo_Expecter) Commit(_a0 interface{}) *SongRepo_Commit_Call {
	return &SongRepo_Commit_Call{Call: _e.mock.On("Commit", _a0)}
}

func (_c *SongRepo_Commit_Call) Run(run func(_a0 context.Context)) *SongRepo_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *SongRepo_Commit_Call) Return(_a0 error) *SongRepo_Commit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SongRepo_Commit_Call) RunAndReturn(run func(context.Context) error) *SongRepo_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// MySong provides a mock function with given fields: _a0, _a1
func (_m *SongRepo) MySong(_a0 context.Context, _a1 postgres.MySongParams) (postgres.MySongRow, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for MySong")
	}

	var r0 postgres.MySongRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, postgres.MySongParams) (postgres.MySongRow, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, postgres.MySongParams) postgres.MySongRow); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(postgres.MySongRow)
	}

	if rf, ok := ret.Get(1).(func(context.Context, postgres.MySongParams) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SongRepo_MySong_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MySong'
type SongRepo_MySong_Call struct {
	*mock.Call
}

// MySong is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 postgres.MySongParams
func (_e *SongRepo_Expecter) MySong(_a0 interface{}, _a1 interface{}) *SongRepo_MySong_Call {
	return &SongRepo_MySong_Call{Call: _e.mock.On("MySong", _a0, _a1)}
}

func (_c *SongRepo_MySong_Call) Run(run func(_a0 context.Context, _a1 postgres.MySongParams)) *SongRepo_MySong_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(postgres.MySongParams))
	})
	return _c
}

func (_c *SongRepo_MySong_Call) Return(_a0 postgres.MySongRow, _a1 error) *SongRepo_MySong_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SongRepo_MySong_Call) RunAndReturn(run func(context.Context, postgres.MySongParams) (postgres.MySongRow, error)) *SongRepo_MySong_Call {
	_c.Call.Return(run)
	return _c
}

// PatchSong provides a mock function with given fields: _a0, _a1
func (_m *SongRepo) PatchSong(_a0 context.Context, _a1 postgres.PatchSongParams) (postgres.Song, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for PatchSong")
	}

	var r0 postgres.Song
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, postgres.PatchSongParams) (postgres.Song, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, postgres.PatchSongParams) postgres.Song); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(postgres.Song)
	}

	if rf, ok := ret.Get(1).(func(context.Context, postgres.PatchSongParams) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SongRepo_PatchSong_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PatchSong'
type SongRepo_PatchSong_Call struct {
	*mock.Call
}

// PatchSong is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 postgres.PatchSongParams
func (_e *SongRepo_Expecter) PatchSong(_a0 interface{}, _a1 interface{}) *SongRepo_PatchSong_Call {
	return &SongRepo_PatchSong_Call{Call: _e.mock.On("PatchSong", _a0, _a1)}
}

func (_c *SongRepo_PatchSong_Call) Run(run func(_a0 context.Context, _a1 postgres.PatchSongParams)) *SongRepo_PatchSong_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(postgres.PatchSongParams))
	})
	return _c
}

func (_c *SongRepo_PatchSong_Call) Return(_a0 postgres.Song, _a1 error) *SongRepo_PatchSong_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SongRepo_PatchSong_Call) RunAndReturn(run func(context.Context, postgres.PatchSongParams) (postgres.Song, error)) *SongRepo_PatchSong_Call {
	_c.Call.Return(run)
	return _c
}

// Rollback provides a mock function with given fields: _a0
func (_m *SongRepo) Rollback(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SongRepo_Rollback_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rollback'
type SongRepo_Rollback_Call struct {
	*mock.Call
}

// Rollback is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *SongRepo_Expecter) Rollback(_a0 interface{}) *SongRepo_Rollback_Call {
	return &SongRepo_Rollback_Call{Call: _e.mock.On("Rollback", _a0)}
}

func (_c *SongRepo_Rollback_Call) Run(run func(_a0 context.Context)) *SongRepo_Rollback_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *SongRepo_Rollback_Call) Return(_a0 error) *SongRepo_Rollback_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SongRepo_Rollback_Call) RunAndReturn(run func(context.Context) error) *SongRepo_Rollback_Call {
	_c.Call.Return(run)
	return _c
}

// NewSongRepo creates a new instance of SongRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSongRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *SongRepo {
	mock := &SongRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}