package cmd_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/genesis-wrapper/cmd/wrapper/cmd"
)

func TestParseVestingAccounts(t *testing.T) {

	totalVestingAmt, _, vestingAccs := cmd.ParseVestingAccounts(cmd.VestingFilePathTest)
	// 100000000 * 2
	require.EqualValues(t, sdk.NewInt(200000000), totalVestingAmt)

	// vesting amt 100000000
	vestingAcc := vestingAccs[len(vestingAccs)-1]

	genesisTime := cmd.ParseTime(cmd.GenesisTime)
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(100000000))), vestingAcc.LockedCoins(genesisTime))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(100000000))), vestingAcc.GetVestingCoins(genesisTime))
	require.True(t, vestingAcc.GetVestedCoins(genesisTime).Empty())

	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(100000000))), vestingAcc.LockedCoins(genesisTime.Add(time.Hour*24*364)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(100000000))), vestingAcc.GetVestingCoins(genesisTime.Add(time.Hour*24*364)))
	require.True(t, vestingAcc.GetVestedCoins(genesisTime.Add(time.Hour*24*364)).Empty())

	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(65999988))), vestingAcc.LockedCoins(genesisTime.Add(time.Hour*24*365)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(65999988))), vestingAcc.GetVestingCoins(genesisTime.Add(time.Hour*24*365)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(34000012))), vestingAcc.GetVestedCoins(genesisTime.Add(time.Hour*24*365)))

	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(65999988-2833333))), vestingAcc.LockedCoins(genesisTime.Add(time.Hour*24*396)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(65999988-2833333))), vestingAcc.GetVestingCoins(genesisTime.Add(time.Hour*24*396)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(34000012+2833333))), vestingAcc.GetVestedCoins(genesisTime.Add(time.Hour*24*396)))

	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(2666666))), vestingAcc.LockedCoins(genesisTime.Add(time.Hour*24*365*3-1)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(2666666))), vestingAcc.GetVestingCoins(genesisTime.Add(time.Hour*24*365*3-1)))
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(100000000-2666666))), vestingAcc.GetVestedCoins(genesisTime.Add(time.Hour*24*365*3-1)))

	require.True(t, vestingAcc.LockedCoins(genesisTime.Add(time.Hour*24*365*3)).Empty())
	require.True(t, vestingAcc.GetVestingCoins(genesisTime.Add(time.Hour*24*365*3)).Empty())
	require.EqualValues(t, sdk.NewCoins(sdk.NewCoin(cmd.BondDenom, sdk.NewInt(100000000))), vestingAcc.GetVestedCoins(genesisTime.Add(time.Hour*24*365*3)))
}
