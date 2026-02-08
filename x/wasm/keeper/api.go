package keeper

import (
	wasmvm "github.com/CosmWasm/wasmvm/v2"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultGasCostHumanAddress is how much SDK gas we charge to convert to a human address format
	DefaultGasCostHumanAddress = 5
	// DefaultGasCostCanonicalAddress is how much SDK gas we charge to convert to a canonical address format
	DefaultGasCostCanonicalAddress = 4
	// DefaultGasCostValidateAddress is how much SDK gas we charge to validate an address
	DefaultGasCostValidateAddress = DefaultGasCostHumanAddress + DefaultGasCostCanonicalAddress

	// DefaultDeserializationCostPerByte The formular should be `len(data) * deserializationCostPerByte`
	DefaultDeserializationCostPerByte = 1
)

var (
	costHumanize            = DefaultGasCostHumanAddress * DefaultGasMultiplier
	costCanonical           = DefaultGasCostCanonicalAddress * DefaultGasMultiplier
	costValidate            = DefaultGasCostValidateAddress * DefaultGasMultiplier
	costJSONDeserialization = wasmvmtypes.UFraction{
		Numerator:   DefaultDeserializationCostPerByte * DefaultGasMultiplier,
		Denominator: 1,
	}
)

func humanizeAddress(canon []byte) (string, uint64, error) {
	if err := sdk.VerifyAddressFormat(canon); err != nil {
		return "", costHumanize, err
	}
	return sdk.AccAddress(canon).String(), costHumanize, nil
}

func canonicalizeAddress(human string) ([]byte, uint64, error) {
	bz, err := sdk.AccAddressFromBech32(human)
	return bz, costCanonical, err
}

func validateAddress(human string) (uint64, error) {
	canonicalized, err := sdk.AccAddressFromBech32(human)
	if err != nil {
		return costValidate, err
	}
	// AccAddressFromBech32 already calls VerifyAddressFormat, so we can just humanize and compare
	if canonicalized.String() != human {
		return costValidate, wasmvmtypes.InvalidRequest{Err: "address not normalized"}
	}
	return costValidate, nil
}

var cosmwasmAPI = wasmvm.GoAPI{
	HumanizeAddress:     humanizeAddress,
	CanonicalizeAddress: canonicalizeAddress,
	ValidateAddress:     validateAddress,
}
