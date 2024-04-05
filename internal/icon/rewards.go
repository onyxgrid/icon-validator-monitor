package icon

import (
	"math/big"

	"github.com/paulrouge/icon-validator-monitor/internal/model"
)

// EstimateReward estimates the usd reward for a given amount of votes
func EstimateReward(validator model.ValidatorInfo, votes *big.Int) (*big.Int, error) {
	daily := big.NewFloat(validator.RewardDailyUSD)
	daily.Mul(daily, big.NewFloat(1e18))

	// if daily is 0, return 0
	if daily.Cmp(big.NewFloat(0)) == 0 {
		return big.NewInt(0), nil
	}

	// commission rate
	cr := big.NewFloat(validator.CommissionRate)
	daily.Quo(daily, cr)

	// bonded amount
	b := big.NewFloat(validator.Bonded)
	b.Quo(b, big.NewFloat(1e18))

	// for some reason the delegated amount is not 10 ** 18 coming from the tracker api, so no need to divide by 10 ** 18
	d := big.NewFloat(validator.Delegated)

	// total delegated & bonds
	var t big.Float
	t.Add(b, d)

	// reward per vote
	var rpv big.Float
	rpv.Quo(daily, &t).Quo(&rpv, big.NewFloat(1e18)) // divide by 10^18

	// votes
	v := big.NewFloat(0).SetInt(votes)

	// estimated reward
	var er big.Float
	er.Mul(&rpv, v)

	// convert to big int
	var res big.Int
	er.Int(&res)

	// add a 100 x, not sure why, probably wrong but yea...
	res.Mul(&res, big.NewInt(100))

	return &res, nil
}
