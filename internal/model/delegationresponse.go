package model

type Delegation struct {
    Address string `json:"address"`
    Value   string `json:"value"`
	Name string
}

type DelegationResponse struct {
    Delegations   []Delegation `json:"delegations"`
    TotalDelegated string      `json:"totalDelegated"`
    VotingPower    string      `json:"votingPower"`
}