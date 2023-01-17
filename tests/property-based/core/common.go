package core

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const P = "provider"
const C = "consumer"

var initState InitState

// ValStates represents the total delegation
// and bond status of a validator
type ValStates struct {
	Delegation           []int
	Tokens               []int
	ValidatorExtraTokens []int
	Status               []stakingtypes.BondStatus
}

type InitState struct {
	NumValidators          int
	MaxValidators          int
	InitialDelegatorTokens int
	SlashDoublesign        sdk.Dec
	SlashDowntime          sdk.Dec
	UnbondingP             time.Duration
	UnbondingC             time.Duration
	Trusting               time.Duration
	MaxClockDrift          time.Duration
	BlockInterval          time.Duration
	ConsensusParams        *abci.ConsensusParams
	ValStates              ValStates
	MaxEntries             int
}

func init() {
	//	tokens === power
	sdk.DefaultPowerReduction = sdk.NewInt(1)
	/*
		These initial values heuristically lead to reasonably good exploration of behaviors.
	*/
	initState = InitState{
		NumValidators:          4,
		MaxValidators:          2,
		InitialDelegatorTokens: 10000000000000,
		SlashDoublesign:        sdk.NewDec(0),
		SlashDowntime:          sdk.NewDec(0),
		UnbondingP:             time.Second * 70,
		UnbondingC:             time.Second * 50,
		Trusting:               time.Second * 49, // Must be less than least unbonding
		MaxClockDrift:          time.Second * 10000,
		BlockInterval:          time.Second * 6, // Time between blocks in setup
		ValStates: ValStates{
			Delegation:           []int{4000, 3000, 2000, 1000},
			Tokens:               []int{5000, 4000, 3000, 2000},
			ValidatorExtraTokens: []int{1000, 1000, 1000, 1000},
			Status: []stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded,
				stakingtypes.Unbonded, stakingtypes.Unbonded},
		},
		MaxEntries: 1000000,
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxBytes: 9223372036854775807,
				MaxGas:   9223372036854775807,
			},
			Evidence: &tmproto.EvidenceParams{
				MaxAgeNumBlocks: 302400,
				MaxAgeDuration:  504 * time.Hour, // 3 weeks
				MaxBytes:        10000,
			},
			Validator: &tmproto.ValidatorParams{
				PubKeyTypes: []string{
					tmtypes.ABCIPubKeyTypeEd25519,
				},
			},
		},
	}
}
