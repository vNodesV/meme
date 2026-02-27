package app

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmWasm/wasmd/x/wasm"
)

var emptyWasmOpts []wasm.Option = nil

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (EmptyAppOptions) Get(string) interface{} { return nil }

func TestWasmdExport(t *testing.T) {
	db, err := dbm.NewDB("test", dbm.MemDBBackend, "")
	require.NoError(t, err)
	gapp := NewWasmApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, t.TempDir(), 0, MakeEncodingConfig(), wasm.EnableAllProposals, EmptyAppOptions{}, emptyWasmOpts)

	// Build a genesis state that includes a validator so InitChain succeeds.
	valSet, err := sims.CreateRandomValidatorSet()
	require.NoError(t, err)
	minterAddr := sdk.AccAddress("test_minter_address__")
	genAccs := []authtypes.GenesisAccount{authtypes.NewBaseAccountWithAddress(minterAddr)}
	genesisState := NewDefaultGenesisState()
	genesisState, err = sims.GenesisStateWithValSet(
		MakeEncodingConfig().Marshaler,
		genesisState,
		valSet,
		genAccs,
		banktypes.Balance{
			Address: minterAddr.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000_000)),
		},
	)
	require.NoError(t, err)

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	_, err = gapp.InitChain(&abci.RequestInitChain{
		Validators:    cmttypes.TM2PB.ValidatorUpdates(valSet),
		AppStateBytes: stateBytes,
	})
	require.NoError(t, err)
	_, err = gapp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)
	_, err = gapp.Commit()
	require.NoError(t, err)

	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewWasmApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, t.TempDir(), 0, MakeEncodingConfig(), wasm.EnableAllProposals, EmptyAppOptions{}, emptyWasmOpts)
	_, err = newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// ensure that blocked addresses are properly set in bank keeper
func TestBlockedAddrs(t *testing.T) {
	db, err := dbm.NewDB("test", dbm.MemDBBackend, "")
	require.NoError(t, err)
	gapp := NewWasmApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, t.TempDir(), 0, MakeEncodingConfig(), wasm.EnableAllProposals, EmptyAppOptions{}, emptyWasmOpts)

	for acc := range maccPerms {
		t.Run(acc, func(t *testing.T) {
			require.True(t, gapp.bankKeeper.BlockedAddr(gapp.accountKeeper.GetModuleAddress(acc)),
				"ensure that blocked addresses are properly set in bank keeper",
			)
		})
	}
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func TestGetEnabledProposals(t *testing.T) {
	cases := map[string]struct {
		proposalsEnabled string
		specificEnabled  string
		expected         []wasm.ProposalType
	}{
		"all disabled": {
			proposalsEnabled: "false",
			expected:         wasm.DisableAllProposals,
		},
		"all enabled": {
			proposalsEnabled: "true",
			expected:         wasm.EnableAllProposals,
		},
		"some enabled": {
			proposalsEnabled: "okay",
			specificEnabled:  "StoreCode,InstantiateContract",
			expected:         []wasm.ProposalType{wasm.ProposalTypeStoreCode, wasm.ProposalTypeInstantiateContract},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ProposalsEnabled = tc.proposalsEnabled
			EnableSpecificProposals = tc.specificEnabled
			proposals := GetEnabledProposals()
			assert.Equal(t, tc.expected, proposals)
		})
	}
}

func setGenesis(gapp *WasmApp) error {
	genesisState := NewDefaultGenesisState()
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		return err
	}

	// Initialize the chain
	_, err = gapp.InitChain(&abci.RequestInitChain{
		Validators:    []abci.ValidatorUpdate{},
		AppStateBytes: stateBytes,
	})
	if err != nil {
		return err
	}

	_, err = gapp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	if err != nil {
		return err
	}

	_, err = gapp.Commit()
	return err
}
