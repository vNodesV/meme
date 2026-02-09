package app

import (
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"

	wasmappparams "github.com/CosmWasm/wasmd/app/params"
	"github.com/CosmWasm/wasmd/x/wasm"
)

// MakeEncodingConfig creates a new EncodingConfig with all modules registered.
// This uses empty AppModuleBasic structs just for codec registration.
// The actual ModuleBasics with proper codecs is created from the module manager.
func MakeEncodingConfig() wasmappparams.EncodingConfig {
	encodingConfig := wasmappparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	
	// Create a temporary basic manager just for registering codecs
	// Empty structs are OK here since we're only calling RegisterLegacyAminoCodec and RegisterInterfaces
	// which don't need the codec fields
	tempBasicManager := module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		wasm.AppModuleBasic{},
	)
	
	tempBasicManager.RegisterLegacyAminoCodec(encodingConfig.Amino)
	tempBasicManager.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
