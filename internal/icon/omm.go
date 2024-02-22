package icon

import (
	"fmt"
	"math/big"

	"github.com/eyeonicon/go-icon-sdk/transactions"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
)

const ommDelegationContract = "cx841f29ec6ce98b527d49a275e87d427627f1afe5"

func (i *Icon) GetOmmVotes(address string) []model.OmmResponse {
	params := map[string]interface{}{
		"_user": address,
	}

	callObject := transactions.CallBuilder(ommDelegationContract, "getUserICXDelegation", params)
	// make the call
	response, err := i.client.Call(callObject)
	if err != nil {
		fmt.Println(err)
	}

	responses, ok := response.([]interface{})
	if !ok {
		fmt.Println("Response is not of type map[string]interface{}")
		return nil
	}
	
	var ommVotes []model.OmmResponse
	for _, response := range responses {
		responseData, ok := response.(map[string]interface{})
		if !ok {
			fmt.Println("Response is not of type map[string]interface{}")
			return nil
		}

		votes := responseData["_votes_in_icx"].(string)
		votes = votes[2:]

		votesbi := new(big.Int)
		votesbi.SetString(votes, 16)

		validator := responseData["_address"].(string)

		// Get the name of the validator
		name, err := i.GetValidatorName(validator)
		if err != nil {
			fmt.Println(err)
		}

		ov := model.OmmResponse{
			Address: validator,
			VotesInIcx: votesbi,
			Name: name,
		}

		ommVotes = append(ommVotes, ov)
	}

	return ommVotes
}