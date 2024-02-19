package model

import "math/big"

type Bond struct {
    Address string `json:"address"`
    Value   *big.Int `json:"value"`
	Name string
}

type BondResponse struct {
    Bonds       []Bond `json:"bonds"`
    TotalBonded string `json:"totalBonded"`
    Unbonds     []Bond `json:"unbonds"`
    VotingPower string `json:"votingPower"`
}
