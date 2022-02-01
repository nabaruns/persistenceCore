/*
 Copyright [2019] - [2021], PERSISTENCE TECHNOLOGIES PTE. LTD. and the persistenceCore contributors
 SPDX-License-Identifier: Apache-2.0
*/

package application_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	simulation2 "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/persistenceOne/persistenceCore/application"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	dbm "github.com/tendermint/tm-db"
)

// SimAppChainID hardcoded chainID for simulation
const (
	SimAppChainID = "gaia-app"
)

func init() {
	simapp.GetSimulatorFlags()
}

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/GaiaApp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	config, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	defer func() {

		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			b.Fatal(err)
		}
	}()

	app := application.NewApplication().Initialize(application.Name, application.MakeEncodingConfig(), application.ModuleAccountPermissions, logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, application.DefaultNodeHome, simapp.EmptyAppOptions{}, interBlockCacheOpt())

	// Run randomized simulation:w
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		simapp.AppStateFn(app.ApplicationCodec(), app.SimulationManager()),
		simulation2.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(app, app.ApplicationCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.ApplicationCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = simapp.CheckExportSimulation(app, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		simapp.PrintStats(db)
	}
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping Application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := application.NewApplication().Initialize(application.Name, application.MakeEncodingConfig(), application.ModuleAccountPermissions, logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, application.DefaultNodeHome, simapp.EmptyAppOptions{}, interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				simapp.AppStateFn(app.ApplicationCodec(), app.SimulationManager()),
				simulation2.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simapp.SimulationOperations(app, app.ApplicationCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.ApplicationCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := app.BaseApp.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
