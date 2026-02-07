package keeper

import (
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"

	"github.com/CosmWasm/wasmd/x/wasm/types"
)

// bindIbcPort will reserve the port.
// returns a string name of the port or error if we cannot bind it.
// this will fail if call twice.
func (k Keeper) bindIbcPort(ctx sdk.Context, portID string) error {
	return k.portKeeper.BindPort(ctx, portID)
}

// ensureIbcPort is like registerIbcPort, but it checks if we already hold the port
// before calling register, so this is safe to call multiple times.
// Returns success if we already registered or just registered and error if we cannot
// (lack of permissions or someone else has it)
func (k Keeper) ensureIbcPort(ctx sdk.Context, contractAddr sdk.AccAddress) (string, error) {
	portID := PortIDForContract(contractAddr)
	if err := k.bindIbcPort(ctx, portID); err != nil {
		if errors.Is(err, porttypes.ErrPortAlreadyBound) {
			return portID, nil
		}
		return portID, err
	}
	return portID, nil
}

const portIDPrefix = "wasm."

func PortIDForContract(addr sdk.AccAddress) string {
	return portIDPrefix + addr.String()
}

func ContractFromPortID(portID string) (sdk.AccAddress, error) {
	if !strings.HasPrefix(portID, portIDPrefix) {
		return nil, sdkerrors.Wrapf(types.ErrInvalid, "without prefix")
	}
	return sdk.AccAddressFromBech32(portID[len(portIDPrefix):])
}

// Capability-based IBC methods were removed in IBC v10.
