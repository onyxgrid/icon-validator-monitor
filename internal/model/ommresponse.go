package model

import "math/big"

type OmmResponse struct {
    Address    string `json:"_address"`
    VotesInIcx *big.Int `json:"_votes_in_icx"`
    VotesInPer string `json:"_votes_in_per"`
    Name      string
}
