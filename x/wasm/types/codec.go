package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/gogoproto/proto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the account types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) { //nolint:staticcheck
	cdc.RegisterConcrete(&MsgStoreCode{}, "wasm/MsgStoreCode", nil)
	cdc.RegisterConcrete(&MsgInstantiateContract{}, "wasm/MsgInstantiateContract", nil)
	cdc.RegisterConcrete(&MsgExecuteContract{}, "wasm/MsgExecuteContract", nil)
	cdc.RegisterConcrete(&MsgMigrateContract{}, "wasm/MsgMigrateContract", nil)
	cdc.RegisterConcrete(&MsgUpdateAdmin{}, "wasm/MsgUpdateAdmin", nil)
	cdc.RegisterConcrete(&MsgClearAdmin{}, "wasm/MsgClearAdmin", nil)
	cdc.RegisterConcrete(&PinCodesProposal{}, "wasm/PinCodesProposal", nil)
	cdc.RegisterConcrete(&UnpinCodesProposal{}, "wasm/UnpinCodesProposal", nil)

	cdc.RegisterConcrete(&StoreCodeProposal{}, "wasm/StoreCodeProposal", nil)
	cdc.RegisterConcrete(&InstantiateContractProposal{}, "wasm/InstantiateContractProposal", nil)
	cdc.RegisterConcrete(&MigrateContractProposal{}, "wasm/MigrateContractProposal", nil)
	cdc.RegisterConcrete(&UpdateAdminProposal{}, "wasm/UpdateAdminProposal", nil)
	cdc.RegisterConcrete(&ClearAdminProposal{}, "wasm/ClearAdminProposal", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface("ContractInfoExtension", (*ContractInfoExtension)(nil))

	// SDK 0.50: Register message implementations with explicit type URLs
	// Use a type assertion to access the concrete registry's RegisterCustomTypeURL method
	// This is needed because gogoproto-generated messages don't properly populate proto.MessageName()
	type customRegistry interface {
		RegisterCustomTypeURL(iface interface{}, typeURL string, impl proto.Message)
	}
	
	cr, ok := registry.(customRegistry)
	if !ok {
		// Type assertion failed - fall back to panic-inducing behavior
		// to help debug the issue
		panic("InterfaceRegistry does not implement RegisterCustomTypeURL - cannot register wasm messages")
	}
	
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgStoreCode", &MsgStoreCode{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgInstantiateContract", &MsgInstantiateContract{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgExecuteContract", &MsgExecuteContract{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgMigrateContract", &MsgMigrateContract{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgUpdateAdmin", &MsgUpdateAdmin{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgClearAdmin", &MsgClearAdmin{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgIBCCloseChannel", &MsgIBCCloseChannel{})
	cr.RegisterCustomTypeURL((*sdk.Msg)(nil), "/cosmwasm.wasm.v1.MsgIBCSend", &MsgIBCSend{})
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/wasm module codec.

	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
