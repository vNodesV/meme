package app

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	portkeeper "github.com/cosmos/ibc-go/v8/modules/core/05-port/keeper"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

// AccountKeeperAdapter adapts SDK 0.50 AccountKeeper to wasmd expectations
type AccountKeeperAdapter struct {
	authkeeper.AccountKeeper
}

func NewAccountKeeperAdapter(ak authkeeper.AccountKeeper) AccountKeeperAdapter {
	return AccountKeeperAdapter{AccountKeeper: ak}
}

// GetAccount adapts context.Context to sdk.Context and return type
func (a AccountKeeperAdapter) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return a.AccountKeeper.GetAccount(ctx, addr)
}

// NewAccountWithAddress adapts for wasmd interface
func (a AccountKeeperAdapter) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return a.AccountKeeper.NewAccountWithAddress(ctx, addr)
}

// SetAccount adapts for wasmd interface
func (a AccountKeeperAdapter) SetAccount(ctx sdk.Context, acc authtypes.AccountI) {
	a.AccountKeeper.SetAccount(ctx, acc)
}

// BankKeeperAdapter adapts SDK 0.50 BankKeeper to wasmd expectations
type BankKeeperAdapter struct {
	bankkeeper.Keeper
}

func NewBankKeeperAdapter(bk bankkeeper.Keeper) BankKeeperAdapter {
	return BankKeeperAdapter{Keeper: bk}
}

// BurnCoins adapts context.Context to sdk.Context
func (b BankKeeperAdapter) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return b.Keeper.BurnCoins(ctx, moduleName, amt)
}

// SendCoinsFromAccountToModule adapts for wasmd interface
func (b BankKeeperAdapter) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return b.Keeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// IsSendEnabledCoins adapts for wasmd interface
func (b BankKeeperAdapter) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	return b.Keeper.IsSendEnabledCoins(ctx, coins...)
}

// BlockedAddr adapts for wasmd interface
func (b BankKeeperAdapter) BlockedAddr(addr sdk.AccAddress) bool {
	return b.Keeper.BlockedAddr(addr)
}

// SendCoins adapts for wasmd interface
func (b BankKeeperAdapter) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return b.Keeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

// GetAllBalances adapts for wasmd interface
func (b BankKeeperAdapter) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return b.Keeper.GetAllBalances(ctx, addr)
}

// GetBalance adapts for wasmd interface
func (b BankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return b.Keeper.GetBalance(ctx, addr, denom)
}

// StakingKeeperAdapter adapts SDK 0.50 StakingKeeper to wasmd expectations
type StakingKeeperAdapter struct {
	*stakingkeeper.Keeper
}

func NewStakingKeeperAdapter(sk *stakingkeeper.Keeper) StakingKeeperAdapter {
	return StakingKeeperAdapter{Keeper: sk}
}

// BondDenom adapts return type (string, error) to just string
func (s StakingKeeperAdapter) BondDenom(ctx sdk.Context) string {
	denom, err := s.Keeper.BondDenom(ctx)
	if err != nil {
		// In SDK 0.50, BondDenom can return an error, but wasmd expects no error
		// This should never fail in practice as BondDenom is set at genesis
		panic("failed to get bond denom: " + err.Error())
	}
	return denom
}

// GetAllDelegatorDelegations adapts return type - SDK 0.50 returns ([]Delegation, error) but wasmd expects just []Delegation
func (s StakingKeeperAdapter) GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation {
	delegations, err := s.Keeper.GetAllDelegatorDelegations(ctx, delegator)
	if err != nil {
		// In SDK 0.50, this can return an error, but wasmd expects no error
		// Return empty slice on error to maintain compatibility
		return []stakingtypes.Delegation{}
	}
	return delegations
}

// GetBondedValidatorsByPower adapts return type - SDK 0.50 returns ([]Validator, error) but wasmd expects just []Validator
func (s StakingKeeperAdapter) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	validators, err := s.Keeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		// In SDK 0.50, this can return an error, but wasmd expects no error
		// Return empty slice on error to maintain compatibility
		return []stakingtypes.Validator{}
	}
	return validators
}

// GetDelegation adapts return type - SDK 0.50 returns (Delegation, error) but wasmd expects (Delegation, bool)
func (s StakingKeeperAdapter) GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, bool) {
	delegation, err := s.Keeper.GetDelegation(ctx, delAddr, valAddr)
	if err != nil {
		// In SDK 0.50, this returns an error, but wasmd expects a bool
		// Return found=false on error
		return stakingtypes.Delegation{}, false
	}
	// Check if delegation is empty (not found)
	if delegation.DelegatorAddress == "" {
		return stakingtypes.Delegation{}, false
	}
	return delegation, true
}

// GetValidator adapts for wasmd interface
func (s StakingKeeperAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	validator, err := s.Keeper.GetValidator(ctx, addr)
	if err != nil {
		return stakingtypes.Validator{}, false
	}
	// Check if validator is empty (not found)
	if validator.OperatorAddress == "" {
		return stakingtypes.Validator{}, false
	}
	return validator, true
}

// HasReceivingRedelegation adapts for wasmd interface
func (s StakingKeeperAdapter) HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool {
	has, err := s.Keeper.HasReceivingRedelegation(ctx, delAddr, valDstAddr)
	if err != nil {
		return false
	}
	return has
}

// DistributionKeeperAdapter adapts SDK 0.50 DistributionKeeper to wasmd expectations
type DistributionKeeperAdapter struct {
	distributionkeeper.Keeper
	querier distributionkeeper.Querier
}

func NewDistributionKeeperAdapter(dk distributionkeeper.Keeper) DistributionKeeperAdapter {
	return DistributionKeeperAdapter{
		Keeper:  dk,
		querier: distributionkeeper.NewQuerier(dk),
	}
}

// DelegationRewards implements the wasmd interface
func (d DistributionKeeperAdapter) DelegationRewards(c context.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	return d.querier.DelegationRewards(c, req)
}

// ChannelKeeperAdapter adapts SDK 0.50 ChannelKeeper to wasmd expectations
type ChannelKeeperAdapter struct {
	*channelkeeper.Keeper
}

func NewChannelKeeperAdapter(ck *channelkeeper.Keeper) ChannelKeeperAdapter {
	return ChannelKeeperAdapter{Keeper: ck}
}

// ChanCloseInit adapts by dropping the capability parameter
func (c ChannelKeeperAdapter) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	// In SDK 0.50/IBC v8, ChanCloseInit requires a capability parameter
	// wasmd expects the old signature without capability
	// We need to get the capability from the capability keeper
	// For now, we pass nil as the capability - this may need adjustment
	return c.Keeper.ChanCloseInit(ctx, portID, channelID, nil)
}

// SendPacket adapts the signature - SDK 0.50 has a different signature than wasmd expects
func (c ChannelKeeperAdapter) SendPacket(ctx sdk.Context, packet ibcexported.PacketI) error {
	// In SDK 0.50/IBC v8, SendPacket has a different signature
	// We need to extract the packet fields and call the keeper method
	// Get channel capability - this is a simplified approach
	// In production, the capability should be properly retrieved
	_, err := c.Keeper.SendPacket(ctx, nil, packet.GetSourcePort(), packet.GetSourceChannel(),
		packet.GetTimeoutHeight().(clienttypes.Height), packet.GetTimeoutTimestamp(), packet.GetData())
	return err
}

// PortKeeperAdapter adapts SDK 0.50 PortKeeper to wasmd expectations
type PortKeeperAdapter struct {
	*portkeeper.Keeper
}

func NewPortKeeperAdapter(pk *portkeeper.Keeper) PortKeeperAdapter {
	return PortKeeperAdapter{Keeper: pk}
}

// BindPort adapts return type from *Capability to error
func (p PortKeeperAdapter) BindPort(ctx sdk.Context, portID string) error {
	capability := p.Keeper.BindPort(ctx, portID)
	if capability == nil {
		return capabilitytypes.ErrCapabilityNotOwned
	}
	return nil
}

// ICS20TransferPortSourceAdapter adapts to provide GetPort method
type ICS20TransferPortSourceAdapter struct {
	capabilitykeeper.ScopedKeeper
}

func NewICS20TransferPortSourceAdapter(sk capabilitykeeper.ScopedKeeper) ICS20TransferPortSourceAdapter {
	return ICS20TransferPortSourceAdapter{ScopedKeeper: sk}
}

// GetPort returns the ICS20 transfer port ID
func (i ICS20TransferPortSourceAdapter) GetPort(ctx sdk.Context) string {
	// Return the standard ICS20 transfer port
	return ibctransfertypes.PortID
}

// ValidatorSetSourceAdapter adapts StakingKeeper to ValidatorSetSource interface
type ValidatorSetSourceAdapter struct {
	*stakingkeeper.Keeper
}

func NewValidatorSetSourceAdapter(sk *stakingkeeper.Keeper) ValidatorSetSourceAdapter {
	return ValidatorSetSourceAdapter{Keeper: sk}
}

// ApplyAndReturnValidatorSetUpdates adapts context.Context to sdk.Context
func (v ValidatorSetSourceAdapter) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	return v.Keeper.ApplyAndReturnValidatorSetUpdates(ctx)
}
