package fjord

import (
	"context"
	"log/slog"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	"github.com/ethereum-optimism/optimism/devnet-sdk/types"
	fjordChecks "github.com/ethereum-optimism/optimism/op-chain-ops/cmd/check-fjord/checks"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestCheckFjordScript ensures the op-chain-ops/cmd/check-fjord script runs successfully
// against a test chain with the fjord hardfork activated/unactivated
func TestCheckFjordScript(t *testing.T) {

	l2ChainIndex := uint64(0)

	lowLevelSystemGetter, lowLevelSystemValidator := validators.AcquireLowLevelSystem()
	walletGetter, walletValidator := validators.AcquireL2WalletWithFunds(l2ChainIndex, types.NewBalance(big.NewInt(1_000_000)))
	forkConfigGetter, forkValidatorA := validators.AcquireL2WithFork(l2ChainIndex, rollup.Fjord)
	_, forkValidatorB := validators.AcquireL2WithoutFork(l2ChainIndex, rollup.Granite)
	systest.SystemTest(t,
		checkFjordScriptScenario(lowLevelSystemGetter, walletGetter, forkConfigGetter, l2ChainIndex),
		lowLevelSystemValidator,
		walletValidator,
		forkValidatorA,
		forkValidatorB,
	)

	forkConfigGetter, notForkValidator := validators.AcquireL2WithoutFork(l2ChainIndex, rollup.Fjord)
	systest.SystemTest(t,
		checkFjordScriptScenario(lowLevelSystemGetter, walletGetter, forkConfigGetter, l2ChainIndex),
		lowLevelSystemValidator,
		walletValidator,
		notForkValidator,
	)

}

func checkFjordScriptScenario(lowLevelSystemGetter validators.LowLevelSystemGetter, walletGetter validators.WalletGetter, chainConfigGetter validators.ChainConfigGetter, chainIndex uint64) systest.SystemTestFunc {
	return func(t systest.T, sys system.System) {
		llsys := lowLevelSystemGetter(t.Context())
		wallet := walletGetter(t.Context())
		chainConfig := chainConfigGetter(t.Context())

		l2 := sys.L2s()[chainIndex]
		l2LowLevelClient, err := llsys.L2s()[chainIndex].Client()
		require.NoError(t, err)

		// Get the wallet's private key and address
		privateKey := wallet.PrivateKey()
		walletAddr := wallet.Address()

		checkFjordConfig := &fjordChecks.CheckFjordConfig{
			Log:  log.NewLogger(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})),
			L2:   l2LowLevelClient,
			Key:  privateKey,
			Addr: walletAddr,
		}

		block, err := l2.Node().BlockByNumber(t.Context(), nil)
		require.NoError(t, err)
		time := block.Header().Time

		isFjordActivated, err := validators.IsForkActivated(chainConfig, rollup.Fjord, time)
		require.NoError(t, err)

		if !isFjordActivated {
			err = fjordChecks.CheckRIP7212(context.Background(), checkFjordConfig)
			require.Error(t, err, "expected error for CheckRIP7212")
			err = fjordChecks.CheckGasPriceOracle(context.Background(), checkFjordConfig)
			require.Error(t, err, "expected error for CheckGasPriceOracle")
			err = fjordChecks.CheckTxEmpty(context.Background(), checkFjordConfig)
			require.Error(t, err, "expected error for CheckTxEmpty")
			err = fjordChecks.CheckTxAllZero(context.Background(), checkFjordConfig)
			require.Error(t, err, "expected error for CheckTxAllZero")
			err = fjordChecks.CheckTxAll42(context.Background(), checkFjordConfig)
			require.Error(t, err, "expected error for CheckTxAll42")
			err = fjordChecks.CheckTxRandom(context.Background(), checkFjordConfig)
			require.Error(t, err, "expected error for CheckTxRandom")
		} else {
			err = fjordChecks.CheckAll(context.Background(), checkFjordConfig)
			require.NoError(t, err, "should not error on CheckAll")
		}
	}
}
