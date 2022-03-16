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
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	budgettypes "github.com/tendermint/budget/x/budget/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	claimtypes "github.com/crescent-network/crescent/x/claim/types"
	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
	liquidstakingtypes "github.com/crescent-network/crescent/x/liquidstaking/types"
	minttypes "github.com/crescent-network/crescent/x/mint/types"
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
$ %s prepare-genesis testnet mooncat-1-1
$ %s prepare-genesis t mooncat-1-1

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

	// Distribution module app state
	distrGenState := distrtypes.DefaultGenesisState()
	distrGenState.Params = genParams.DistributionParams
	distrGenStateBz := cdc.MustMarshalJSON(distrGenState)
	appState[distrtypes.ModuleName] = distrGenStateBz

	// Staking module app state
	stakingGenState := stakingtypes.DefaultGenesisState()
	stakingGenState.Params = genParams.StakingParams
	stakingGenStateBz := cdc.MustMarshalJSON(stakingGenState)
	appState[stakingtypes.ModuleName] = stakingGenStateBz

	// Mint module app state
	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params = genParams.MintParams
	mintGenStateBz := cdc.MustMarshalJSON(mintGenState)
	appState[minttypes.ModuleName] = mintGenStateBz

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

	// Liquidity module app state
	liquidityGenState := liquiditytypes.DefaultGenesis()
	liquidityGenState.Params = genParams.LiquidityParams
	liquidityGenStateBz := cdc.MustMarshalJSON(liquidityGenState)
	appState[liquiditytypes.ModuleName] = liquidityGenStateBz

	// Farming module app state
	farmingGenState := farmingtypes.DefaultGenesisState()
	farmingGenState.Params = genParams.FarmingParams
	farmingGenStateBz := cdc.MustMarshalJSON(farmingGenState)
	appState[farmingtypes.ModuleName] = farmingGenStateBz

	// Claim module app state
	claimGenState := claimtypes.DefaultGenesis()
	claimGenState.Airdrops = genParams.ClaimGenesisState.Airdrops
	claimGenState.ClaimRecords = genParams.ClaimGenesisState.ClaimRecords
	claimGenStateBz := cdc.MustMarshalJSON(claimGenState)
	appState[claimtypes.ModuleName] = claimGenStateBz

	// Budget module app state
	budgetGenState := budgettypes.DefaultGenesisState()
	budgetGenState.Params = genParams.BudgetParams
	budgetGenStatebz := cdc.MustMarshalJSON(budgetGenState)
	appState[budgettypes.ModuleName] = budgetGenStatebz

	return appState, genDoc, nil
}

type GenesisParams struct {
	DEXdropSupply   sdk.Coin
	BoostdropSupply sdk.Coin
	BondDenom       string

	GenesisTime     time.Time
	ChainId         string
	ConsensusParams *tmproto.ConsensusParams

	BankParams          banktypes.Params
	DistributionParams  distrtypes.Params
	StakingParams       stakingtypes.Params
	GovParams           govtypes.Params
	MintParams          minttypes.Params
	LiquidityParams     liquiditytypes.Params
	LiquidStakingParams liquidstakingtypes.Params
	FarmingParams       farmingtypes.Params
	BudgetParams        budgettypes.Params

	BankGenesisStates banktypes.GenesisState
	ClaimGenesisState claimtypes.GenesisState
}

func TestnetGenesisParams() *GenesisParams {
	genParams := &GenesisParams{}
	genParams.BondDenom = "ucre"
	genParams.DEXdropSupply = sdk.NewInt64Coin(genParams.BondDenom, 50_000_000_000_000)   // 50mil
	genParams.BoostdropSupply = sdk.NewInt64Coin(genParams.BondDenom, 50_000_000_000_000) // 50mil

	// Set genesis time
	genParams.GenesisTime = parseTime("2022-03-17T00:00:00Z")

	// Set consensus params
	genParams.ConsensusParams = &tmproto.ConsensusParams{
		Block: tmproto.BlockParams{
			MaxBytes: 10000000,
			MaxGas:   100000000,
		},
		Evidence: tmproto.EvidenceParams{
			MaxAgeNumBlocks: 403200,
			MaxAgeDuration:  1209600000000000,
			MaxBytes:        1000000,
		},
	}

	// Set distribution params
	genParams.DistributionParams = distrtypes.Params{
		CommunityTax:        sdk.MustNewDecFromStr("0.285714285700000000"),
		BaseProposerReward:  sdk.MustNewDecFromStr("0.007142857143000000"),
		BonusProposerReward: sdk.MustNewDecFromStr("0.028571428570000000"),
		WithdrawAddrEnabled: true,
	}

	// Set staking params
	genParams.StakingParams = stakingtypes.Params{
		UnbondingTime:     1209600 * time.Second, // 2 weeks
		MaxValidators:     20,
		MaxEntries:        28,
		HistoricalEntries: 10000,
		BondDenom:         genParams.BondDenom,
	}

	// Set mint params
	genParams.MintParams = minttypes.Params{
		MintDenom:          genParams.BondDenom,
		BlockTimeThreshold: 10 * time.Second,
		InflationSchedules: []minttypes.InflationSchedule{
			{
				StartTime: parseTime(""),
				EndTime:   parseTime(""),
				Amount:    sdk.NewInt(0),
			},
		},
	}

	// Set liquidstaking params
	genParams.LiquidStakingParams = liquidstakingtypes.Params{
		LiquidBondDenom: "ubcre",
		WhitelistedValidators: []liquidstakingtypes.WhitelistedValidator{
			{
				ValidatorAddress: "cosmosvaloper1zaavvzxez0elundtn32qnk9lkm8kmcsz8ycjrl", // alice operator address
				TargetWeight:     sdk.NewInt(1_000_000_000),
			},
		},
		UnstakeFeeRate:         sdk.MustNewDecFromStr("0.001000000000000000"),
		MinLiquidStakingAmount: sdk.NewInt(0),
	}

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
			EndTime:   genParams.GenesisTime.AddDate(0, 6, 0),
		},
	}

	// Set claim records
	records, balances, totalInitialGenesisCoin := GetClaimRecords(genParams)
	genParams.ClaimGenesisState.ClaimRecords = records

	// Set source account balance and the total supply
	balances = append(balances, banktypes.Balance{
		Address: "", // source address
		Coins: sdk.NewCoins(
			genParams.DEXdropSupply.Add(genParams.BoostdropSupply).Add(totalInitialGenesisCoin),
		),
	})
	genParams.BankGenesisStates.Balances = balances
	genParams.BankGenesisStates.Supply = sdk.NewCoins(
		genParams.DEXdropSupply.Add(genParams.BoostdropSupply).Add(totalInitialGenesisCoin),
	)
	genParams.BankParams = banktypes.Params{
		SendEnabled: []*banktypes.SendEnabled{
			{
				Denom:   genParams.BondDenom,
				Enabled: true,
			},
		},
		DefaultSendEnabled: true,
	}

	// Set farming params
	genParams.FarmingParams.PrivatePlanCreationFee = sdk.NewCoins(sdk.NewInt64Coin(genParams.BondDenom, 100000000))

	// Set liquidity params
	genParams.LiquidityParams = liquiditytypes.Params{
		BatchSize:                1,
		TickPrecision:            3,
		FeeCollectorAddress:      "cre1zdew6yxyw92z373yqp756e0x4rvd2het37j0a2wjp7fj48eevxvq303p8d",
		DustCollectorAddress:     "cre1suads2mkd027cmfphmk9fpuwcct4d8ys02frk8e64hluswfwfj0s4xymnj",
		PairCreationFee:          sdk.NewCoins(sdk.NewInt64Coin(genParams.BondDenom, 100000000)),
		PoolCreationFee:          sdk.NewCoins(sdk.NewInt64Coin(genParams.BondDenom, 100000000)),
		MinInitialPoolCoinSupply: sdk.NewInt(1000000000000),
		MaxPriceLimitRatio:       sdk.MustNewDecFromStr("0.100000000000000000"),
		MaxOrderLifespan:         86400 * time.Second,
		SwapFeeRate:              sdk.MustNewDecFromStr("0"),
		WithdrawFeeRate:          sdk.MustNewDecFromStr("0"),
	}

	// Set gov params
	genParams.GovParams = govtypes.Params{
		DepositParams: govtypes.DepositParams{
			MinDeposit: sdk.NewCoins(
				sdk.NewInt64Coin(genParams.BondDenom, 0),
			),
		},
		VotingParams: govtypes.VotingParams{
			VotingPeriod: 17200 * time.Second,
		},
		TallyParams: govtypes.TallyParams{
			Quorum: sdk.MustNewDecFromStr(""),
		},
	}

	return genParams
}

func MainnetGenesisParams() *GenesisParams {
	genParams := &GenesisParams{}
	genParams.GenesisTime = parseTime("2022-04-14T00:00:00Z")

	// TODO: TBD

	return genParams
}

func GetClaimRecords(genParams *GenesisParams) ([]claimtypes.ClaimRecord, []banktypes.Balance, sdk.Coin) {
	results, err := readCSVFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read %s", filePath))
	}

	totalInitialGenesisAmt := sdk.ZeroInt()
	balances := []banktypes.Balance{}
	records := []claimtypes.ClaimRecord{}

	// Loop through each result and get
	for i, r := range results {
		// Remove header
		if i == 0 {
			continue
		}

		recipientAddr := r[0]
		dexClaimableAmt, _ := sdk.NewIntFromString(r[1])

		// Skip the zero amount
		if dexClaimableAmt.IsZero() {
			continue
		}

		// Out of the total claimable amount, 20% is set in genesis and
		// the rest 80% is set in their claim record
		initialGenesisAmt := dexClaimableAmt.Quo(sdk.NewInt(5))
		initialClaimableAmt := dexClaimableAmt.Sub(initialGenesisAmt)
		totalInitialGenesisAmt = totalInitialGenesisAmt.Add(initialGenesisAmt)

		balances = append(balances, banktypes.Balance{
			Address: recipientAddr,
			Coins:   sdk.NewCoins(sdk.NewCoin("ucre", initialClaimableAmt)),
		})

		records = append(records, claimtypes.ClaimRecord{
			AirdropId:             1,
			Recipient:             recipientAddr,
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, initialClaimableAmt)),
			ClaimableCoins:        sdk.NewCoins(sdk.NewCoin(genParams.DEXdropSupply.Denom, initialClaimableAmt)),
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

	return records, balances, sdk.NewCoin("ucre", totalInitialGenesisAmt)
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
