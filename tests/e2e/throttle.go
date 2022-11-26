package e2e

import (
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providertypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
)

// TestBasicSlashPacketThrottling tests slash packet throttling with a single consumer,
// two slash packets, and no VSC matured packets. The most basic scenario.
func (s *CCVTestSuite) TestBasicSlashPacketThrottling() {

	// setupValidatePowers gives the default 4 validators 25% power each (1000 power).
	// Note this in test cases.
	testCases := []struct {
		replenishFraction                string
		expectedMeterBeforeFirstSlash    int64
		expectedMeterAfterFirstSlash     int64
		expectedAllowanceAfterFirstSlash int64
		expectedReplenishesTillPositive  int
	}{
		{"0.2", 800, -200, 600, 1},
		{"0.1", 400, -600, 300, 3}, // 600/300 = 2, so 3 replenishes to reach positive
		{"0.05", 200, -800, 150, 6},
		{"0.01", 40, -960, 30, 33}, // 960/30 = 32, so 33 replenishes to reach positive
	}

	for _, tc := range testCases {

		s.SetupTest()
		s.SetupAllCCVChannels()
		s.setupValidatorPowers()

		providerStakingKeeper := s.providerApp.GetE2eStakingKeeper()

		// Use default params (incl replenish period), but set replenish fraction to tc value.
		params := providertypes.DefaultParams()
		params.SlashMeterReplenishFraction = tc.replenishFraction
		s.providerApp.GetProviderKeeper().SetParams(s.providerCtx(), params)

		// Elapse a replenish period and check for replenishment, so new param is fully in effect.
		customCtx := s.getCtxWithReplenishPeriodElapsed(s.providerCtx())
		s.providerApp.GetProviderKeeper().CheckForSlashMeterReplenishment(customCtx)

		slashMeter := s.providerApp.GetProviderKeeper().GetSlashMeter(s.providerCtx())
		s.Require().Equal(tc.expectedMeterBeforeFirstSlash, slashMeter.Int64())

		// Assert that we start out with no jailings
		vals := providerStakingKeeper.GetAllValidators(s.providerCtx())
		for _, val := range vals {
			s.Require().False(val.IsJailed())
		}

		// Send a slash packet from consumer to provider
		s.setDefaultValSigningInfo(*s.providerChain.Vals.Validators[0])
		packet := s.constructSlashPacketFromConsumer(s.getFirstBundle(), 0, stakingtypes.Downtime, 1)
		sendOnConsumerRecvOnProvider(s, s.getFirstBundle().Path, packet)

		// Assert validator 0 is jailed and has no power
		vals = providerStakingKeeper.GetAllValidators(s.providerCtx())
		slashedVal := vals[0]
		s.Require().True(slashedVal.IsJailed())
		lastValPower := providerStakingKeeper.GetLastValidatorPower(s.providerCtx(), slashedVal.GetOperator())
		s.Require().Equal(int64(0), lastValPower)

		// Assert expected slash meter and allowance value
		slashMeter = s.providerApp.GetProviderKeeper().GetSlashMeter(s.providerCtx())
		s.Require().Equal(tc.expectedMeterAfterFirstSlash, slashMeter.Int64())
		s.Require().Equal(tc.expectedAllowanceAfterFirstSlash,
			s.providerApp.GetProviderKeeper().GetSlashMeterAllowance(s.providerCtx()).Int64())

		// Now send a second slash packet from consumer to provider for a different validator.
		s.setDefaultValSigningInfo(*s.providerChain.Vals.Validators[2])
		packet = s.constructSlashPacketFromConsumer(s.getFirstBundle(), 2, stakingtypes.Downtime, 2)
		sendOnConsumerRecvOnProvider(s, s.getFirstBundle().Path, packet)

		// Require that slash packet has not been handled
		vals = providerStakingKeeper.GetAllValidators(s.providerCtx())
		s.Require().False(vals[2].IsJailed())

		// Assert slash meter value is still the same
		slashMeter = s.providerApp.GetProviderKeeper().GetSlashMeter(s.providerCtx())
		s.Require().Equal(tc.expectedMeterAfterFirstSlash, slashMeter.Int64())

		// Replenish slash meter until it is positive
		for i := 0; i < tc.expectedReplenishesTillPositive; i++ {

			// Mutate context with a block time where replenish period has passed.
			customCtx = s.getCtxWithReplenishPeriodElapsed(s.providerCtx())

			// CheckForSlashMeterReplenishment should replenish meter here.
			slashMeterBefore := s.providerApp.GetProviderKeeper().GetSlashMeter(s.providerCtx())
			s.providerApp.GetProviderKeeper().CheckForSlashMeterReplenishment(customCtx)
			slashMeter = s.providerApp.GetProviderKeeper().GetSlashMeter(s.providerCtx())
			s.Require().True(slashMeter.GT(slashMeterBefore))

			// Check that slash meter is still negative or 0,
			// unless we are on the last iteration.
			if i != tc.expectedReplenishesTillPositive-1 {
				s.Require().False(slashMeter.IsPositive())
			}
		}

		// Meter is positive at this point, and ready to handle the second slash packet.
		slashMeter = s.providerApp.GetProviderKeeper().GetSlashMeter(s.providerCtx())
		s.Require().True(slashMeter.IsPositive())

		// Assert validator 2 is jailed once pending slash packets are handled in ccv endblocker.
		s.providerChain.NextBlock()
		vals = providerStakingKeeper.GetAllValidators(s.providerCtx())
		slashedVal = vals[2]
		s.Require().True(slashedVal.IsJailed())

		// Assert validator 2 has no power, this should be apparent next block,
		// since the staking endblocker runs before the ccv endblocker.
		s.providerChain.NextBlock()
		lastValPower = providerStakingKeeper.GetLastValidatorPower(s.providerCtx(), slashedVal.GetOperator())
		s.Require().Equal(int64(0), lastValPower)
	}
}

// TestSlashingSmallValidators tests that multiple slash packets from validators with small
// power can be handled by the provider chain in a non-throttled manner.
func (s *CCVTestSuite) TestSlashingSmallValidators() {

	s.SetupAllCCVChannels()

	// Setup first val with 1000 power and the rest with 10 power.
	delAddr := s.providerChain.SenderAccount.GetAddress()
	delegateByIdx(s, delAddr, sdktypes.NewInt(999999999), 0)
	delegateByIdx(s, delAddr, sdktypes.NewInt(9999999), 1)
	delegateByIdx(s, delAddr, sdktypes.NewInt(9999999), 2)
	delegateByIdx(s, delAddr, sdktypes.NewInt(9999999), 3)
	s.providerChain.NextBlock()

	// Replenish slash meter with default params and new total voting power.
	customCtx := s.getCtxWithReplenishPeriodElapsed(s.providerCtx())
	s.providerApp.GetProviderKeeper().CheckForSlashMeterReplenishment(customCtx)

	// Assert that we start out with no jailings
	providerStakingKeeper := s.providerApp.GetE2eStakingKeeper()
	vals := providerStakingKeeper.GetAllValidators(s.providerCtx())
	for _, val := range vals {
		s.Require().False(val.IsJailed())
	}

	// Setup signing info for jailings
	s.setDefaultValSigningInfo(*s.providerChain.Vals.Validators[1])
	s.setDefaultValSigningInfo(*s.providerChain.Vals.Validators[2])
	s.setDefaultValSigningInfo(*s.providerChain.Vals.Validators[3])

	// Send slash packets from consumer to provider for small validators.
	packet1 := s.constructSlashPacketFromConsumer(s.getFirstBundle(), 1, stakingtypes.DoubleSign, 1)
	packet2 := s.constructSlashPacketFromConsumer(s.getFirstBundle(), 2, stakingtypes.Downtime, 2)
	packet3 := s.constructSlashPacketFromConsumer(s.getFirstBundle(), 3, stakingtypes.Downtime, 3)
	sendOnConsumerRecvOnProvider(s, s.getFirstBundle().Path, packet1)
	sendOnConsumerRecvOnProvider(s, s.getFirstBundle().Path, packet2)
	sendOnConsumerRecvOnProvider(s, s.getFirstBundle().Path, packet3)

	// Default slash meter replenish fraction is 0.05, so all sent packets should be handled immediately.
	vals = providerStakingKeeper.GetAllValidators(s.providerCtx())
	s.Require().False(vals[0].IsJailed())
	s.Require().Equal(int64(1000),
		providerStakingKeeper.GetLastValidatorPower(s.providerCtx(), vals[0].GetOperator()))
	s.Require().True(vals[1].IsJailed())
	s.Require().Equal(int64(0),
		providerStakingKeeper.GetLastValidatorPower(s.providerCtx(), vals[1].GetOperator()))
	s.Require().True(vals[2].IsJailed())
	s.Require().Equal(int64(0),
		providerStakingKeeper.GetLastValidatorPower(s.providerCtx(), vals[2].GetOperator()))
	s.Require().True(vals[3].IsJailed())
	s.Require().Equal(int64(0),
		providerStakingKeeper.GetLastValidatorPower(s.providerCtx(), vals[3].GetOperator()))
}

func (s *CCVTestSuite) TestSlashMeterAllowanceChanges() {
	s.SetupAllCCVChannels()

	providerKeeper := s.providerApp.GetProviderKeeper()

	// At first, allowance is based on 4 vals all with 1 power, min allowance is in effect.
	s.Require().Equal(int64(1), providerKeeper.GetSlashMeterAllowance(s.providerCtx()).Int64())

	s.setupValidatorPowers()

	// Now all 4 validators have 1000 power (4000 total power) so allowance should be:
	// default replenish frac * 4000
	expectedAllowance := sdktypes.MustNewDecFromStr(
		providertypes.DefaultSlashMeterReplenishFraction).MulInt64(4000).RoundInt64()
	s.Require().Equal(expectedAllowance, providerKeeper.GetSlashMeterAllowance(s.providerCtx()).Int64())

	// Now we change replenish fraction and assert new expected allowance.
	params := providerKeeper.GetParams(s.providerCtx())
	params.SlashMeterReplenishFraction = "0.3"
	providerKeeper.SetParams(s.providerCtx(), params)
	s.Require().Equal(int64(1200), providerKeeper.GetSlashMeterAllowance(s.providerCtx()).Int64())

}

func (s *CCVTestSuite) getCtxWithReplenishPeriodElapsed(ctx sdktypes.Context) sdktypes.Context {

	providerKeeper := s.providerApp.GetProviderKeeper()
	lastReplenishTime := providerKeeper.GetLastSlashMeterReplenishTime(ctx)
	replenishPeriod := providerKeeper.GetSlashMeterReplenishPeriod(ctx)

	return ctx.WithBlockTime(lastReplenishTime.Add(replenishPeriod).Add(time.Minute))
}

// TODO: assert more logic about meter level, etc.

// TODO(Shawn): Add more e2e tests for edge cases

// TODO: test vsc matured stuff too, or add to above test?

// TODO: multiple consumers
