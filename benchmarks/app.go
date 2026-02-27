package benchmarks

import (
	"context"
	"encoding/json"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktxsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/CosmWasm/wasmd/app"
	"github.com/CosmWasm/wasmd/x/wasm"
)

// AppInfo holds state needed to run benchmark transactions against a WasmApp.
type AppInfo struct {
	App          *app.WasmApp
	TxConfig     client.TxConfig
	Denom        string
	MinterAddr   sdk.AccAddress
	PrivKey      cryptotypes.PrivKey
	ContractAddr string
	AccNum       uint64
	SeqNum       uint64
}

// InitializeWasmApp creates a WasmApp with a funded account and an initialized
// chain.  numAccounts controls how many additional genesis accounts are created
// (for state-size simulation); the minter account is always created.
func InitializeWasmApp(b testing.TB, db dbm.DB, numAccounts int) AppInfo {
	b.Helper()

	encConfig := app.MakeEncodingConfig()

	wasmApp := app.NewWasmApp(
		log.NewNopLogger(), db, nil, true,
		map[int64]bool{}, b.TempDir(), 0,
		encConfig, wasm.EnableAllProposals,
		sims.NewAppOptionsWithFlagHome(b.TempDir()), nil,
	)

	// Minter key used to sign benchmark transactions.
	privKey := secp256k1.GenPrivKey()
	minterAddr := sdk.AccAddress(privKey.PubKey().Address())

	denom := stakingtypes.DefaultParams().BondDenom

	genAccs := make([]authtypes.GenesisAccount, 0, numAccounts+1)
	balances := make([]banktypes.Balance, 0, numAccounts+1)

	genAccs = append(genAccs, authtypes.NewBaseAccountWithAddress(minterAddr))
	balances = append(balances, banktypes.Balance{
		Address: minterAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000_000_000)),
	})

	for i := 0; i < numAccounts; i++ {
		addr := sdk.AccAddress(append([]byte("bench_acct_"), byte(i)))
		genAccs = append(genAccs, authtypes.NewBaseAccountWithAddress(addr))
		balances = append(balances, banktypes.Balance{
			Address: addr.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000)),
		})
	}

	valSet, err := sims.CreateRandomValidatorSet()
	require.NoError(b, err)

	genesisState := app.NewDefaultGenesisState()
	genesisState, err = sims.GenesisStateWithValSet(
		encConfig.Marshaler,
		genesisState,
		valSet,
		genAccs,
		balances...,
	)
	require.NoError(b, err)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(b, err)

	valUpdates := cmttypes.TM2PB.ValidatorUpdates(valSet)
	_, err = wasmApp.InitChain(&abci.RequestInitChain{
		Validators:      valUpdates,
		AppStateBytes:   stateBytes,
		ConsensusParams: sims.DefaultConsensusParams,
	})
	require.NoError(b, err)

	_, err = wasmApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(b, err)

	_, err = wasmApp.Commit()
	require.NoError(b, err)

	return AppInfo{
		App:      wasmApp,
		TxConfig: encConfig.TxConfig,
		Denom:    denom,

		MinterAddr: minterAddr,
		PrivKey:    privKey,
		AccNum:     0,
		SeqNum:     1,
	}
}

// GenSequenceOfTxs builds numToCreate signed transactions using msgGen.
func GenSequenceOfTxs(b testing.TB, info *AppInfo, msgGen func(*AppInfo) ([]sdk.Msg, error), numToCreate int) []sdk.Tx {
	b.Helper()

	txs := make([]sdk.Tx, numToCreate)
	for i := 0; i < numToCreate; i++ {
		msgs, err := msgGen(info)
		require.NoError(b, err)

		txBuilder := info.TxConfig.NewTxBuilder()
		require.NoError(b, txBuilder.SetMsgs(msgs...))
		txBuilder.SetGasLimit(200_000)
		txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(info.Denom, 1_000)))

		seq := info.SeqNum + uint64(i)

		// Placeholder signature required before the real one can be generated.
		placeholder := sdktxsigning.SignatureV2{
			PubKey: info.PrivKey.PubKey(),
			Data: &sdktxsigning.SingleSignatureData{
				SignMode:  sdktxsigning.SignMode_SIGN_MODE_DIRECT,
				Signature: nil,
			},
			Sequence: seq,
		}
		require.NoError(b, txBuilder.SetSignatures(placeholder))

		signerData := authsigning.SignerData{
			ChainID:       "testing",
			AccountNumber: info.AccNum,
			Sequence:      seq,
		}
		sigV2, err := clienttx.SignWithPrivKey(
			context.Background(),
			sdktxsigning.SignMode_SIGN_MODE_DIRECT,
			signerData,
			txBuilder,
			info.PrivKey,
			info.TxConfig,
			seq,
		)
		require.NoError(b, err)
		require.NoError(b, txBuilder.SetSignatures(sigV2))

		txs[i] = txBuilder.GetTx()
	}
	return txs
}
