package keeper

import (
	"testing"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	"github.com/CosmWasm/wasmd/x/wasm/types"
)

func TestConstructorOptions(t *testing.T) {
	specs := map[string]struct {
		srcOpt Option
		verify func(*testing.T, Keeper)
	}{
		// NOTE: WithWasmEngine test skipped - wasmvm v2 uses concrete *wasmvm.VM type
		// and cannot accept mock WasmerEngine interfaces.
		"message handler": {
			srcOpt: WithMessageHandler(&wasmtesting.MockMessageHandler{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockMessageHandler{}, k.messenger)
			},
		},
		"query plugins": {
			srcOpt: WithQueryHandler(&wasmtesting.MockQueryHandler{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockQueryHandler{}, k.wasmVMQueryHandler)
			},
		},
		"message handler decorator": {
			srcOpt: WithMessageHandlerDecorator(func(old Messenger) Messenger {
				require.IsType(t, &MessageHandlerChain{}, old)
				return &wasmtesting.MockMessageHandler{}
			}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockMessageHandler{}, k.messenger)
			},
		},
		"query plugins decorator": {
			srcOpt: WithQueryHandlerDecorator(func(old WasmVMQueryHandler) WasmVMQueryHandler {
				require.IsType(t, QueryPlugins{}, old)
				return &wasmtesting.MockQueryHandler{}
			}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockQueryHandler{}, k.wasmVMQueryHandler)
			},
		},
		"coin transferrer": {
			srcOpt: WithCoinTransferrer(&wasmtesting.MockCoinTransferrer{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockCoinTransferrer{}, k.bank)
			},
		},
		"api costs": {
			srcOpt: WithAPICosts(1, 2),
			verify: func(t *testing.T, k Keeper) {
				t.Cleanup(setApiDefaults)
				assert.Equal(t, uint64(1), costHumanize)
				assert.Equal(t, uint64(2), costCanonical)
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			k := NewKeeper(nil, nil, paramtypes.NewSubspace(nil, nil, nil, nil, ""), nil, nil, nil, nil, nil, nil, nil, nil, nil, "tempDir", types.DefaultWasmConfig(), SupportedFeatures, spec.srcOpt)
			spec.verify(t, k)
		})
	}

}

func setApiDefaults() {
	costHumanize = DefaultGasCostHumanAddress * DefaultGasMultiplier
	costCanonical = DefaultGasCostCanonicalAddress * DefaultGasMultiplier
}
