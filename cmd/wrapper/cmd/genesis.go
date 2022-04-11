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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
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

type GenesisStates struct {
	DEXdropSupply   sdk.Coin
	BoostdropSupply sdk.Coin
	BondDenom       string

	GenesisTime         time.Time
	ChainId             string
	ConsensusParams     *tmproto.ConsensusParams
	AuthParams          authtypes.Params
	AuthGenesisState    authtypes.GenesisState
	BankParams          banktypes.Params
	DistributionParams  distrtypes.Params
	StakingParams       stakingtypes.Params
	GovParams           govtypes.Params
	SlashingParams      slashingtypes.Params
	MintParams          minttypes.Params
	LiquidityParams     liquiditytypes.Params
	LiquidStakingParams liquidstakingtypes.Params
	FarmingParams       farmingtypes.Params
	BudgetParams        budgettypes.Params
	BankGenesisStates   banktypes.GenesisState
	CrisisStates        crisistypes.GenesisState
	ClaimGenesisState   claimtypes.GenesisState
}

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
			genStates := parseNetworkType(networkType)
			if genStates == nil {
				return fmt.Errorf("you must choose between mainnet (m) or testnet (t): %s", args[0])
			}

			// Prepare genesis
			chainID := args[1]
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, genStates, chainID)
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
	genParams *GenesisStates,
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
	bankGenState.Params = genParams.BankParams
	bankGenStateBz := cdc.MustMarshalJSON(bankGenState)
	appState[banktypes.ModuleName] = bankGenStateBz

	// Auth module app state
	authGenState := authtypes.DefaultGenesisState()
	authGenState.Params = genParams.AuthParams
	authGenState.Accounts = genParams.AuthGenesisState.Accounts
	authGenStateBz := cdc.MustMarshalJSON(authGenState)
	appState[authtypes.ModuleName] = authGenStateBz

	// Crisis module app state
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = genParams.CrisisStates.ConstantFee
	crisisGenStateBz := cdc.MustMarshalJSON(crisisGenState)
	appState[crisistypes.ModuleName] = crisisGenStateBz

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

	// Slashing module app state
	slashingGenState := slashingtypes.DefaultGenesisState()
	slashingGenState.Params = genParams.SlashingParams
	slashingGenStateBz := cdc.MustMarshalJSON(slashingGenState)
	appState[slashingtypes.ModuleName] = slashingGenStateBz

	// Gov module app state
	govGenState := govtypes.DefaultGenesisState()
	govGenState.DepositParams = genParams.GovParams.DepositParams
	govGenState.VotingParams = genParams.GovParams.VotingParams
	govGenState.TallyParams = genParams.GovParams.TallyParams
	govGenStateBz := cdc.MustMarshalJSON(govGenState)
	appState[govtypes.ModuleName] = govGenStateBz

	// Farming module app state
	farmingGenState := farmingtypes.DefaultGenesisState()
	farmingGenState.Params = genParams.FarmingParams
	farmingGenStateBz := cdc.MustMarshalJSON(farmingGenState)
	appState[farmingtypes.ModuleName] = farmingGenStateBz

	// Budget module app state
	budgetGenState := budgettypes.DefaultGenesisState()
	budgetGenState.Params = genParams.BudgetParams
	budgetGenStatebz := cdc.MustMarshalJSON(budgetGenState)
	appState[budgettypes.ModuleName] = budgetGenStatebz

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

	// Claim module app state
	claimGenState := claimtypes.DefaultGenesis()
	claimGenState.Airdrops = genParams.ClaimGenesisState.Airdrops
	claimGenState.ClaimRecords = genParams.ClaimGenesisState.ClaimRecords
	claimGenStateBz := cdc.MustMarshalJSON(claimGenState)
	appState[claimtypes.ModuleName] = claimGenStateBz

	return appState, genDoc, nil
}

// parseNetworkType returns GenesisStates based on the network type.
func parseNetworkType(networkType string) *GenesisStates {
	switch strings.ToLower(networkType) {
	case "t", "testnet":
		// return TestnetGenesisStates()
		return &GenesisStates{}
	case "m", "mainnet":
		return MainnetGenesisStates()
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

// ParseTime parses and returns time.Time in time.RFC3339 format.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
