package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	portkeeper "github.com/cosmos/ibc-go/v8/modules/core/05-port/keeper"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/stretchr/testify/require"

	wasmappparams "github.com/CosmWasm/wasmd/app/params"

	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	"github.com/CosmWasm/wasmd/x/wasm/types"
)

// testStakingKeeperAdapter adapts SDK 0.50 staking keeper to the wasmd StakingKeeper interface
type testStakingKeeperAdapter struct {
	*stakingkeeper.Keeper
}

func (s testStakingKeeperAdapter) BondDenom(ctx sdk.Context) string {
	denom, err := s.Keeper.BondDenom(ctx)
	if err != nil {
		panic("failed to get bond denom: " + err.Error())
	}
	return denom
}

func (s testStakingKeeperAdapter) GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation {
	delegations, err := s.Keeper.GetAllDelegatorDelegations(ctx, delegator)
	if err != nil {
		return nil
	}
	return delegations
}

func (s testStakingKeeperAdapter) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	validators, err := s.Keeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return nil
	}
	return validators
}

func (s testStakingKeeperAdapter) GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, bool) {
	delegation, err := s.Keeper.GetDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return stakingtypes.Delegation{}, false
	}
	return delegation, true
}

func (s testStakingKeeperAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	validator, err := s.Keeper.GetValidator(ctx, addr)
	if err != nil {
		return stakingtypes.Validator{}, false
	}
	return validator, true
}

func (s testStakingKeeperAdapter) HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool {
	has, err := s.Keeper.HasReceivingRedelegation(ctx, delAddr, valDstAddr)
	if err != nil {
		return false
	}
	return has
}

// testDistributionKeeperAdapter adapts SDK 0.50 distribution keeper to the wasmd DistributionKeeper interface
type testDistributionKeeperAdapter struct {
	distributionkeeper.Keeper
	querier distributionkeeper.Querier
}

func newTestDistributionKeeperAdapter(dk distributionkeeper.Keeper) testDistributionKeeperAdapter {
	return testDistributionKeeperAdapter{
		Keeper:  dk,
		querier: distributionkeeper.NewQuerier(dk),
	}
}

func (d testDistributionKeeperAdapter) DelegationRewards(c context.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	return d.querier.DelegationRewards(c, req)
}

// testAccountKeeperAdapter adapts SDK 0.50 account keeper to the wasmd AccountKeeper interface
type testAccountKeeperAdapter struct {
	authkeeper.AccountKeeper
}

func (a testAccountKeeperAdapter) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return a.AccountKeeper.GetAccount(ctx, addr)
}

func (a testAccountKeeperAdapter) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return a.AccountKeeper.NewAccountWithAddress(ctx, addr)
}

func (a testAccountKeeperAdapter) SetAccount(ctx sdk.Context, acc authtypes.AccountI) {
	a.AccountKeeper.SetAccount(ctx, acc)
}

// testBankKeeperAdapter adapts SDK 0.50 bank keeper to the wasmd BankKeeper interface
type testBankKeeperAdapter struct {
	bankkeeper.Keeper
}

func (b testBankKeeperAdapter) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return b.Keeper.BurnCoins(ctx, moduleName, amt)
}

func (b testBankKeeperAdapter) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return b.Keeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

func (b testBankKeeperAdapter) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	return b.Keeper.IsSendEnabledCoins(ctx, coins...)
}

func (b testBankKeeperAdapter) BlockedAddr(addr sdk.AccAddress) bool {
	return b.Keeper.BlockedAddr(addr)
}

func (b testBankKeeperAdapter) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return b.Keeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

func (b testBankKeeperAdapter) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return b.Keeper.GetAllBalances(ctx, addr)
}

func (b testBankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return b.Keeper.GetBalance(ctx, addr, denom)
}

// testChannelKeeperAdapter adapts IBC channel keeper to the wasmd ChannelKeeper interface
type testChannelKeeperAdapter struct {
	ibcChannelKeeper channelkeeper.Keeper
}

func (c testChannelKeeperAdapter) GetChannel(ctx sdk.Context, srcPort, srcChan string) (channeltypes.Channel, bool) {
	return c.ibcChannelKeeper.GetChannel(ctx, srcPort, srcChan)
}

func (c testChannelKeeperAdapter) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	return c.ibcChannelKeeper.GetNextSequenceSend(ctx, portID, channelID)
}

func (c testChannelKeeperAdapter) SendPacket(ctx sdk.Context, packet ibcexported.PacketI) error {
	_, err := c.ibcChannelKeeper.SendPacket(ctx, nil, packet.GetSourcePort(), packet.GetSourceChannel(),
		packet.GetTimeoutHeight().(clienttypes.Height), packet.GetTimeoutTimestamp(), packet.GetData())
	return err
}

func (c testChannelKeeperAdapter) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return c.ibcChannelKeeper.ChanCloseInit(ctx, portID, channelID, nil)
}

func (c testChannelKeeperAdapter) GetAllChannels(ctx sdk.Context) []channeltypes.IdentifiedChannel {
	return c.ibcChannelKeeper.GetAllChannels(ctx)
}

func (c testChannelKeeperAdapter) IterateChannels(ctx sdk.Context, cb func(channeltypes.IdentifiedChannel) bool) {
	c.ibcChannelKeeper.IterateChannels(ctx, cb)
}

// testPortKeeperAdapter adapts IBC port keeper to the wasmd PortKeeper interface
type testPortKeeperAdapter struct {
	ibcPortKeeper *portkeeper.Keeper
}

func (p testPortKeeperAdapter) BindPort(ctx sdk.Context, portID string) error {
	cap := p.ibcPortKeeper.BindPort(ctx, portID)
	if cap == nil {
		return fmt.Errorf("failed to bind port %s", portID)
	}
	return nil
}

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.NewAppModuleBasic(nil), // Pass nil for legacy proposal handlers
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	ibc.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transfer.AppModuleBasic{},
)

func MakeTestCodec(t testing.TB) codec.Codec {
	return MakeEncodingConfig(t).Marshaler
}

func MakeEncodingConfig(_ testing.TB) wasmappparams.EncodingConfig {
	encodingConfig := wasmappparams.MakeEncodingConfig()
	amino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	ModuleBasics.RegisterLegacyAminoCodec(amino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	// add wasmd types
	types.RegisterInterfaces(interfaceRegistry)
	types.RegisterLegacyAminoCodec(amino)

	return encodingConfig
}

var TestingStakeParams = stakingtypes.Params{
	UnbondingTime:     100,
	MaxValidators:     10,
	MaxEntries:        10,
	HistoricalEntries: 10,
	BondDenom:         "stake",
	MinCommissionRate: math.LegacyZeroDec(),
}

type TestFaucet struct {
	t                testing.TB
	bankKeeper       bankkeeper.Keeper
	sender           sdk.AccAddress
	balance          sdk.Coins
	minterModuleName string
}

func NewTestFaucet(t testing.TB, ctx sdk.Context, bankKeeper bankkeeper.Keeper, minterModuleName string, initialAmount ...sdk.Coin) *TestFaucet {
	require.NotEmpty(t, initialAmount)
	r := &TestFaucet{t: t, bankKeeper: bankKeeper, minterModuleName: minterModuleName}
	_, _, addr := keyPubAddr()
	r.sender = addr
	r.Mint(ctx, addr, initialAmount...)
	r.balance = initialAmount
	return r
}

func (f *TestFaucet) Mint(parentCtx sdk.Context, addr sdk.AccAddress, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.MintCoins(ctx, f.minterModuleName, amounts)
	require.NoError(f.t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(ctx, f.minterModuleName, addr, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Add(amounts...)
}

func (f *TestFaucet) Fund(parentCtx sdk.Context, receiver sdk.AccAddress, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	// ensure faucet is always filled
	if !f.balance.IsAllGTE(amounts) {
		f.Mint(parentCtx, f.sender, amounts...)
	}
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.SendCoins(ctx, f.sender, receiver, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Sub(amounts...)
}

func (f *TestFaucet) NewFundedAccount(ctx sdk.Context, amounts ...sdk.Coin) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	f.Fund(ctx, addr, amounts...)
	return addr
}

type TestKeepers struct {
	AccountKeeper  authkeeper.AccountKeeper
	StakingKeeper  *stakingkeeper.Keeper
	DistKeeper     distributionkeeper.Keeper
	BankKeeper     bankkeeper.Keeper
	GovKeeper      *govkeeper.Keeper
	ContractKeeper types.ContractOpsKeeper
	WasmKeeper     *Keeper
	IBCKeeper      *ibckeeper.Keeper
	EncodingConfig wasmappparams.EncodingConfig
	Faucet         *TestFaucet
}

// CreateDefaultTestInput common settings for CreateTestInput
func CreateDefaultTestInput(t testing.TB) (sdk.Context, TestKeepers) {
	return CreateTestInput(t, false, "staking")
}

// CreateTestInput encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func CreateTestInput(t testing.TB, isCheckTx bool, supportedFeatures string, opts ...Option) (sdk.Context, TestKeepers) {
	// Load default wasm config
	return createTestInput(t, isCheckTx, supportedFeatures, types.DefaultWasmConfig(), dbm.NewMemDB(), opts...)
}

// encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func createTestInput(
	t testing.TB,
	isCheckTx bool,
	supportedFeatures string,
	wasmConfig types.WasmConfig,
	db dbm.DB,
	opts ...Option,
) (sdk.Context, TestKeepers) {
	tempDir := t.TempDir()

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distributiontypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibcexported.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		feegrant.StoreKey, authzkeeper.StoreKey,
		capabilitytypes.StoreKey,
		types.StoreKey,
	)
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	for _, v := range keys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
	}
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	for _, v := range tkeys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeTransient, db)
	}

	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	for _, v := range memKeys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeMemory, db)
	}

	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, isCheckTx, log.NewNopLogger())
	ctx = types.WithTXCounter(ctx, 0)

	encodingConfig := MakeEncodingConfig(t)
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	paramsKeeper := paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)
	for _, m := range []string{authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		distributiontypes.ModuleName,
		slashingtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		govtypes.ModuleName,
		types.ModuleName} {
		paramsKeeper.Subspace(m)
	}
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}

	// Create authority address for modules
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Create address codecs
	ac := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	valAddrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())
	consAddrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix())

	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:     nil,
		distributiontypes.ModuleName:   nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		types.ModuleName:               {authtypes.Burner},
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
		ac,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authority,
	)
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddrs,
		authority,
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority,
		valAddrCodec,
		consAddrCodec,
	)
	require.NoError(t, stakingKeeper.SetParams(ctx, TestingStakeParams))

	distKeeper := distributionkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distributiontypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authority,
	)
	require.NoError(t, distKeeper.Params.Set(ctx, distributiontypes.DefaultParams()))
	require.NoError(t, distKeeper.FeePool.Set(ctx, distributiontypes.InitialFeePool()))
	stakingKeeper.SetHooks(distKeeper.Hooks())

	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		tempDir,
		nil,
		authority,
	)

	faucet := NewTestFaucet(t, ctx, bankKeeper, minttypes.ModuleName, sdk.NewCoin("stake", math.NewInt(100_000_000_000)))

	// set some funds to pay out validators, based on code from:
	// https://github.com/cosmos/cosmos-sdk/blob/fea231556aee4d549d7551a6190389c4328194eb/x/distribution/keeper/keeper_test.go#L50-L57
	distrAcc := distKeeper.GetDistributionAccount(ctx)
	faucet.Fund(ctx, distrAcc.GetAddress(), sdk.NewCoin("stake", math.NewInt(2000000)))
	accountKeeper.SetModuleAccount(ctx, distrAcc)

	// Initialize capability keeper for IBC
	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	capabilityKeeper.Seal()

	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		subspace(ibcexported.ModuleName),
		stakingKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
		authority,
	)

	querier := baseapp.NewGRPCQueryRouter()
	querier.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)

	cfg := sdk.GetConfig()
	cfg.SetAddressVerifier(types.VerifyAddressLen())

	// Create keeper adapters for wasm keeper (SDK 0.50 interface bridge)
	accountKeeperAdapter := testAccountKeeperAdapter{accountKeeper}
	bankKeeperAdapter := testBankKeeperAdapter{bankKeeper}
	stakingKeeperAdapter := testStakingKeeperAdapter{stakingKeeper}
	distKeeperAdapter := newTestDistributionKeeperAdapter(distKeeper)

	// Create IBC adapters for wasm keeper
	channelKeeperAdapter := testChannelKeeperAdapter{ibcChannelKeeper: ibcKeeper.ChannelKeeper}
	portKeeperAdapter := testPortKeeperAdapter{ibcPortKeeper: ibcKeeper.PortKeeper}

	keeper := NewKeeper(
		appCodec,
		keys[types.StoreKey],
		subspace(types.ModuleName),
		accountKeeperAdapter,
		bankKeeperAdapter,
		stakingKeeperAdapter,
		distKeeperAdapter,
		channelKeeperAdapter,
		portKeeperAdapter,
		wasmtesting.MockIBCTransferKeeper{},
		msgRouter,
		querier,
		tempDir,
		wasmConfig,
		supportedFeatures,
		opts...,
	)
	keeper.SetParams(ctx, types.DefaultParams())
	// add wasm handler so we can loop-back (contracts calling contracts)
	contractKeeper := NewDefaultPermissionKeeper(&keeper)

	am := module.NewManager( // minimal module set that we use for message/ query tests
		bank.NewAppModule(appCodec, bankKeeper, accountKeeper, subspace(banktypes.ModuleName)),
		staking.NewAppModule(appCodec, stakingKeeper, accountKeeper, bankKeeper, subspace(stakingtypes.ModuleName)),
		distribution.NewAppModule(appCodec, distKeeper, accountKeeper, bankKeeper, stakingKeeper, subspace(distributiontypes.ModuleName)),
	)
	am.RegisterServices(module.NewConfigurator(appCodec, msgRouter, querier))
	types.RegisterMsgServer(msgRouter, NewMsgServerImpl(NewDefaultPermissionKeeper(keeper)))
	types.RegisterQueryServer(querier, NewGrpcQuerier(appCodec, keys[types.ModuleName], keeper, keeper.queryGasLimit))

	govRouter := govv1beta1.NewRouter().
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(types.RouterKey, NewWasmProposalHandler(&keeper, types.EnableAllProposals))

	govConfig := govtypes.DefaultConfig()
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distKeeper,
		msgRouter,
		govConfig,
		authority,
	)
	govKeeper.SetLegacyRouter(govRouter)

	require.NoError(t, govKeeper.Params.Set(ctx, govv1.DefaultParams()))
	require.NoError(t, govKeeper.ProposalID.Set(ctx, govv1.DefaultStartingProposalID))
	require.NoError(t, govKeeper.Constitution.Set(ctx, ""))

	keepers := TestKeepers{
		AccountKeeper:  accountKeeper,
		StakingKeeper:  stakingKeeper,
		DistKeeper:     distKeeper,
		ContractKeeper: contractKeeper,
		WasmKeeper:     &keeper,
		BankKeeper:     bankKeeper,
		GovKeeper:      govKeeper,
		IBCKeeper:      ibcKeeper,
		EncodingConfig: encodingConfig,
		Faucet:         faucet,
	}
	return ctx, keepers
}

func RandomAccountAddress(_ testing.TB) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress(t testing.TB) string {
	return RandomAccountAddress(t).String()
}

type ExampleContract struct {
	InitialAmount sdk.Coins
	Creator       crypto.PrivKey
	CreatorAddr   sdk.AccAddress
	CodeID        uint64
}

func StoreHackatomExampleContract(t testing.TB, ctx sdk.Context, keepers TestKeepers) ExampleContract {
	return StoreExampleContract(t, ctx, keepers, "./testdata/hackatom.wasm")
}

func StoreBurnerExampleContract(t testing.TB, ctx sdk.Context, keepers TestKeepers) ExampleContract {
	return StoreExampleContract(t, ctx, keepers, "./testdata/burner.wasm")
}

func StoreIBCReflectContract(t testing.TB, ctx sdk.Context, keepers TestKeepers) ExampleContract {
	return StoreExampleContract(t, ctx, keepers, "./testdata/ibc_reflect.wasm")
}

func StoreReflectContract(t testing.TB, ctx sdk.Context, keepers TestKeepers) uint64 {
	wasmCode, err := os.ReadFile("./testdata/reflect.wasm")
	require.NoError(t, err)

	_, _, creatorAddr := keyPubAddr()
	codeID, err := keepers.ContractKeeper.Create(ctx, creatorAddr, wasmCode, nil)
	require.NoError(t, err)
	return codeID
}

func StoreExampleContract(t testing.TB, ctx sdk.Context, keepers TestKeepers, wasmFile string) ExampleContract {
	anyAmount := sdk.NewCoins(sdk.NewInt64Coin("denom", 1000))
	creator, _, creatorAddr := keyPubAddr()
	fundAccounts(t, ctx, keepers.AccountKeeper, keepers.BankKeeper, creatorAddr, anyAmount)

	wasmCode, err := os.ReadFile(wasmFile)
	require.NoError(t, err)

	codeID, err := keepers.ContractKeeper.Create(ctx, creatorAddr, wasmCode, nil)
	require.NoError(t, err)
	return ExampleContract{anyAmount, creator, creatorAddr, codeID}
}

var wasmIdent = []byte("\x00\x61\x73\x6D")

type ExampleContractInstance struct {
	ExampleContract
	Contract sdk.AccAddress
}

// SeedNewContractInstance sets the mock wasmerEngine in keeper and calls store + instantiate to init the contract's metadata
func SeedNewContractInstance(t testing.TB, ctx sdk.Context, keepers TestKeepers, mock types.WasmerEngine) ExampleContractInstance {
	t.Helper()
	exampleContract := StoreRandomContract(t, ctx, keepers, mock)
	contractAddr, _, err := keepers.ContractKeeper.Instantiate(ctx, exampleContract.CodeID, exampleContract.CreatorAddr, exampleContract.CreatorAddr, []byte(`{}`), "", nil)
	require.NoError(t, err)
	return ExampleContractInstance{
		ExampleContract: exampleContract,
		Contract:        contractAddr,
	}
}

// StoreRandomContract sets the mock wasmerEngine in keeper and calls store
func StoreRandomContract(t testing.TB, ctx sdk.Context, keepers TestKeepers, mock types.WasmerEngine) ExampleContract {
	t.Helper()
	anyAmount := sdk.NewCoins(sdk.NewInt64Coin("denom", 1000))
	creator, _, creatorAddr := keyPubAddr()
	fundAccounts(t, ctx, keepers.AccountKeeper, keepers.BankKeeper, creatorAddr, anyAmount)
	// NOTE: In wasmvm v2, keeper.wasmVM is concrete *wasmvm.VM and cannot accept
	// mock WasmerEngine interfaces. Mock-based tests require wasmvm v2 migration.
	// For now, the keeper retains its default VM.
	wasmCode := append(wasmIdent, rand.Bytes(10)...) //nolint:gocritic
	codeID, err := keepers.ContractKeeper.Create(ctx, creatorAddr, wasmCode, nil)
	require.NoError(t, err)
	exampleContract := ExampleContract{InitialAmount: anyAmount, Creator: creator, CreatorAddr: creatorAddr, CodeID: codeID}
	return exampleContract
}

type HackatomExampleInstance struct {
	ExampleContract
	Contract        sdk.AccAddress
	Verifier        crypto.PrivKey
	VerifierAddr    sdk.AccAddress
	Beneficiary     crypto.PrivKey
	BeneficiaryAddr sdk.AccAddress
}

// InstantiateHackatomExampleContract load and instantiate the "./testdata/hackatom.wasm" contract
func InstantiateHackatomExampleContract(t testing.TB, ctx sdk.Context, keepers TestKeepers) HackatomExampleInstance {
	contract := StoreHackatomExampleContract(t, ctx, keepers)

	verifier, _, verifierAddr := keyPubAddr()
	fundAccounts(t, ctx, keepers.AccountKeeper, keepers.BankKeeper, verifierAddr, contract.InitialAmount)

	beneficiary, _, beneficiaryAddr := keyPubAddr()
	initMsgBz := HackatomExampleInitMsg{
		Verifier:    verifierAddr,
		Beneficiary: beneficiaryAddr,
	}.GetBytes(t)
	initialAmount := sdk.NewCoins(sdk.NewInt64Coin("denom", 100))

	adminAddr := contract.CreatorAddr
	contractAddr, _, err := keepers.ContractKeeper.Instantiate(ctx, contract.CodeID, contract.CreatorAddr, adminAddr, initMsgBz, "demo contract to query", initialAmount)
	require.NoError(t, err)
	return HackatomExampleInstance{
		ExampleContract: contract,
		Contract:        contractAddr,
		Verifier:        verifier,
		VerifierAddr:    verifierAddr,
		Beneficiary:     beneficiary,
		BeneficiaryAddr: beneficiaryAddr,
	}
}

type HackatomExampleInitMsg struct {
	Verifier    sdk.AccAddress `json:"verifier"`
	Beneficiary sdk.AccAddress `json:"beneficiary"`
}

func (m HackatomExampleInitMsg) GetBytes(t testing.TB) []byte {
	initMsgBz, err := json.Marshal(m)
	require.NoError(t, err)
	return initMsgBz
}

type IBCReflectExampleInstance struct {
	Contract      sdk.AccAddress
	Admin         sdk.AccAddress
	CodeID        uint64
	ReflectCodeID uint64
}

// InstantiateIBCReflectContract load and instantiate the "./testdata/ibc_reflect.wasm" contract
func InstantiateIBCReflectContract(t testing.TB, ctx sdk.Context, keepers TestKeepers) IBCReflectExampleInstance {
	reflectID := StoreReflectContract(t, ctx, keepers)
	ibcReflectID := StoreIBCReflectContract(t, ctx, keepers).CodeID

	initMsgBz := IBCReflectInitMsg{
		ReflectCodeID: reflectID,
	}.GetBytes(t)
	adminAddr := RandomAccountAddress(t)

	contractAddr, _, err := keepers.ContractKeeper.Instantiate(ctx, ibcReflectID, adminAddr, adminAddr, initMsgBz, "ibc-reflect-factory", nil)
	require.NoError(t, err)
	return IBCReflectExampleInstance{
		Admin:         adminAddr,
		Contract:      contractAddr,
		CodeID:        ibcReflectID,
		ReflectCodeID: reflectID,
	}
}

type IBCReflectInitMsg struct {
	ReflectCodeID uint64 `json:"reflect_code_id"`
}

func (m IBCReflectInitMsg) GetBytes(t testing.TB) []byte {
	initMsgBz, err := json.Marshal(m)
	require.NoError(t, err)
	return initMsgBz
}

type BurnerExampleInitMsg struct {
	Payout sdk.AccAddress `json:"payout"`
}

func (m BurnerExampleInitMsg) GetBytes(t testing.TB) []byte {
	initMsgBz, err := json.Marshal(m)
	require.NoError(t, err)
	return initMsgBz
}

func fundAccounts(t testing.TB, ctx sdk.Context, am authkeeper.AccountKeeper, bank bankkeeper.Keeper, addr sdk.AccAddress, coins sdk.Coins) {
	acc := am.NewAccountWithAddress(ctx, addr)
	am.SetAccount(ctx, acc)
	NewTestFaucet(t, ctx, bank, minttypes.ModuleName, coins...).Fund(ctx, addr, coins...)
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := ed25519.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}
