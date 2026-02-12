package params

import (
	"cosmossdk.io/x/tx/signing"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
)

// MakeEncodingConfig creates an EncodingConfig with proper address codecs for SDK 0.50.
// The InterfaceRegistry is initialized with address codecs to enable transaction signing.
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	
	// Get SDK config for Bech32 prefixes
	cfg := sdk.GetConfig()
	
	// Create address codecs
	accCodec := addresscodec.NewBech32Codec(cfg.GetBech32AccountAddrPrefix())
	valCodec := addresscodec.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix())
	
	// Create InterfaceRegistry with proper address codecs for transaction signing
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          accCodec,
			ValidatorAddressCodec: valCodec,
		},
	})
	if err != nil {
		panic(err)
	}
	
	// Register ed25519 public key
	interfaceRegistry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ed25519.PubKey{})
	
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
