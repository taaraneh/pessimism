package e2e_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/base-org/pessimism/e2e"
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBalanceEnforcement ... Tests the E2E flow of a single
// balance enforcement heuristic session on L2 network.
func TestBalanceEnforcement(t *testing.T) {

	ts := e2e.CreateL2TestSuite(t)
	defer ts.Close()

	alice := ts.L2Cfg.Secrets.Addresses().Alice
	bob := ts.L2Cfg.Secrets.Addresses().Bob

	alertMsg := "one baby to another says:"
	// Deploy a balance enforcement heuristic session for Alice.
	_, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer2.String(),
		PType:         core.Live.String(),
		HeuristicType: core.BalanceEnforcement.String(),
		StartHeight:   nil,
		EndHeight:     nil,
		AlertingParams: &core.AlertPolicy{
			Sev: core.MEDIUM.String(),
			Msg: alertMsg,
		},
		SessionParams: map[string]interface{}{
			"address": alice.String(),
			"lower":   3, // i.e. alert if balance is less than 3 ETH
		},
	}})

	require.NoError(t, err, "Failed to bootstrap balance enforcement heuristic session")

	// Get Alice's balance.
	aliceAmt, err := ts.L2Geth.L2Client.BalanceAt(context.Background(), alice, nil)
	require.NoError(t, err, "Failed to get Alice's balance")

	// Determine the gas cost of the transaction.
	gasAmt := 1_000_001
	bigAmt := big.NewInt(1_000_001)
	gasPrice := big.NewInt(int64(ts.L2Cfg.DeployConfig.L2GenesisBlockGasLimit))

	gasCost := gasPrice.Mul(gasPrice, bigAmt)

	signer := types.LatestSigner(ts.L2Geth.L2ChainConfig)

	// Create a transaction from Alice to Bob that will drain almost all of Alice's ETH.
	drainAliceTx := types.MustSignNewTx(ts.L2Cfg.Secrets.Alice, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(ts.L2Cfg.DeployConfig.L2ChainID)),
		Nonce:     0,
		GasTipCap: big.NewInt(100),
		GasFeeCap: big.NewInt(100000),
		Gas:       uint64(gasAmt),
		To:        &bob,
		// Subtract the gas cost from the amount sent to Bob.
		Value: aliceAmt.Sub(aliceAmt, gasCost),
		Data:  nil,
	})

	require.Equal(t, len(ts.TestPagerDutyServer.PagerDutyAlerts()), 0, "No alerts should be sent before the transaction is sent")

	// Send the transaction to drain Alice's account of almost all ETH.
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainAliceTx)
	require.NoError(t, err, "Failed to create L2 block with transaction")

	// Wait for Pessimism to process the balance change and send a notification to the mocked Slack server.
	time.Sleep(1 * time.Second)

	// Check that the balance enforcement was triggered using the mocked server cache.
	pdMsgs := ts.TestPagerDutyServer.PagerDutyAlerts()
	slackMsgs := ts.TestSlackSvr.SlackAlerts()
	assert.Greater(t, len(slackMsgs), 1, "No balance enforcement alert was sent")
	assert.Greater(t, len(pdMsgs), 1, "No balance enforcement alert was sent")
	assert.Contains(t, pdMsgs[0].Payload.Summary, "balance_enforcement", "Balance enforcement alert was not sent")

	// Get Bobs's balance.
	bobAmt, err := ts.L2Geth.L2Client.BalanceAt(context.Background(), bob, nil)
	require.NoError(t, err, "Failed to get Alice's balance")

	// Create a transaction to send the ETH back to Alice.
	drainBobTx := types.MustSignNewTx(ts.L2Cfg.Secrets.Bob, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(ts.L2Cfg.DeployConfig.L2ChainID)),
		Nonce:     0,
		GasTipCap: big.NewInt(100),
		GasFeeCap: big.NewInt(100000),
		Gas:       uint64(gasAmt),
		To:        &alice,
		Value:     bobAmt.Sub(bobAmt, gasCost),
		Data:      nil,
	})

	// Send the transaction to re-disperse the ETH from Bob back to Alice.
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainBobTx)
	require.NoError(t, err, "Failed to create L2 block with transaction")

	// Wait for Pessimism to process the balance change.
	time.Sleep(1 * time.Second)

	// Empty the mocked PagerDuty server cache.
	ts.TestPagerDutyServer.ClearAlerts()

	// Wait to ensure that no new alerts are generated.
	time.Sleep(1 * time.Second)

	// Ensure that no new alerts were sent.
	assert.Equal(t, len(ts.TestPagerDutyServer.Payloads), 0, "No alerts should be sent after the transaction is sent")
}

// TestContractEvent ... Tests the E2E flow of a single
// contract event heuristic session on L1 network.
func TestContractEvent(t *testing.T) {

	ts := e2e.CreateSysTestSuite(t)
	defer ts.Close()

	// The string declaration of the event we want to listen for.
	updateSig := "ConfigUpdate(uint256,uint8,bytes)"
	alertMsg := "System config gas config updated"

	// Deploy a contract event heuristic session for the L1 system config address.
	ids, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer1.String(),
		PType:         core.Live.String(),
		HeuristicType: core.ContractEvent.String(),
		StartHeight:   nil,
		EndHeight:     nil,
		AlertingParams: &core.AlertPolicy{
			Msg: alertMsg,
			Sev: core.LOW.String(),
		},
		SessionParams: map[string]interface{}{
			"address": ts.Cfg.L1Deployments.SystemConfigProxy.String(),
			"args":    []interface{}{updateSig},
		},
	}})
	require.NoError(t, err, "Error bootstrapping heuristic session")

	// Get bindings for the L1 system config contract.
	sysCfg, err := bindings.NewSystemConfig(ts.Cfg.L1Deployments.SystemConfigProxy, ts.L1Client)
	require.NoError(t, err, "Error getting system config")

	// Obtain our signer.
	opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.SysCfgOwner, ts.Cfg.L1ChainIDBig())
	require.NoError(t, err, "Error getting system config owner pk")

	// Assign arbitrary gas config values.
	overhead := big.NewInt(10000)
	scalar := big.NewInt(1)

	// Call setGasConfig method on the L1 system config contract.
	tx, err := sysCfg.SetGasConfig(opts, overhead, scalar)
	require.NoError(t, err, "Error setting gas config")

	// Wait for the L1 transaction to be executed.
	receipt, err := wait.ForReceipt(context.Background(), ts.L1Client, tx.Hash(), types.ReceiptStatusSuccessful)

	require.NoError(t, err, "Error waiting for transaction")
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction failed")

	// Wait for Pessimism to process the newly emitted event and send a notification to the mocked Slack server.
	require.NoError(t, wait.For(context.Background(), 500*time.Millisecond, func() (bool, error) {
		pUUID := ids[0].PUUID
		height, err := ts.Subsystems.PipelineHeight(pUUID)
		if err != nil {
			return false, err
		}

		return height.Uint64() > receipt.BlockNumber.Uint64(), nil
	}))

	msgs := ts.TestSlackSvr.SlackAlerts()

	require.Equal(t, len(msgs), 1, "No system contract event alert was sent")
	assert.Contains(t, msgs[0].Text, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, msgs[0].Text, alertMsg, "System contract event message was not propagated")
}

// TestWithdrawalEnforcement ... Tests the E2E flow of a withdrawal
// / enforcement heuristic session. This test uses two L2ToL1 message passer contracts;
// one that is configured to be "faulty" and one that is not. The heuristic session
// should only produce an alert when the faulty L2ToL1 message passer is used given
// that it's state is empty.
func TestWithdrawalEnforcement(t *testing.T) {

	// 1 - Misconfigured testing environment
	// 2 - Change in chain state that affects the heuristic
	ts := e2e.CreateSysTestSuite(t)
	defer ts.Close()

	opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.Alice, ts.Cfg.L2ChainIDBig())
	require.NoError(t, err, "Error getting system config owner pk")

	alertMsg := "disrupting centralized finance"

	// Deploy a dummy L2ToL1 message passer for testing.
	fakeAddr, tx, _, err := bindings.DeployL2ToL1MessagePasser(opts, ts.L2Client)
	require.NoError(t, err, "error deploying dummy message passer on L2")

	_, err = wait.ForReceipt(context.Background(), ts.L2Client, tx.Hash(), types.ReceiptStatusSuccessful)
	require.NoError(t, err, "error waiting for transaction")

	// Setup Pessimism to listen for fraudulent withdrawals
	// We use two heuristics here; one configured with a dummy L1 message passer
	// and one configured with the real L2->L1 message passer contract. This allows us to
	// ensure that an alert is only produced using the faulty message passer since it's state
	// initiated withdrawal state is empty.
	ids, err := ts.App.BootStrap([]*models.SessionRequestParams{
		{
			// This is the one that should produce an alert
			Network:       core.Layer1.String(),
			PType:         core.Live.String(),
			HeuristicType: core.WithdrawalEnforcement.String(),
			StartHeight:   nil,
			EndHeight:     nil,
			AlertingParams: &core.AlertPolicy{
				Sev: core.LOW.String(),
				Msg: alertMsg,
			},
			SessionParams: map[string]interface{}{
				core.L1Portal:            ts.Cfg.L1Deployments.OptimismPortalProxy.String(),
				core.L2ToL1MessagePasser: fakeAddr.String(),
			},
		},
	})
	require.NoError(t, err, "Error bootstrapping heuristic session")

	optimismPortal, err := bindings.NewOptimismPortal(ts.Cfg.L1Deployments.OptimismPortalProxy, ts.L1Client)
	require.NoError(t, err)
	l2ToL1MessagePasser, err := bindings.NewL2ToL1MessagePasser(predeploys.L2ToL1MessagePasserAddr, ts.L2Client)
	require.NoError(t, err)

	aliceAddr := ts.Cfg.Secrets.Addresses().Alice

	// attach 1 ETH to the withdrawal and random calldata
	calldata := []byte{byte(1), byte(2), byte(3)}
	l2Opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.Alice, ts.Cfg.L2ChainIDBig())
	require.NoError(t, err)
	l2Opts.Value = big.NewInt(params.Ether)

	// Ensure L1 has enough funds for the withdrawal by depositing an equal amount into the OptimismPortal
	l1Opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.Alice, ts.Cfg.L1ChainIDBig())
	require.NoError(t, err)
	l1Opts.Value = l2Opts.Value
	depositTx, err := optimismPortal.Receive(l1Opts)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(context.Background(), ts.L1Client, depositTx.Hash())
	require.NoError(t, err)

	// Initiate and prove a withdrawal
	withdrawTx, err := l2ToL1MessagePasser.InitiateWithdrawal(l2Opts, aliceAddr, big.NewInt(100_000), calldata)
	require.NoError(t, err)
	withdrawReceipt, err := wait.ForReceiptOK(context.Background(), ts.L2Client, withdrawTx.Hash())
	require.NoError(t, err)

	_, proveReceipt := op_e2e.ProveWithdrawal(t, *ts.Cfg, ts.L1Client, ts.Sys.EthInstances["sequencer"], ts.Cfg.Secrets.Alice, withdrawReceipt)

	// Wait for Pessimism to process the withdrawal and send a notification to the mocked Slack server.
	require.NoError(t, wait.For(context.Background(), 500*time.Millisecond, func() (bool, error) {
		pUUID := ids[0].PUUID
		height, err := ts.Subsystems.PipelineHeight(pUUID)
		if err != nil {
			return false, err
		}

		return height.Uint64() > proveReceipt.BlockNumber.Uint64(), nil
	}))

	time.Sleep(time.Second * 10)

	// Ensure Pessimism has detected what it considers a "faulty" withdrawal
	alerts := ts.TestSlackSvr.SlackAlerts()
	require.Equal(t, 1, len(alerts), "expected 1 alert")
	assert.Contains(t, alerts[0].Text, "withdrawal_enforcement", "expected alert to be for withdrawal_enforcement")
	assert.Contains(t, alerts[0].Text, fakeAddr.String(), "expected alert to be for dummy L2ToL1MessagePasser")
	assert.Contains(t, alerts[0].Text, alertMsg, "expected alert to have alert message")
}

// TestFaultDetector ... Ensures that an alert is produced in the presence of a faulty L2Output root
// on the L1 Optimism portal contract.
func TestFaultDetector(t *testing.T) {
	ts := e2e.CreateSysTestSuite(t)
	defer ts.Close()

	// Generate transactor opts
	l1Opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.Proposer, ts.Cfg.L1ChainIDBig())
	require.Nil(t, err)

	// Generate output oracle bindings
	outputOracle, err := bindings.NewL2OutputOracleTransactor(ts.Cfg.L1Deployments.L2OutputOracleProxy, ts.L1Client)
	require.Nil(t, err)

	reader, err := bindings.NewL2OutputOracleCaller(ts.Cfg.L1Deployments.L2OutputOracleProxy, ts.L1Client)
	require.Nil(t, err)

	alertMsg := "the fault, dear Brutus, is not in our stars, but in ourselves"

	// Deploys a fault detector heuristic session instance using the locally spun-up Op-Stack chain
	ids, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer1.String(),
		PType:         core.Live.String(),
		HeuristicType: core.FaultDetector.String(),
		StartHeight:   big.NewInt(0),
		EndHeight:     nil,
		AlertingParams: &core.AlertPolicy{
			Sev: core.LOW.String(),
			Msg: alertMsg,
		},
		SessionParams: map[string]interface{}{
			core.L2OutputOracle:      ts.Cfg.L1Deployments.L2OutputOracleProxy.String(),
			core.L2ToL1MessagePasser: predeploys.L2ToL1MessagePasser,
		},
	}})

	require.Nil(t, err)
	require.Len(t, ids, 1)

	// Propose a forged L2 output root.

	dummyRoot := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	l1Hash := [32]byte{0}

	latestNum, err := reader.NextBlockNumber(&bind.CallOpts{})
	require.Nil(t, err)

	tx, err := outputOracle.ProposeL2Output(l1Opts, dummyRoot, latestNum, l1Hash, big.NewInt(0))
	require.Nil(t, err)

	receipt, err := wait.ForReceipt(context.Background(), ts.L1Client, tx.Hash(), types.ReceiptStatusSuccessful)
	require.Nil(t, err)

	require.NoError(t, wait.For(context.Background(), 500*time.Millisecond, func() (bool, error) {
		pUUID := ids[0].PUUID
		height, err := ts.Subsystems.PipelineHeight(pUUID)
		if err != nil {
			return false, err
		}

		return height.Uint64() > receipt.BlockNumber.Uint64(), nil
	}))

	alerts := ts.TestSlackSvr.SlackAlerts()
	require.Equal(t, 1, len(alerts), "expected 1 alert")
	assert.Contains(t, alerts[0].Text, "fault_detector", "expected alert to be for fault_detector")
	assert.Contains(t, alerts[0].Text, alertMsg, "expected alert to have alert message")
}
