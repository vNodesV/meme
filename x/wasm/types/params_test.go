package types

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateParams(t *testing.T) {
	var (
		anyAddress     sdk.AccAddress = make([]byte, ContractAddrLen)
		invalidAddress                = "invalid address"
	)

	specs := map[string]struct {
		src    Params
		expErr bool
	}{
		"all good with defaults": {
			src: DefaultParams(),
		},
		"all good with nobody": {
			src: Params{
				CodeUploadAccess:             AllowNobody,
				InstantiateDefaultPermission: AccessTypeNobody,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
		},
		"all good with everybody": {
			src: Params{
				CodeUploadAccess:             AllowEverybody,
				InstantiateDefaultPermission: AccessTypeEverybody,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
		},
		"all good with only address": {
			src: Params{
				CodeUploadAccess:             AccessTypeOnlyAddress.With(anyAddress),
				InstantiateDefaultPermission: AccessTypeOnlyAddress,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
		},
		"reject empty type in instantiate permission": {
			src: Params{
				CodeUploadAccess: AllowNobody,
				MaxWasmCodeSize:  DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject unknown type in instantiate": {
			src: Params{
				CodeUploadAccess:             AllowNobody,
				InstantiateDefaultPermission: 1111,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject invalid address in only address": {
			src: Params{
				CodeUploadAccess:             AccessConfig{Permission: AccessTypeOnlyAddress, Address: invalidAddress},
				InstantiateDefaultPermission: AccessTypeOnlyAddress,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject CodeUploadAccess Everybody with obsolete address": {
			src: Params{
				CodeUploadAccess:             AccessConfig{Permission: AccessTypeEverybody, Address: anyAddress.String()},
				InstantiateDefaultPermission: AccessTypeOnlyAddress,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject CodeUploadAccess Nobody with obsolete address": {
			src: Params{
				CodeUploadAccess:             AccessConfig{Permission: AccessTypeNobody, Address: anyAddress.String()},
				InstantiateDefaultPermission: AccessTypeOnlyAddress,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject empty CodeUploadAccess": {
			src: Params{
				InstantiateDefaultPermission: AccessTypeOnlyAddress,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject undefined permission in CodeUploadAccess": {
			src: Params{
				CodeUploadAccess:             AccessConfig{Permission: AccessTypeUnspecified},
				InstantiateDefaultPermission: AccessTypeOnlyAddress,
				MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
			},
			expErr: true,
		},
		"reject empty max wasm code size": {
			src: Params{
				CodeUploadAccess:             AllowNobody,
				InstantiateDefaultPermission: AccessTypeNobody,
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccessTypeMarshalJson(t *testing.T) {
	specs := map[string]struct {
		src AccessType
		exp string
	}{
		// MarshalJSON now emits proto-registered names for jsonpb round-trip compatibility
		"Unspecified": {src: AccessTypeUnspecified, exp: `"ACCESS_TYPE_UNSPECIFIED"`},
		"Nobody":      {src: AccessTypeNobody, exp: `"ACCESS_TYPE_NOBODY"`},
		"OnlyAddress": {src: AccessTypeOnlyAddress, exp: `"ACCESS_TYPE_ONLY_ADDRESS"`},
		"Everybody":   {src: AccessTypeEverybody, exp: `"ACCESS_TYPE_EVERYBODY"`},
		"unknown":     {src: 999, exp: `999`}, // unknown values marshal as integer
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got, err := json.Marshal(spec.src)
			require.NoError(t, err)
			assert.Equal(t, []byte(spec.exp), got)
		})
	}
}

func TestAccessTypeUnmarshalJson(t *testing.T) {
	specs := map[string]struct {
		src string
		exp AccessType
	}{
		"Unspecified": {src: `"Unspecified"`, exp: AccessTypeUnspecified},
		"Nobody":      {src: `"Nobody"`, exp: AccessTypeNobody},
		"OnlyAddress": {src: `"OnlyAddress"`, exp: AccessTypeOnlyAddress},
		"Everybody":   {src: `"Everybody"`, exp: AccessTypeEverybody},
		"unknown":     {src: `""`, exp: AccessTypeUnspecified},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			var got AccessType
			err := json.Unmarshal([]byte(spec.src), &got)
			require.NoError(t, err)
			assert.Equal(t, spec.exp, got)
		})
	}
}
func TestParamsUnmarshalJson(t *testing.T) {
	specs := map[string]struct {
		src string
		exp Params
	}{

		"defaults": {
			src: `{"code_upload_access": {"permission": "ACCESS_TYPE_EVERYBODY"},
				"instantiate_default_permission": "ACCESS_TYPE_EVERYBODY",
				"max_wasm_code_size": 1228800}`,
			exp: DefaultParams(),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			var val Params
			interfaceRegistry := codectypes.NewInterfaceRegistry()
			marshaler := codec.NewProtoCodec(interfaceRegistry)

			err := marshaler.UnmarshalJSON([]byte(spec.src), &val)
			require.NoError(t, err)
			assert.Equal(t, spec.exp, val)
		})
	}
}
