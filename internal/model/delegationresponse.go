package model

import "math/big"

type Delegation struct {
	Address string   `json:"address"`
	Value   *big.Int `json:"value"`
	Name    string
}

type DelegationResponse struct {
	Delegations    []Delegation `json:"delegations"`
	TotalDelegated string       `json:"totalDelegated"`
	VotingPower    string       `json:"votingPower"`
}
