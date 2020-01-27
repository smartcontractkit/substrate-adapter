package main

import (
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/stretchr/testify/require"
	"math/big"
	"reflect"
	"testing"
)

type Argument struct {
	Type  string
	Value string
}

func Test_convertTypes(t *testing.T) {
	addr1, err := types.NewAddressFromHexAccountID("0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d")
	require.NoError(t, err)
	addr2, err := types.NewAddressFromHexAccountID("0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48")
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []Argument
		want    []interface{}
		wantErr bool
	}{
		{
			"converts bool",
			[]Argument{{
				Type:  "bool",
				Value: "false",
			}, {
				Type:  "bool",
				Value: "true",
			}},
			[]interface{}{
				types.Bool(false),
				types.Bool(true),
			},
			false,
		},
		{
			"fails on invalid bool type",
			[]Argument{{
				Type:  "bool",
				Value: "123",
			}},
			[]interface{}{},
			true,
		},
		{
			"converts uint256",
			[]Argument{{
				Type:  "uint256",
				Value: "1234567890",
			}, {
				Type:  "uint256",
				Value: "99999999999999",
			}},
			[]interface{}{
				types.NewU256(*big.NewInt(1234567890)),
				types.NewU256(*big.NewInt(99999999999999)),
			},
			false,
		},
		{
			"fails on invalid uint256",
			[]Argument{{
				Type:  "uint256",
				Value: "abcdef",
			}},
			[]interface{}{},
			true,
		},
		{
			"converts int256",
			[]Argument{{
				Type:  "int256",
				Value: "1234567890",
			}, {
				Type:  "int256",
				Value: "-99999999999999",
			}},
			[]interface{}{
				types.NewI256(*big.NewInt(1234567890)),
				types.NewI256(*big.NewInt(-99999999999999)),
			},
			false,
		},
		{
			"fails on invalid int256",
			[]Argument{{
				Type:  "uint256",
				Value: "abcdef",
			}},
			[]interface{}{},
			true,
		},
		{
			"converts ucompact",
			[]Argument{{
				Type:  "ucompact",
				Value: "1234567890",
			}, {
				Type:  "ucompact",
				Value: "99999999999999",
			}},
			[]interface{}{
				types.UCompact(1234567890),
				types.UCompact(99999999999999),
			},
			false,
		},
		{
			"fails on invalid ucompact",
			[]Argument{{
				Type:  "ucompact",
				Value: "abcdef",
			}},
			[]interface{}{},
			true,
		},
		{
			"converts bytes",
			[]Argument{{
				Type:  "bytes",
				Value: "some string",
			}, {
				Type:  "bytes",
				Value: "{\"key\":\"value\"}",
			}},
			[]interface{}{
				types.Bytes("some string"),
				types.Bytes(`{"key":"value"}`),
			},
			false,
		},
		{
			"converts addresses",
			[]Argument{{
				Type:  "address",
				Value: "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d",
			}, {
				Type:  "address",
				Value: "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48",
			}},
			[]interface{}{
				addr1,
				addr2,
			},
			false,
		},
		{
			"fails on invalid addresses",
			[]Argument{{
				Type:  "address",
				Value: "not an address",
			}},
			[]interface{}{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, arg := range tt.args {
				got, err := convertTypes(arg.Type, arg.Value)
				if (err != nil) != tt.wantErr {
					t.Errorf("convertTypes() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err == nil && !reflect.DeepEqual(got, tt.want[i]) {
					t.Errorf("convertTypes() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
