package keeper_test

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cryptotestutil "github.com/cosmos/interchain-security/v2/testutil/crypto"
	testkeeper "github.com/cosmos/interchain-security/v2/testutil/keeper"
	"github.com/cosmos/interchain-security/v2/x/ccv/provider/types"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
	"testing"
	"time"
)

// TestJailAndTombstoneValidator tests that the jailing of a validator is only executed
// under the conditions that the validator is neither unbonded, nor jailed, nor tombstoned.
func TestJailAndTombstoneValidator(t *testing.T) {
	providerConsAddr := cryptotestutil.NewCryptoIdentityFromIntSeed(7842334).ProviderConsAddress()
	testCases := []struct {
		name          string
		provAddr      types.ProviderConsAddress
		expectedCalls func(sdk.Context, testkeeper.MockedKeepers, types.ProviderConsAddress) []*gomock.Call
	}{
		{
			"unfound validator",
			providerConsAddr,
			func(ctx sdk.Context, mocks testkeeper.MockedKeepers,
				provAddr types.ProviderConsAddress,
			) []*gomock.Call {
				return []*gomock.Call{
					// We only expect a single call to GetValidatorByConsAddr.
					// Method will return once validator is not found.
					mocks.MockStakingKeeper.EXPECT().GetValidatorByConsAddr(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						stakingtypes.Validator{}, false, // false = Not found.
					).Times(1),
				}
			},
		},
		{
			"unbonded validator",
			providerConsAddr,
			func(ctx sdk.Context, mocks testkeeper.MockedKeepers,
				provAddr types.ProviderConsAddress,
			) []*gomock.Call {
				return []*gomock.Call{
					// We only expect a single call to GetValidatorByConsAddr.
					mocks.MockStakingKeeper.EXPECT().GetValidatorByConsAddr(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						stakingtypes.Validator{Status: stakingtypes.Unbonded}, true,
					).Times(1),
				}
			},
		},
		{
			"tombstoned validator",
			providerConsAddr,
			func(ctx sdk.Context, mocks testkeeper.MockedKeepers,
				provAddr types.ProviderConsAddress,
			) []*gomock.Call {
				return []*gomock.Call{
					mocks.MockStakingKeeper.EXPECT().GetValidatorByConsAddr(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						stakingtypes.Validator{}, true,
					).Times(1),
					mocks.MockSlashingKeeper.EXPECT().IsTombstoned(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						true,
					).Times(1),
				}
			},
		},
		{
			"jailed validator",
			providerConsAddr,
			func(ctx sdk.Context, mocks testkeeper.MockedKeepers,
				provAddr types.ProviderConsAddress,
			) []*gomock.Call {
				return []*gomock.Call{
					mocks.MockStakingKeeper.EXPECT().GetValidatorByConsAddr(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						stakingtypes.Validator{Jailed: true}, true,
					).Times(1),
					mocks.MockSlashingKeeper.EXPECT().IsTombstoned(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						false,
					).Times(1),
					mocks.MockSlashingKeeper.EXPECT().JailUntil(
						ctx, providerConsAddr.ToSdkConsAddr(), evidencetypes.DoubleSignJailEndTime).
						Times(1),
					mocks.MockSlashingKeeper.EXPECT().Tombstone(
						ctx, providerConsAddr.ToSdkConsAddr()).
						Times(1),
				}
			},
		},
		{
			"bonded validator",
			providerConsAddr,
			func(ctx sdk.Context, mocks testkeeper.MockedKeepers,
				provAddr types.ProviderConsAddress,
			) []*gomock.Call {
				return []*gomock.Call{
					mocks.MockStakingKeeper.EXPECT().GetValidatorByConsAddr(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						stakingtypes.Validator{Status: stakingtypes.Bonded}, true,
					).Times(1),
					mocks.MockSlashingKeeper.EXPECT().IsTombstoned(
						ctx, providerConsAddr.ToSdkConsAddr()).Return(
						false,
					).Times(1),
					mocks.MockStakingKeeper.EXPECT().Jail(
						ctx, providerConsAddr.ToSdkConsAddr()).
						Times(1),
					mocks.MockSlashingKeeper.EXPECT().JailUntil(
						ctx, providerConsAddr.ToSdkConsAddr(), evidencetypes.DoubleSignJailEndTime).
						Times(1),
					mocks.MockSlashingKeeper.EXPECT().Tombstone(
						ctx, providerConsAddr.ToSdkConsAddr()).
						Times(1),
				}
			},
		},
	}

	for _, tc := range testCases {
		providerKeeper, ctx, ctrl, mocks := testkeeper.GetProviderKeeperAndCtx(
			t, testkeeper.NewInMemKeeperParams(t))

		// Setup expected mock calls
		gomock.InOrder(tc.expectedCalls(ctx, mocks, tc.provAddr)...)

		// Execute method and assert expected mock calls
		providerKeeper.JailAndTombstoneValidator(ctx, tc.provAddr)

		ctrl.Finish()
	}
}

// createUndelegation creates an undelegation with `len(initialBalances)` entries
func createUndelegation(initialBalances []int64, completionTimes []time.Time) stakingtypes.UnbondingDelegation {
	var entries []stakingtypes.UnbondingDelegationEntry
	for i, balance := range initialBalances {
		entry := stakingtypes.UnbondingDelegationEntry{
			InitialBalance: sdk.NewInt(balance),
			CompletionTime: completionTimes[i],
		}
		entries = append(entries, entry)
	}

	return stakingtypes.UnbondingDelegation{Entries: entries}
}

// createRedelegation creates a redelegation with `len(initialBalances)` entries
func createRedelegation(initialBalances []int64, completionTimes []time.Time) stakingtypes.Redelegation {
	var entries []stakingtypes.RedelegationEntry
	for i, balance := range initialBalances {
		entry := stakingtypes.RedelegationEntry{
			InitialBalance: sdk.NewInt(balance),
			CompletionTime: completionTimes[i],
		}
		entries = append(entries, entry)
	}

	return stakingtypes.Redelegation{Entries: entries}
}

// TestComputePowerToSlash tests that `ComputePowerToSlash` computes the correct power to be slashed based on
// the tokens in non-mature undelegation and redelegation entries, as well as the current power of the validator
func TestComputePowerToSlash(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	// undelegation or redelegation entries with completion time `now` have matured
	now := ctx.BlockHeader().Time
	// undelegation or redelegation entries with completion time one hour in the future have not yet matured
	nowPlus1Hour := now.Add(time.Hour)

	testCases := []struct {
		name           string
		undelegations  []stakingtypes.UnbondingDelegation
		redelegations  []stakingtypes.Redelegation
		power          int64
		powerReduction sdk.Int
		expectedPower  int64
	}{
		{
			"both undelegations and redelegations 1",
			// 1000 total undelegation tokens
			[]stakingtypes.UnbondingDelegation{
				createUndelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createUndelegation([]int64{500}, []time.Time{nowPlus1Hour, nowPlus1Hour})},
			// 1000 total redelegation tokens
			[]stakingtypes.Redelegation{
				createRedelegation([]int64{500}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createRedelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
			},
			int64(1000),
			sdk.NewInt(1),
			int64(2000/1 + 1000),
		},
		{
			"both undelegations and redelegations 2",
			// 2000 total undelegation tokens
			[]stakingtypes.UnbondingDelegation{
				createUndelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createUndelegation([]int64{}, []time.Time{}),
				createUndelegation([]int64{100, 100}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createUndelegation([]int64{800}, []time.Time{nowPlus1Hour}),
				createUndelegation([]int64{500}, []time.Time{nowPlus1Hour})},
			// 3500 total redelegation tokens
			[]stakingtypes.Redelegation{
				createRedelegation([]int64{}, []time.Time{}),
				createRedelegation([]int64{1600}, []time.Time{nowPlus1Hour}),
				createRedelegation([]int64{350, 250}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createRedelegation([]int64{700, 200}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createRedelegation([]int64{}, []time.Time{}),
				createRedelegation([]int64{400}, []time.Time{nowPlus1Hour}),
			},
			int64(8391),
			sdk.NewInt(2),
			int64((2000+3500)/2 + 8391),
		},
		{
			"no undelegations or redelegations, return provided power",
			[]stakingtypes.UnbondingDelegation{},
			[]stakingtypes.Redelegation{},
			int64(3000),
			sdk.NewInt(5),
			int64(3000), // expectedPower is 0/5 + 3000
		},
		{
			"no undelegations",
			[]stakingtypes.UnbondingDelegation{},
			// 2000 total redelegation tokens
			[]stakingtypes.Redelegation{
				createRedelegation([]int64{}, []time.Time{}),
				createRedelegation([]int64{500}, []time.Time{nowPlus1Hour}),
				createRedelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createRedelegation([]int64{700, 200}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createRedelegation([]int64{}, []time.Time{}),
				createRedelegation([]int64{100}, []time.Time{nowPlus1Hour}),
			},
			int64(17),
			sdk.NewInt(3),
			int64(2000/3 + 17),
		},
		{
			"no redelegations",
			// 2000 total undelegation tokens
			[]stakingtypes.UnbondingDelegation{
				createUndelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createUndelegation([]int64{}, []time.Time{}),
				createUndelegation([]int64{100, 100}, []time.Time{nowPlus1Hour, nowPlus1Hour}),
				createUndelegation([]int64{800}, []time.Time{nowPlus1Hour}),
				createUndelegation([]int64{500}, []time.Time{nowPlus1Hour})},
			[]stakingtypes.Redelegation{},
			int64(1),
			sdk.NewInt(3),
			int64(2000/3 + 1),
		},
		{
			"both (mature) undelegations and redelegations",
			// 2000 total undelegation tokens, 250 + 100 + 500 = 850 of those are from mature undelegations,
			// so 2000 - 850 = 1150
			[]stakingtypes.UnbondingDelegation{
				createUndelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, now}),
				createUndelegation([]int64{}, []time.Time{}),
				createUndelegation([]int64{100, 100}, []time.Time{now, nowPlus1Hour}),
				createUndelegation([]int64{800}, []time.Time{nowPlus1Hour}),
				createUndelegation([]int64{500}, []time.Time{now})},
			// 3500 total redelegation tokens, 350 + 200 + 400 = 950 of those are from mature redelegations
			// so 3500 - 950 = 2550
			[]stakingtypes.Redelegation{
				createRedelegation([]int64{}, []time.Time{}),
				createRedelegation([]int64{1600}, []time.Time{nowPlus1Hour}),
				createRedelegation([]int64{350, 250}, []time.Time{now, nowPlus1Hour}),
				createRedelegation([]int64{700, 200}, []time.Time{nowPlus1Hour, now}),
				createRedelegation([]int64{}, []time.Time{}),
				createRedelegation([]int64{400}, []time.Time{now}),
			},
			int64(8391),
			sdk.NewInt(2),
			int64((1150+2550)/2 + 8391),
		},
	}

	for _, tc := range testCases {
		actualPower := providerKeeper.ComputePowerToSlash(now,
			tc.undelegations, tc.redelegations, tc.power, tc.powerReduction)

		if tc.expectedPower != actualPower {
			require.Fail(t, fmt.Sprintf("\"%s\" failed", tc.name),
				"expected is %d but actual is %d", tc.expectedPower, actualPower)
		}
	}
}

// TestSlashValidator asserts that `SlashValidator` calls the staking module's `Slash` method
// with the correct arguments (i.e., `infractionHeight` of 0 and the expected slash power)
func TestSlashValidator(t *testing.T) {
	keeper, ctx, ctrl, mocks := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	// undelegation or redelegation entries with completion time `now` have matured
	now := ctx.BlockHeader().Time
	// undelegation or redelegation entries with completion time one hour in the future have not yet matured
	nowPlus1Hour := now.Add(time.Hour)

	keeperParams := testkeeper.NewInMemKeeperParams(t)
	testkeeper.NewInMemProviderKeeper(keeperParams, mocks)

	pubKey, _ := cryptocodec.FromTmPubKeyInterface(tmtypes.NewMockPV().PrivKey.PubKey())
	pkAny, _ := codectypes.NewAnyWithValue(pubKey)

	// manually build a validator instead of using `stakingtypes.NewValidator` to guarantee that the validator is bonded
	validator := stakingtypes.Validator{
		OperatorAddress:         sdk.ValAddress(pubKey.Address().Bytes()).String(),
		ConsensusPubkey:         pkAny,
		Jailed:                  false,
		Status:                  stakingtypes.Bonded,
		Tokens:                  sdk.ZeroInt(),
		DelegatorShares:         sdk.ZeroDec(),
		Description:             stakingtypes.Description{},
		UnbondingHeight:         int64(0),
		UnbondingTime:           time.Unix(0, 0).UTC(),
		Commission:              stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		UnbondingOnHoldRefCount: 0,
		ValidatorBondShares:     sdk.ZeroDec(),
		LiquidShares:            sdk.ZeroDec(),
	}

	consAddr, _ := validator.GetConsAddr()
	providerAddr := types.NewProviderConsAddress(consAddr)

	// we create 1000 tokens worth of undelegations, 750 of them are non-matured
	// we also create 1000 tokens worth of redelegations, 750 of them are non-matured
	undelegations := []stakingtypes.UnbondingDelegation{
		createUndelegation([]int64{250, 250}, []time.Time{nowPlus1Hour, now}),
		createUndelegation([]int64{500}, []time.Time{nowPlus1Hour})}
	redelegations := []stakingtypes.Redelegation{
		createRedelegation([]int64{250, 250}, []time.Time{now, nowPlus1Hour}),
		createRedelegation([]int64{500}, []time.Time{nowPlus1Hour})}

	// validator's current power
	currentPower := int64(3000)

	powerReduction := sdk.NewInt(2)
	slashFraction, _ := sdk.NewDecFromStr("0.5")

	// the call to `Slash` should provide an `infractionHeight` of 0 and an expected power of
	// (750 (undelegations) + 750 (redelegations)) / 2 (= powerReduction) + 3000 (currentPower) = 3750
	expectedInfractionHeight := int64(0)
	expectedSlashPower := int64(3750)

	expectedCalls := []*gomock.Call{
		mocks.MockStakingKeeper.EXPECT().
			GetValidatorByConsAddr(ctx, gomock.Any()).
			Return(validator, true),
		mocks.MockSlashingKeeper.EXPECT().
			IsTombstoned(ctx, consAddr).
			Return(false),
		mocks.MockStakingKeeper.EXPECT().
			GetUnbondingDelegationsFromValidator(ctx, validator.GetOperator()).
			Return(undelegations),
		mocks.MockStakingKeeper.EXPECT().
			GetRedelegationsFromSrcValidator(ctx, validator.GetOperator()).
			Return(redelegations),
		mocks.MockStakingKeeper.EXPECT().
			GetLastValidatorPower(ctx, validator.GetOperator()).
			Return(currentPower),
		mocks.MockStakingKeeper.EXPECT().
			PowerReduction(ctx).
			Return(powerReduction),
		mocks.MockSlashingKeeper.EXPECT().
			SlashFractionDoubleSign(ctx).
			Return(slashFraction),
		mocks.MockStakingKeeper.EXPECT().
			Slash(ctx, consAddr, expectedInfractionHeight, expectedSlashPower, slashFraction, stakingtypes.DoubleSign).
			Times(1),
	}

	gomock.InOrder(expectedCalls...)
	keeper.SlashValidator(ctx, providerAddr)
}

// TestSlashValidatorDoesNotSlashIfValidatorIsUnbonded asserts that `SlashValidator` does not call
// the staking module's `Slash` method if the validator to be slashed is unbonded
func TestSlashValidatorDoesNotSlashIfValidatorIsUnbonded(t *testing.T) {
	keeper, ctx, ctrl, mocks := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	keeperParams := testkeeper.NewInMemKeeperParams(t)
	testkeeper.NewInMemProviderKeeper(keeperParams, mocks)

	pubKey, _ := cryptocodec.FromTmPubKeyInterface(tmtypes.NewMockPV().PrivKey.PubKey())

	// validator is initially unbonded
	validator, _ := stakingtypes.NewValidator(pubKey.Address().Bytes(), pubKey, stakingtypes.Description{})

	consAddr, _ := validator.GetConsAddr()
	providerAddr := types.NewProviderConsAddress(consAddr)

	expectedCalls := []*gomock.Call{
		mocks.MockStakingKeeper.EXPECT().
			GetValidatorByConsAddr(ctx, gomock.Any()).
			Return(validator, true),
	}

	gomock.InOrder(expectedCalls...)
	keeper.SlashValidator(ctx, providerAddr)
}
