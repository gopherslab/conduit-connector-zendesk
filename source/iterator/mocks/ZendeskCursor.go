// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

// ZendeskCursor is an autogenerated mock type for the ZendeskCursor type
type ZendeskCursor struct {
	mock.Mock
}

// FetchRecords provides a mock function with given fields: ctx
func (_m *ZendeskCursor) FetchRecords(ctx context.Context) ([]sdk.Record, error) {
	ret := _m.Called(ctx)

	var r0 []sdk.Record
	if rf, ok := ret.Get(0).(func(context.Context) []sdk.Record); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sdk.Record)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
