package params

import (
	"cosmossdk.io/x/tx/signing"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
)

const (
	// Bech32 address prefixes for the meme chain
	Bech32PrefixAccAddr  = "meme"
	Bech32PrefixAccPub   = Bech32PrefixAccAddr + "pub"
	Bech32PrefixValAddr  = Bech32PrefixAccAddr + "valoper"
	Bech32PrefixValPub   = Bech32PrefixAccAddr + "valoperpub"
	Bech32PrefixConsAddr = Bech32PrefixAccAddr + "valcons"
	Bech32PrefixConsPub  = Bech32PrefixAccAddr + "valconspub"
)

// MakeEncodingConfig creates an EncodingConfig with proper address codecs for SDK 0.50.
// The InterfaceRegistry is initialized with address codecs to enable transaction signing.
// NOTE: Bech32 prefixes must be set on sdk.Config BEFORE calling this function.
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()

	// Create address codecs using hardcoded prefixes that match app.go constants
	accCodec := addresscodec.NewBech32Codec(Bech32PrefixAccAddr)
	valCodec := addresscodec.NewBech32Codec(Bech32PrefixValAddr)

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
