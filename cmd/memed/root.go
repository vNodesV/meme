package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	feegrantcli "cosmossdk.io/x/feegrant/client/cli"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradecli "cosmossdk.io/x/upgrade/client/cli"
	cmtcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzcli "github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkparams "github.com/cosmos/cosmos-sdk/x/params"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfercli "github.com/cosmos/ibc-go/v8/modules/apps/transfer/client/cli"
	ibccli "github.com/cosmos/ibc-go/v8/modules/core/client/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/CosmWasm/wasmd/app"
	"github.com/CosmWasm/wasmd/app/params"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// NewRootCmd creates a new root command for wasmd. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig()

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())
	cfg.Seal()

	accountAddressCodec := addresscodec.NewBech32Codec(cfg.GetBech32AccountAddrPrefix())
	validatorAddressCodec := addresscodec.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix())
	consensusAddressCodec := addresscodec.NewBech32Codec(cfg.GetBech32ConsensusAddrPrefix())

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("").
		WithAddressCodec(accountAddressCodec).
		WithValidatorAddressCodec(validatorAddressCodec).
		WithConsensusAddressCodec(consensusAddressCodec)

	rootCmd := &cobra.Command{
		Use:   version.AppName,
		Short: "MeMe Chain Daemon (server)",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			cmtConfig := cmtcfg.DefaultConfig()
			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, cmtConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	// Enhance root command with AutoCLI-generated query and tx commands
	// for modules that use AutoCLI (auth, bank, staking, distribution, gov,
	// slashing, mint, params, feegrant, authz, evidence, upgrade).
	// This only adds missing commands and does not override manually registered ones.
	autoCliOpts := initAutoCliOptions(initClientCtx)
	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	// Create address codecs
	cfg := sdk.GetConfig()
	accountAddressCodec := addresscodec.NewBech32Codec(cfg.GetBech32AccountAddrPrefix())
	validatorAddressCodec := addresscodec.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix())

	// Get basic manager for CLI commands
	// Note: This uses empty structs for modules, which is OK for InitCmd, ValidateGenesisCmd, etc.
	// For commands that need codecs (like AddTxCommands), they are handled separately
	basicManager := app.MakeBasicManager()

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, genutiltypes.DefaultMessageValidator, validatorAddressCodec),
		genutilcli.GenTxCmd(basicManager, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, accountAddressCodec),
		genutilcli.ValidateGenesisCmd(basicManager),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		AddGenesisWasmMsgCmd(app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		// testnetCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
	)

	// Create app creator and exporter functions
	appCreatorFunc := makeAppCreator(encodingConfig)
	appExporterFunc := makeAppExporter(encodingConfig)

	server.AddCommands(rootCmd, app.DefaultNodeHome, appCreatorFunc, appExporterFunc, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(basicManager),
		txCommand(basicManager),
		keys.Commands(),
	)
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	wasm.AddModuleInitFlags(startCmd)
}

// initAppConfig returns custom app config template and config
func initAppConfig() (string, interface{}) {
	// Use default config from server package (must return pointer for Viper)
	srvCfg := serverconfig.DefaultConfig()

	// Customize config if needed (for now using defaults)

	return "", srvCfg
}

// initAutoCliOptions builds AutoCLI AppOptions by extracting AutoCLI module
// options from zero-value AppModule instances. The AutoCLIOptions() method on
// each module only returns static proto service descriptors and does not
// require keepers or state. This enables AutoCLI-generated query/tx commands
// for all core SDK modules.
//
// Core SDK modules are provided via ModuleOptions (not Modules) to avoid
// panics from GetTxCmd()/GetQueryCmd() on zero-value structs with nil codecs.
// Wasm is provided via Modules so its custom GetQueryCmd() is detected.
func initAutoCliOptions(clientCtx client.Context) autocli.AppOptions {
	cfg := sdk.GetConfig()

	// Extract AutoCLI module options from zero-value AppModule instances.
	// AutoCLIOptions() only returns static protobuf service descriptors and
	// is safe to call on zero-value structs.
	moduleOptions := map[string]*autocliv1.ModuleOptions{
		authtypes.ModuleName:     auth.AppModule{}.AutoCLIOptions(),
		banktypes.ModuleName:     bank.AppModule{}.AutoCLIOptions(),
		stakingtypes.ModuleName:  staking.AppModule{}.AutoCLIOptions(),
		distrtypes.ModuleName:    distr.AppModule{}.AutoCLIOptions(),
		govtypes.ModuleName:      gov.AppModule{}.AutoCLIOptions(),
		slashingtypes.ModuleName: slashing.AppModule{}.AutoCLIOptions(),
		minttypes.ModuleName:     mint.AppModule{}.AutoCLIOptions(),
		paramstypes.ModuleName:   sdkparams.AppModule{}.AutoCLIOptions(),
		"feegrant":               feegrantmodule.AppModule{}.AutoCLIOptions(),
		authz.ModuleName:         authzmodule.AppModule{}.AutoCLIOptions(),
		"evidence":               evidence.AppModule{}.AutoCLIOptions(),
		"upgrade":                upgrade.AppModule{}.AutoCLIOptions(),
	}

	// Wasm is included via Modules so its custom GetQueryCmd()/GetTxCmd()
	// are detected by AutoCLI's HasCustomQueryCommand/HasCustomTxCommand.
	modules := map[string]appmodule.AppModule{
		"wasm": wasm.AppModule{},
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         moduleOptions,
		AddressCodec:          addresscodec.NewBech32Codec(cfg.GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: addresscodec.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: addresscodec.NewBech32Codec(cfg.GetBech32ConsensusAddrPrefix()),
		ClientCtx:             clientCtx,
	}
}

func queryCommand(basicManager module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	// AddQueryCommands registers query commands from modules that implement
	// GetQueryCmd() (ibc, transfer, wasm). Core SDK modules use AutoCLI for
	// queries in SDK 0.50 and do not implement GetQueryCmd().
	basicManager.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand(basicManager module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cfg := sdk.GetConfig()
	accCodec := addresscodec.NewBech32Codec(cfg.GetBech32AccountAddrPrefix())
	valCodec := addresscodec.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix())

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
		// Module tx commands requiring address codecs
		bankcli.NewTxCmd(accCodec),
		stakingcli.NewTxCmd(valCodec, accCodec),
		distrcli.NewTxCmd(valCodec, accCodec),
		govcli.NewTxCmd(nil),
		authzcli.GetTxCmd(accCodec),
		feegrantcli.GetTxCmd(accCodec),
		upgradecli.GetTxCmd(accCodec),
		// IBC and wasm tx commands (no address codecs needed)
		ibccli.GetTxCmd(),
		transfercli.NewTxCmd(),
		wasmcli.GetTxCmd(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// makeAppCreator returns an AppCreator function
func makeAppCreator(encodingConfig params.EncodingConfig) servertypes.AppCreator {
	return func(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {

		var cache storetypes.MultiStorePersistentCache

		if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
			cache = store.NewCommitKVStoreCacheManager()
		}

		skipUpgradeHeights := make(map[int64]bool)
		for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
			skipUpgradeHeights[int64(h)] = true
		}

		pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
		if err != nil {
			panic(err)
		}

		snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
		snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
		if err != nil {
			panic(err)
		}
		snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
		if err != nil {
			panic(err)
		}

		snapshotOptions := snapshottypes.NewSnapshotOptions(
			cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
			cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
		)

		var wasmOpts []wasm.Option
		if cast.ToBool(appOpts.Get("telemetry.enabled")) {
			wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
		}

		return app.NewWasmApp(logger, db, traceStore, true, skipUpgradeHeights,
			cast.ToString(appOpts.Get(flags.FlagHome)),
			cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
			encodingConfig,
			app.GetEnabledProposals(),
			appOpts,
			wasmOpts,
			baseapp.SetPruning(pruningOpts),
			baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
			baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
			baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
			baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
			baseapp.SetInterBlockCache(cache),
			baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
			baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
			baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		)
	}
}

// makeAppExporter returns an AppExporter function
func makeAppExporter(encodingConfig params.EncodingConfig) servertypes.AppExporter {
	return func(logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOpts servertypes.AppOptions, modulesToExport []string) (servertypes.ExportedApp, error) {

		var wasmApp *app.WasmApp
		homePath, ok := appOpts.Get(flags.FlagHome).(string)
		if !ok || homePath == "" {
			return servertypes.ExportedApp{}, errors.New("application home is not set")
		}

		loadLatest := height == -1
		var emptyWasmOpts []wasm.Option
		wasmApp = app.NewWasmApp(
			logger,
			db,
			traceStore,
			loadLatest,
			map[int64]bool{},
			homePath,
			cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
			encodingConfig,
			app.GetEnabledProposals(),
			appOpts,
			emptyWasmOpts,
		)

		if height != -1 {
			if err := wasmApp.LoadHeight(height); err != nil {
				return servertypes.ExportedApp{}, err
			}
		}

		return wasmApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
	}
}
