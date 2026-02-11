package app

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// UpgradeName defines the on-chain upgrade name for the SDK 0.50 migration.
// This must match the name used in the governance upgrade proposal.
const UpgradeName = "v2-sdk50"

// RegisterUpgradeHandlers registers the upgrade handler that migrates state
// from Cosmos SDK 0.47.x to 0.50.x. This includes:
// - Migrating consensus params from the legacy x/params subspace to the new
//   x/consensus module collections store.
// - Running all registered module migrations (e.g., mint v1→v2, staking v4→v5)
//   which migrate module params from x/params subspaces to their respective
//   module collections stores.
func (app *WasmApp) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)

			// Step 1: Migrate consensus params from the legacy x/params "baseapp"
			// subspace to the new x/consensus module's collections store.
			// In SDK 0.47, consensus params were stored in the x/params module
			// under the "baseapp" subspace. SDK 0.50 stores them in the
			// x/consensus module using collections.
			baseAppLegacySS := app.getSubspace(baseapp.Paramspace)
			if err := baseapp.MigrateParams(sdkCtx, baseAppLegacySS, app.consensusKeeper.ParamsStore); err != nil {
				return nil, err
			}

			sdkCtx.Logger().Info("migrated consensus params from x/params to x/consensus module")

			// Step 2: Run all module migrations. This executes registered
			// migrations for each module (e.g., mint Migrate1to2 which moves
			// params from x/params subspace to module collections storage).
			// The fromVM contains the old module versions from the SDK 0.47
			// database; RunMigrations compares them against current versions
			// and runs the necessary migration functions.
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	// Configure the store loader to add new store keys that didn't exist in
	// SDK 0.47. The "Consensus" and "crisis" store keys are new in SDK 0.50.
	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{consensuskeeper.StoreKey, crisistypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
