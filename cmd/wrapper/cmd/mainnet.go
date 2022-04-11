package cmd

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	claimtypes "github.com/crescent-network/crescent/x/claim/types"
	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
	liquidstakingtypes "github.com/crescent-network/crescent/x/liquidstaking/types"
	minttypes "github.com/crescent-network/crescent/x/mint/types"
	budgettypes "github.com/tendermint/budget/x/budget/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	filePath            = "./data/result.csv"              // airdrop result file
	VestingFilePath     = "./data/vesting.csv"             // vesting file
	VestingFilePathTest = "../../../data/vesting_test.csv" // vesting file
)

var (
	FarmingFeeCollector           = "cre1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mq4p6cjy"
	LiquidityFeeCollectorAddress  = "cre1zdew6yxyw92z373yqp756e0x4rvd2het37j0a2wjp7fj48eevxvq303p8d"
	LiquidityDustCollectorAddress = "cre1suads2mkd027cmfphmk9fpuwcct4d8ys02frk8e64hluswfwfj0s4xymnj"
	InflationFeeCollector         = "cre17xpfvakm2amg962yls6f84z3kell8c5l53s97s"
	EcosystemIncentive            = "cre1kgshua58cjr2p7hnrvgun68yrqf7ktdzyz2yxv54fqj6uwl4gc4q95txqa"
	EcosystemIncentiveLP          = "cre1wht0xhmuqph4rhzulhejgatthnpeatzjgnnkvqvphq97xr26np0qdvun2s"
	EcosystemIncentiveMM          = "cre1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sq3qhhde"
	EcosystemIncentiveBoost       = "cre17zftu6rg7mkmemqxv4whjkvecl0e2ja7j6um9t8qaczp79y72d7q2su2xm"
	AirdropSourceAddress          = "cre1rq9dzurree0ruj4xvuss33ysfus3lkneg3jnfdsy4ah8gxjta3mqlr2sax"
	FoundationAddress             = "cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e" // multisig
	DevTeamAddress                = "cre1ge2jm9nkvu2l8cvhc2un4m33d4yy4p0wfag09j" // multisig
)

var (
	GenesisTime      = "2022-04-13T00:00:00Z"
	GenesisTimeUnix  = ParseTime(GenesisTime).Unix()
	BondDenom        = "ucre"
	LiquidBondDenom  = "ubcre"
	FoundationSupply = sdk.NewInt(100_000_000_000_000) // 100mil
	DEXDropSupply    = sdk.NewInt(50_000_000_000_000)  // 50mil
	BoostDropSupply  = sdk.NewInt(50_000_000_000_000)  // 50mil
)

func MainnetGenesisStates() *GenesisStates {
	genParams := &GenesisStates{}
	genParams.BondDenom = BondDenom
	genParams.DEXdropSupply = sdk.NewCoin(genParams.BondDenom, DEXDropSupply)     // 50mil
	genParams.BoostdropSupply = sdk.NewCoin(genParams.BondDenom, BoostDropSupply) // 50mil

	// Set genesis time
	genParams.GenesisTime = ParseTime(GenesisTime)

	// Set consensus params
	genParams.ConsensusParams = &tmproto.ConsensusParams{
		Block: tmproto.BlockParams{
			MaxBytes:   10000000,
			MaxGas:     100000000,
			TimeIotaMs: 1000,
		},
		Evidence: tmproto.EvidenceParams{
			MaxAgeNumBlocks: 201600,
			MaxAgeDuration:  1209600000000000,
			MaxBytes:        1000000,
		},
		Validator: tmproto.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
		Version: tmproto.VersionParams{},
	}

	// Set auth params
	genParams.AuthParams = authtypes.DefaultParams()
	genParams.AuthParams.MaxMemoCharacters = 512

	// Set bank params
	genParams.BankParams = banktypes.DefaultParams()

	// Set crisis genesis states
	genParams.CrisisStates = crisistypes.GenesisState{
		ConstantFee: sdk.NewInt64Coin(genParams.BondDenom, 1000),
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
		MaxValidators:     50,
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
				StartTime: genParams.GenesisTime,
				EndTime:   genParams.GenesisTime.AddDate(1, 0, 0),
				Amount:    sdk.NewInt(108_700000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(1, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(2, 0, 0),
				Amount:    sdk.NewInt(216_100000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(2, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(3, 0, 0),
				Amount:    sdk.NewInt(151_300000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(3, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(4, 0, 0),
				Amount:    sdk.NewInt(105_900000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(4, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(5, 0, 0),
				Amount:    sdk.NewInt(74_100000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(5, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(6, 0, 0),
				Amount:    sdk.NewInt(51_900000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(6, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(7, 0, 0),
				Amount:    sdk.NewInt(36_300000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(7, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(8, 0, 0),
				Amount:    sdk.NewInt(25_400000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(8, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(9, 0, 0),
				Amount:    sdk.NewInt(17_800000_000000),
			},
			{
				StartTime: genParams.GenesisTime.AddDate(9, 0, 0),
				EndTime:   genParams.GenesisTime.AddDate(10, 0, 0),
				Amount:    sdk.NewInt(12_500000_000000),
			},
		},
	}

	// Set slashing params
	genParams.SlashingParams = slashingtypes.Params{
		SignedBlocksWindow:      30000,
		MinSignedPerWindow:      sdk.MustNewDecFromStr("0.050000000000000000"),
		DowntimeJailDuration:    60 * time.Second,
		SlashFractionDoubleSign: sdk.MustNewDecFromStr("0.050000000000000000"),
		SlashFractionDowntime:   sdk.MustNewDecFromStr("0.000000000000000000"),
	}

	// Set farming params
	genParams.FarmingParams = farmingtypes.Params{
		PrivatePlanCreationFee: sdk.NewCoins(sdk.NewInt64Coin(genParams.BondDenom, 100000000)),
		NextEpochDays:          1,
		FarmingFeeCollector:    FarmingFeeCollector,
		DelayedStakingGasFee:   sdk.Gas(100000),
		MaxNumPrivatePlans:     10000,
	}

	// Set liquidstaking params
	genParams.LiquidStakingParams = liquidstakingtypes.Params{
		LiquidBondDenom: LiquidBondDenom,
		WhitelistedValidators: []liquidstakingtypes.WhitelistedValidator{
			{
				ValidatorAddress: "crevaloper1n3mhyp9fvcmuu8l0q8qvjy07x0rql8q4ep74jz",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper17muws0zgrd0vzh37guea7960ym7aqf2j9v6l7s",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper1ls9w867xu0q5zjze5vrakfa2zluahtv44gwn7y",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper10rdgqczxyp69x9llq62cc3xs4w8w0k7p42x9jq",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper1dad8evf6vw72seljuzhjgurq48egaqfndvq38v",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper1zuucyy5v49lwnrdupqqafqdu29qy6wgnadwkuu",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper14lultfckehtszvzw4ehu0apvsr77afvy35naks",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper1qvdyzetkqq6rt4xu234xpvee5wt45a75rt2afe",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper18zvtvhzrqq5ny2jpmlc6new9k4c4uzzh6tcfpt",
				TargetWeight:     sdk.NewInt(10),
			},
			{
				ValidatorAddress: "crevaloper1pxexdsms050v35zu0vc07dk4ml647lsrjff52g",
				TargetWeight:     sdk.NewInt(10),
			},
		},
		UnstakeFeeRate:         sdk.MustNewDecFromStr("0.000000000000000000"),
		MinLiquidStakingAmount: sdk.NewInt(1000000),
	}

	// Set liquidity params
	genParams.LiquidityParams = liquiditytypes.Params{
		BatchSize:                1,
		TickPrecision:            3,
		FeeCollectorAddress:      LiquidityFeeCollectorAddress,
		DustCollectorAddress:     LiquidityDustCollectorAddress,
		MinInitialPoolCoinSupply: sdk.NewInt(1_000000_000000),
		PairCreationFee:          sdk.NewCoins(sdk.NewInt64Coin(genParams.BondDenom, 100_000_000)),
		PoolCreationFee:          sdk.NewCoins(sdk.NewInt64Coin(genParams.BondDenom, 100_000_000)),
		MinInitialDepositAmount:  sdk.NewInt(100000000),
		DepositExtraGas:          sdk.Gas(60000),
		WithdrawExtraGas:         sdk.Gas(64000),
		OrderExtraGas:            sdk.Gas(37000),
		MaxPriceLimitRatio:       sdk.MustNewDecFromStr("0.100000000000000000"),
		MaxOrderLifespan:         86400 * time.Second,
		SwapFeeRate:              sdk.MustNewDecFromStr("0.000000000000000000"),
		WithdrawFeeRate:          sdk.MustNewDecFromStr("0.000000000000000000"),
	}

	// Set gov params
	genParams.GovParams = govtypes.Params{
		DepositParams: govtypes.DepositParams{
			MinDeposit: sdk.NewCoins(
				sdk.NewInt64Coin(genParams.BondDenom, 500000000),
			),
			MaxDepositPeriod: 432000 * time.Second, // 5 days
		},
		VotingParams: govtypes.VotingParams{
			VotingPeriod: 432000 * time.Second,
		},
		TallyParams: govtypes.TallyParams{
			Quorum:        sdk.MustNewDecFromStr("0.400000000000000000"),
			Threshold:     sdk.MustNewDecFromStr("0.500000000000000000"),
			VetoThreshold: sdk.MustNewDecFromStr("0.334000000000000000"),
		},
	}

	// Set budget params
	genParams.BudgetParams = budgettypes.Params{
		EpochBlocks: 1,
		Budgets: []budgettypes.Budget{
			{
				Name:               "budget-ecosystem-incentive",
				Rate:               sdk.MustNewDecFromStr("0.662500000000000000"),
				SourceAddress:      InflationFeeCollector,
				DestinationAddress: EcosystemIncentive,
				StartTime:          genParams.GenesisTime,
				EndTime:            genParams.GenesisTime.AddDate(10, 0, 0),
			},
			{
				Name:               "budget-dev-team",
				Rate:               sdk.MustNewDecFromStr("0.250000000000000000"),
				SourceAddress:      InflationFeeCollector,
				DestinationAddress: DevTeamAddress,
				StartTime:          genParams.GenesisTime,
				EndTime:            genParams.GenesisTime.AddDate(10, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-lp-1",
				Rate:               sdk.MustNewDecFromStr("0.500000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveLP,
				StartTime:          genParams.GenesisTime,
				EndTime:            genParams.GenesisTime.AddDate(1, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-mm-1",
				Rate:               sdk.MustNewDecFromStr("0.300000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveMM,
				StartTime:          genParams.GenesisTime,
				EndTime:            genParams.GenesisTime.AddDate(1, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-boost-1",
				Rate:               sdk.MustNewDecFromStr("0.200000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveBoost,
				StartTime:          genParams.GenesisTime,
				EndTime:            genParams.GenesisTime.AddDate(1, 0, 0),
			},

			{
				Name:               "budget-ecosystem-incentive-lp-2",
				Rate:               sdk.MustNewDecFromStr("0.200000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveLP,
				StartTime:          genParams.GenesisTime.AddDate(1, 0, 0),
				EndTime:            genParams.GenesisTime.AddDate(2, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-mm-2",
				Rate:               sdk.MustNewDecFromStr("0.300000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveMM,
				StartTime:          genParams.GenesisTime.AddDate(1, 0, 0),
				EndTime:            genParams.GenesisTime.AddDate(2, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-boost-2",
				Rate:               sdk.MustNewDecFromStr("0.500000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveBoost,
				StartTime:          genParams.GenesisTime.AddDate(1, 0, 0),
				EndTime:            genParams.GenesisTime.AddDate(2, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-lp-3-10",
				Rate:               sdk.MustNewDecFromStr("0.100000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveLP,
				StartTime:          genParams.GenesisTime.AddDate(2, 0, 0),
				EndTime:            genParams.GenesisTime.AddDate(10, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-mm-3-10",
				Rate:               sdk.MustNewDecFromStr("0.300000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveMM,
				StartTime:          genParams.GenesisTime.AddDate(2, 0, 0),
				EndTime:            genParams.GenesisTime.AddDate(10, 0, 0),
			},
			{
				Name:               "budget-ecosystem-incentive-boost-3-10",
				Rate:               sdk.MustNewDecFromStr("0.600000000000000000"),
				SourceAddress:      EcosystemIncentive,
				DestinationAddress: EcosystemIncentiveBoost,
				StartTime:          genParams.GenesisTime.AddDate(2, 0, 0),
				EndTime:            genParams.GenesisTime.AddDate(10, 0, 0),
			},
		},
	}

	// Set claim genesis states
	airdrop := claimtypes.Airdrop{
		Id:            1,
		SourceAddress: AirdropSourceAddress, // airdrop source address
		Conditions: []claimtypes.ConditionType{
			claimtypes.ConditionTypeDeposit,
			claimtypes.ConditionTypeSwap,
			claimtypes.ConditionTypeLiquidStake,
			claimtypes.ConditionTypeVote,
		},
		StartTime: genParams.GenesisTime,
		EndTime:   genParams.GenesisTime.AddDate(0, 6, 0),
	}
	genParams.ClaimGenesisState.Airdrops = []claimtypes.Airdrop{airdrop}

	// Parse claim records, balances, and total initial genesis coin from the airdrop result file
	records, balances, totalInitialGenesisCoin := parseClaimRecords(genParams)

	// Deduct 20% initial airdrop amount
	dexDropSupply := genParams.DEXdropSupply.Sub(totalInitialGenesisCoin)

	// Set source account balances
	balances = append(balances, banktypes.Balance{
		Address: airdrop.SourceAddress,
		Coins:   sdk.NewCoins(dexDropSupply.Add(genParams.BoostdropSupply)), // DEXDropSupply + BoostDropSupply
	})

	// Add accounts
	newBalances, totalValidatorBalances := addValidatorBalances()
	balances = append(balances, newBalances...)

	// Sub validator amount from foundation
	FoundationSupply = FoundationSupply.Sub(totalValidatorBalances.AmountOf(BondDenom))
	// Parse and create vesting accounts info
	totalVestingAmt, vestingAccsMap, vestingAccs := ParseVestingAccounts(VestingFilePath)

	// Sub vesting amount from foundation
	FoundationSupply = FoundationSupply.Sub(totalVestingAmt)

	// Add Foundation balance
	balances = append(balances, banktypes.Balance{Address: FoundationAddress,
		Coins: sdk.NewCoins(sdk.NewCoin(genParams.BondDenom, FoundationSupply)), // 100mil - validator amount - vesting amount
	})

	// Add genesis accounts
	genAccounts := []authtypes.GenesisAccount{}
	vestingAccsBalances := []banktypes.Balance{}
	balancesMap := map[string]banktypes.Balance{}

	// Add Foundation as 1st account
	FoundationAcc, err := sdk.AccAddressFromBech32(FoundationAddress)
	if err != nil {
		panic(err)
	}
	genAccount := authtypes.NewBaseAccount(FoundationAcc, nil, 0, 0)
	genAccounts = append(genAccounts, genAccount)

	// Add Other accounts
	for i, balance := range balances {
		balancesMap[balance.Address] = balance

		// add vesting balance on existing account
		if vestingAcc, ok := vestingAccsMap[balance.GetAddress().String()]; ok {
			fmt.Println("added vesting balance on existing account", balance.GetAddress().String(), balances[i].Coins)
			balances[i].Coins = balances[i].Coins.Add(vestingAcc.OriginalVesting...)
		} else if balance.GetAddress().String() != FoundationAddress {
			// add genAccount except vesting accounts
			genAccount := authtypes.NewBaseAccount(balance.GetAddress(), nil, 0, 0)
			genAccounts = append(genAccounts, genAccount)
		}
	}

	// Add vesting accounts
	for _, vestingAcc := range vestingAccs {
		// add balance for new vesting accounts
		if _, ok := balancesMap[vestingAcc.GetAddress().String()]; !ok {
			vestingAccsBalances = append(vestingAccsBalances, banktypes.Balance{
				Address: vestingAcc.Address,
				Coins:   vestingAcc.OriginalVesting,
			})
		}
		genAccounts = append(genAccounts, vestingAcc)
	}
	balances = append(balances, vestingAccsBalances...)

	// Verify genesis accounts
	for _, genAccount := range genAccounts {
		if err := genAccount.Validate(); err != nil {
			panic(fmt.Sprintf("failed to validate genesis account: %s", err.Error()))
		}
	}

	genAccounts = authtypes.SanitizeGenesisAccounts(genAccounts)

	genAccs, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		panic(fmt.Errorf("failed to convert accounts into any's: %w", err))
	}
	genParams.AuthGenesisState.Accounts = genAccs

	// Set claim genesis states
	genParams.ClaimGenesisState.ClaimRecords = records
	genParams.BankGenesisStates.Balances = balances

	// Set supply genesis states
	// Total supply = DEXDropSupply + BoostDropSupply + Foundation + ValidatorBalances + TotalVestingAmount
	genParams.BankGenesisStates.Supply = sdk.NewCoins(
		genParams.DEXdropSupply.Add(genParams.BoostdropSupply)).
		Add(sdk.NewCoin(BondDenom, FoundationSupply)).
		Add(totalValidatorBalances...).Add(sdk.NewCoin(BondDenom, totalVestingAmt))

	fmt.Println("DEXdropSupply :", genParams.DEXdropSupply)
	fmt.Println("BoostdropSupply :", genParams.BoostdropSupply)
	fmt.Println("FoundationSupply :", FoundationSupply)
	fmt.Println("ValidatorBalances :", totalValidatorBalances)
	fmt.Println("totalVestingAmt :", totalVestingAmt)
	fmt.Println("len(vestingAccs) :", len(vestingAccs))
	fmt.Println("TotalSupply :", genParams.BankGenesisStates.Supply)
	return genParams
}

func addValidatorBalances() ([]banktypes.Balance, sdk.Coins) {
	balances := []banktypes.Balance{
		//{
		//	Address: "cre1n3mhyp9fvcmuu8l0q8qvjy07x0rql8q4m476lg", // already balance exist (airdrop recipient)
		//	Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		//},
		{
			Address: "cre17muws0zgrd0vzh37guea7960ym7aqf2j8c6sn6",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		{
			Address: "cre1ls9w867xu0q5zjze5vrakfa2zluahtv4huwunw",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		{
			Address: "cre10rdgqczxyp69x9llq62cc3xs4w8w0k7ph7x2l2",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		{
			Address: "cre1dad8evf6vw72seljuzhjgurq48egaqfn0cq72x",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		{
			Address: "cre1zuucyy5v49lwnrdupqqafqdu29qy6wgnlewe3k",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		//{
		//	Address: "cre14lultfckehtszvzw4ehu0apvsr77afvynqnjm6", // already balance exist (airdrop recipient)
		//	Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		//},
		{
			Address: "cre1qvdyzetkqq6rt4xu234xpvee5wt45a75pl2jyn",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		{
			Address: "cre18zvtvhzrqq5ny2jpmlc6new9k4c4uzzhclcxvp",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
		{
			Address: "cre1pxexdsms050v35zu0vc07dk4ml647lsrsafm8z",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1_000_000)),
		},
	}

	totalValidatorAmt := sdk.Coins{}
	for _, balance := range balances {
		totalValidatorAmt = totalValidatorAmt.Add(balance.Coins...)
	}

	return balances, totalValidatorAmt
}

func parseClaimRecords(genParams *GenesisStates) ([]claimtypes.ClaimRecord, []banktypes.Balance, sdk.Coin) {
	results, err := readCSVFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read csv file %s", filePath))
	}

	totalInitialGenesisAmt := sdk.ZeroInt()
	records := []claimtypes.ClaimRecord{}
	balances := []banktypes.Balance{}

	for i, r := range results {
		if i == 0 {
			continue
		}

		recipientAddr := r[0]
		dexClaimableAmt, ok := sdk.NewIntFromString(r[1])
		if !ok {
			panic("failed NewIntFromString for dexClaimableAmt")
		}

		// Convert bech32 address prefix
		_, converted, err := bech32.DecodeAndConvert(recipientAddr)
		if err != nil {
			panic(err)
		}

		targetPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
		recipientAddr, err = bech32.ConvertAndEncode(targetPrefix, converted)
		if err != nil {
			panic(err)
		}

		// Skip the zero amount
		if dexClaimableAmt.IsZero() {
			continue
		}

		initialGenesisAmt := dexClaimableAmt.Quo(sdk.NewInt(5))
		initialClaimableAmt := dexClaimableAmt.Sub(initialGenesisAmt)

		// 20% is set in genesis
		balance := banktypes.Balance{
			Address: recipientAddr,
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.BondDenom, initialGenesisAmt)),
		}
		balances = append(balances, balance)

		// 80% is set in claim record
		records = append(records, claimtypes.ClaimRecord{
			AirdropId:             1,
			Recipient:             recipientAddr,
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin(genParams.BondDenom, initialClaimableAmt)),
			ClaimableCoins:        sdk.NewCoins(sdk.NewCoin(genParams.BondDenom, initialClaimableAmt)),
		})

		// Track the total initial genesis amount
		totalInitialGenesisAmt = totalInitialGenesisAmt.Add(initialGenesisAmt)
	}

	totalInitialGenesisCoin := sdk.NewCoin(genParams.BondDenom, totalInitialGenesisAmt)

	return records, balances, totalInitialGenesisCoin
}

func ParseVestingAccounts(filePath string) (sdk.Int, map[string]*authvesting.PeriodicVestingAccount, []*authvesting.PeriodicVestingAccount) {
	vestingAccs := []*authvesting.PeriodicVestingAccount{}
	vestingAccMap := make(map[string]*authvesting.PeriodicVestingAccount)
	results, err := readCSVFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read csv file %s", filePath))
	}

	totalVestingAmt := sdk.ZeroInt()

	for i, r := range results {
		if i == 0 {
			continue
		}

		recipientAddr := r[0]
		vestingAmt, ok := sdk.NewIntFromString(r[1])
		if !ok {
			panic("failed NewIntFromString for vestingAmt")
		}

		// Convert bech32 address prefix
		_, converted, err := bech32.DecodeAndConvert(recipientAddr)
		if err != nil {
			panic(err)
		}

		targetPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
		recipientAddr, err = bech32.ConvertAndEncode(targetPrefix, converted)
		if err != nil {
			panic(err)
		}
		recipientAcc, err := sdk.AccAddressFromBech32(recipientAddr)
		if err != nil {
			panic(err)
		}

		// Skip the zero amount
		if vestingAmt.IsZero() {
			continue
		}

		baseAcc := authtypes.NewBaseAccount(recipientAcc, nil, 0, 0)
		periodVestingAcc := authvesting.NewPeriodicVestingAccount(baseAcc, sdk.NewCoins(sdk.NewCoin(BondDenom, vestingAmt)), GenesisTimeUnix, CalcVestingPeriod(vestingAmt))
		vestingAccMap[periodVestingAcc.Address] = periodVestingAcc
		vestingAccs = append(vestingAccs, periodVestingAcc)

		// Track the total vesting amount
		totalVestingAmt = totalVestingAmt.Add(vestingAmt)
	}
	return totalVestingAmt, vestingAccMap, vestingAccs
}

var (
	FirstYearCliff              = int64(60 * 60 * 24 * 365)                              // 31,536,000 1year
	SecondThirdYearMonthlyCliff = int64(60 * 60 * 24 * 365 / 12)                         // 2,628,000 1month
	TotalVestingLength          = FirstYearCliff + SecondThirdYearMonthlyCliff*int64(24) // 94,608,000 3year
	FirstYearRatio              = sdk.MustNewDecFromStr("0.34")
	SecondYearRatio             = sdk.MustNewDecFromStr("0.34")
	ThirdYearRatio              = sdk.MustNewDecFromStr("0.32")
	TotalCliff                  = 25
)

func CalcVestingPeriod(totalVestingAmount sdk.Int) authvesting.Periods {
	periods := authvesting.Periods{}

	firstYearVestingAmount := totalVestingAmount.ToDec().MulTruncate(FirstYearRatio).TruncateInt()
	secondYearMonthlyVestingAmount := totalVestingAmount.ToDec().MulTruncate(SecondYearRatio).QuoInt64(12).TruncateInt()
	thirdYearMonthlyVestingAmount := totalVestingAmount.ToDec().MulTruncate(ThirdYearRatio).QuoInt64(12).TruncateInt()
	crumb := totalVestingAmount.Sub(firstYearVestingAmount).Sub(secondYearMonthlyVestingAmount.MulRaw(12)).Sub(thirdYearMonthlyVestingAmount.MulRaw(12))
	firstYearVestingAmount = firstYearVestingAmount.Add(crumb)
	//firstYearVestingAmount := totalVestingAmount.Sub(secondYearMonthlyVestingAmount.MulRaw(12)).Sub(thirdYearMonthlyVestingAmount.MulRaw(12))

	// 1st year
	periods = append(periods, authvesting.Period{
		Length: FirstYearCliff,
		Amount: sdk.NewCoins(sdk.NewCoin(BondDenom, firstYearVestingAmount)),
	})

	// 2nd year
	for i := 0; i < 12; i++ {
		periods = append(periods, authvesting.Period{
			Length: SecondThirdYearMonthlyCliff,
			Amount: sdk.NewCoins(sdk.NewCoin(BondDenom, secondYearMonthlyVestingAmount)),
		})
	}

	// 3rd year
	for i := 0; i < 12; i++ {
		periods = append(periods, authvesting.Period{
			Length: SecondThirdYearMonthlyCliff,
			Amount: sdk.NewCoins(sdk.NewCoin(BondDenom, thirdYearMonthlyVestingAmount)),
		})
	}

	if len(periods) != TotalCliff {
		panic("error vesting periods number")
	}

	totalLength := int64(0)
	totalAmount := sdk.ZeroInt()
	for _, period := range periods {
		totalLength = totalLength + period.Length
		totalAmount = totalAmount.Add(period.Amount[0].Amount)
	}

	if totalLength != TotalVestingLength {
		panic("error total vesting length")
	}

	if !totalAmount.Equal(totalVestingAmount) {
		fmt.Println(totalAmount, totalVestingAmount)
		panic("error total vesting amount")
	}
	return periods
}
