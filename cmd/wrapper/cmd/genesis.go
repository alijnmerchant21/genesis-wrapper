package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	claimtypes "github.com/crescent-network/crescent/x/claim/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
	liquidstakingtypes "github.com/crescent-network/crescent/x/liquidstaking/types"
)

var (
	// Airdrop result file
	filePath = "./data/result.csv"
)

func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prepare-genesis [network-type] [chain-id]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"pg"},
		Short:   "Prepare a genesis file with initial setup",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Prepare a genesis file with initial setup.

The initial setup includes initial params for Crescent

Example:
$ %s prepare-genesis mainnet crescent-1
$ %s prepare-genesis m crescent-1
$ %s prepare-genesis testnet mooncat-1
$ %s prepare-genesis t mooncat-1

The genesis output file is at $HOME/.crescent/config/genesis.json
`,
				version.AppName,
				version.AppName,
				version.AppName,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)

			serverCfg := serverCtx.Config

			// Read genesis file
			genFile := serverCfg.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// Parse genesis params depending on the network type
			networkType := args[0]
			genParams := parseGenesisParams(networkType)
			if genParams == nil {
				return fmt.Errorf("you must choose between mainnet (m) or testnet (t): %s", args[0])
			}

			// Prepare genesis
			chainID := args[1]
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, genParams, chainID)
			if err != nil {
				return fmt.Errorf("failed to prepare genesis %w", err)
			}

			// Validate genesis
			if err := mbm.ValidateGenesis(clientCtx.Codec, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("failed to validate genesis file: %w", err)
			}

			// Marshal and save the app state
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			// Export the genesis state to a file
			if err := genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return fmt.Errorf("failed to export genesis file %w", err)
			}

			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func PrepareGenesis(
	clientCtx client.Context,
	appState map[string]json.RawMessage,
	genDoc *tmtypes.GenesisDoc,
	genParams *GenesisParams,
	chainID string,
) (map[string]json.RawMessage, *tmtypes.GenesisDoc, error) {
	cdc := clientCtx.Codec

	genDoc.ChainID = chainID
	genDoc.GenesisTime = genParams.GenesisTime
	genDoc.ConsensusParams = genParams.ConsensusParams

	// Bank module app state
	bankGenState := banktypes.DefaultGenesisState()
	bankGenState.Balances = genParams.BankGenesisStates.Balances
	bankGenState.Supply = genParams.BankGenesisStates.Supply
	bankGenStateBz := cdc.MustMarshalJSON(bankGenState)
	appState[banktypes.ModuleName] = bankGenStateBz

	// Gov module app state
	govGenState := govtypes.DefaultGenesisState()
	govGenState.DepositParams = genParams.GovParams.DepositParams
	govGenState.VotingParams = genParams.GovParams.VotingParams
	govGenState.TallyParams = genParams.GovParams.TallyParams
	govGenStateBz := cdc.MustMarshalJSON(govGenState)
	appState[govtypes.ModuleName] = govGenStateBz

	// Liquidstaking module app state
	liquidstakingGenState := liquidstakingtypes.DefaultGenesisState()
	liquidstakingGenState.Params = genParams.LiquidStakingParams
	liquidstakingGenStateBz := cdc.MustMarshalJSON(liquidstakingGenState)
	appState[liquidstakingtypes.ModuleName] = liquidstakingGenStateBz

	// Claim module app state
	claimGenState := claimtypes.DefaultGenesis()
	claimGenState.Airdrops = genParams.ClaimGenesisState.Airdrops
	claimGenState.ClaimRecords = genParams.ClaimGenesisState.ClaimRecords
	claimGenStateBz := cdc.MustMarshalJSON(claimGenState)
	appState[claimtypes.ModuleName] = claimGenStateBz

	return appState, genDoc, nil
}

type GenesisParams struct {
	DEXdropSupply   sdk.Coin
	BoostdropSupply sdk.Coin

	GenesisTime     time.Time
	ChainId         string
	ConsensusParams *tmproto.ConsensusParams

	StakingParams       stakingtypes.Params
	GovParams           govtypes.Params
	LiquidityParams     liquiditytypes.Params
	LiquidStakingParams liquidstakingtypes.Params

	BankGenesisStates banktypes.GenesisState
	ClaimGenesisState claimtypes.GenesisState
}

func MainnetGenesisParams() *GenesisParams {
	genParams := &GenesisParams{}
	genParams.GenesisTime = parseTime("2022-04-14T00:00:00Z")
	genParams.DEXdropSupply = sdk.NewInt64Coin("cre", 50_000_000_000_000)   // 50 milion
	genParams.BoostdropSupply = sdk.NewInt64Coin("cre", 50_000_000_000_000) // 50 milion

	// Set source account balance and the total supply
	genParams.BankGenesisStates.Balances = []banktypes.Balance{
		{
			Address: "",
			Coins: sdk.NewCoins(
				genParams.DEXdropSupply.Add(genParams.BoostdropSupply),
			),
		},
	}
	genParams.BankGenesisStates.Supply = sdk.NewCoins(
		genParams.DEXdropSupply.Add(genParams.BoostdropSupply),
	)

	// Set airdrop
	genParams.ClaimGenesisState.Airdrops = []claimtypes.Airdrop{
		{
			Id:            1,
			SourceAddress: "",
			Conditions: []claimtypes.ConditionType{
				claimtypes.ConditionTypeDeposit,
				claimtypes.ConditionTypeSwap,
				claimtypes.ConditionTypeLiquidStake,
				claimtypes.ConditionTypeVote,
			},
			StartTime: genParams.GenesisTime,
			EndTime:   genParams.GenesisTime.AddDate(0, 1, 0),
		},
	}

	// Set claim records
	results, err := readCSVFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read %s", filePath))
	}

	records := []claimtypes.ClaimRecord{}
	for i, r := range results {
		if i == 0 {
			continue // remove header
		}

		recipientAddr := r[0]
		dexClaimableAmt, _ := sdk.NewIntFromString(r[1])

		if dexClaimableAmt.IsZero() {
			continue // skip the zero amount
		}

		records = append(records, claimtypes.ClaimRecord{
			AirdropId:             1,
			Recipient:             recipientAddr,
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, dexClaimableAmt)),
			ClaimableCoins:        sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, dexClaimableAmt)),
		})
	}
	genParams.ClaimGenesisState.ClaimRecords = records

	// Set active liquid validators
	genParams.LiquidStakingParams.LiquidBondDenom = "bcre"
	genParams.LiquidStakingParams.WhitelistedValidators = []liquidstakingtypes.WhitelistedValidator{
		{
			ValidatorAddress: "",
			TargetWeight:     sdk.NewInt(0),
		},
	}

	return genParams
}

func TestnetGenesisParams() *GenesisParams {
	genParams := &GenesisParams{}
	genParams.GenesisTime = time.Now()
	genParams.DEXdropSupply = sdk.NewInt64Coin("airdrop", 50_000_000_000_000)   // 50 milion
	genParams.BoostdropSupply = sdk.NewInt64Coin("airdrop", 50_000_000_000_000) // 50 milion

	// Set source account balance and the total supply
	genParams.BankGenesisStates.Balances = []banktypes.Balance{
		{
			Address: "cosmos15rz2rwnlgr7nf6eauz52usezffwrxc0mz4pywr",
			Coins: sdk.NewCoins(
				genParams.DEXdropSupply.Add(genParams.BoostdropSupply),
			),
		},
	}
	genParams.BankGenesisStates.Supply = sdk.NewCoins(
		genParams.DEXdropSupply.Add(genParams.BoostdropSupply),
	)

	// Set gov params
	genParams.GovParams.DepositParams = govtypes.DepositParams{
		MinDeposit:       sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)),
		MaxDepositPeriod: 172800 * time.Second,
	}

	// Set airdrop
	genParams.ClaimGenesisState.Airdrops = []claimtypes.Airdrop{
		{
			Id:            1,
			SourceAddress: "cosmos15rz2rwnlgr7nf6eauz52usezffwrxc0mz4pywr",
			Conditions: []claimtypes.ConditionType{
				claimtypes.ConditionTypeDeposit,
				claimtypes.ConditionTypeSwap,
				claimtypes.ConditionTypeLiquidStake,
				claimtypes.ConditionTypeVote,
			},
			StartTime: genParams.GenesisTime,
			EndTime:   genParams.GenesisTime.AddDate(0, 1, 0),
		},
	}

	// Set claim records
	results, err := readCSVFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read %s", filePath))
	}

	records := []claimtypes.ClaimRecord{}
	for i, r := range results {
		if i == 0 {
			continue // remove header
		}

		recipientAddr := r[0]
		dexClaimableAmt, _ := sdk.NewIntFromString(r[1])

		if dexClaimableAmt.IsZero() {
			continue // skip the zero amount
		}

		records = append(records, claimtypes.ClaimRecord{
			AirdropId:             1,
			Recipient:             recipientAddr,
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, dexClaimableAmt)),
			ClaimableCoins:        sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, dexClaimableAmt)),
		})
	}

	// (Testing) Set custom claim records
	records = append(records, claimtypes.ClaimRecord{
		AirdropId:             1,
		Recipient:             "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v", // alice
		InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, sdk.NewInt(5_000_000_000_000))),
		ClaimableCoins:        sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, sdk.NewInt(5_000_000_000_000))),
		ClaimedConditions:     []claimtypes.ConditionType{},
	})
	genParams.ClaimGenesisState.ClaimRecords = records

	// Set active liquid validators
	genParams.LiquidStakingParams.LiquidBondDenom = "bstake"
	genParams.LiquidStakingParams.WhitelistedValidators = []liquidstakingtypes.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper1zaavvzxez0elundtn32qnk9lkm8kmcsz8ycjrl", // alice operator address
			TargetWeight:     sdk.NewInt(1_000_000_000),
		},
	}

	return genParams
}

// parseGenesisParams returns GenesisParams based on the network type.
func parseGenesisParams(networkType string) *GenesisParams {
	switch strings.ToLower(networkType) {
	case "t", "testnet":
		return TestnetGenesisParams()
	case "m", "mainnet":
		return MainnetGenesisParams()
	default:
		return nil
	}
}

// readCSVFile reads csv file and returns all the records.
func readCSVFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// parseTime parses and returns time.Time in time.RFC3339 format.
func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
