package icon

import (
	"fmt"
	"os"
	"regexp"

	"github.com/eyeonicon/go-icon-sdk/transactions"
	"github.com/icon-project/goloop/client"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
)

type Icon struct {
	client *client.ClientV3
}

// NewIcon creates a new Icon client
func NewIcon() (*Icon, error){
	rpc := os.Getenv("ICON_RPC")
	if rpc == "" {
		return nil, fmt.Errorf("ICON_RPC is not set")
	}

	c := client.NewClientV3(rpc)
	if c == nil {
		return nil, fmt.Errorf("failed to create new client")
	}

	return &Icon{client: c}, nil
}

// isValidIconAddress checks if a string is a valid ICON wallet address
func IsValidIconAddress(address string) bool {
	// Regular expression pattern for ICON wallet address
	pattern := "^hx[0-9a-fA-F]{40}$"

	// Compile the regular expression
	reg := regexp.MustCompile(pattern)

	// Check if the address matches the pattern
	return reg.MatchString(address)
}

// getDelegation returns the delegation of a wallet
func (i *Icon) GetDelegation(address string) (model.DelegationResponse, error) {
	// the parameter _tokenId is set to 0x2
	params := map[string]interface{}{
		"address": address, 
	}

	// create call object with params as nil
	callObject := transactions.CallBuilder("cx0000000000000000000000000000000000000000", "getDelegation", params)

	// make the call
	response, err := i.client.Call(callObject)
	if err != nil {
		fmt.Println(err)
	}

	responseData, ok := response.(map[string]interface{})
	if !ok {
		fmt.Println("Response is not of type map[string]interface{}")
		return model.DelegationResponse{}, fmt.Errorf("response is not of type map[string]interface{}")
	}
	
	// Extracting and assigning values to the struct fields
	res := model.DelegationResponse{
		TotalDelegated: fmt.Sprintf("%v", responseData["totalDelegated"]),
		VotingPower:    fmt.Sprintf("%v", responseData["votingPower"]),
	}
	
	delegations, ok := responseData["delegations"].([]interface{})
	if !ok {
		fmt.Println("Delegations field is not of type []interface{}")
		return model.DelegationResponse{}, fmt.Errorf("delegations field is not of type []interface{}")
	}
	
	for _, delegationData := range delegations {
		delegation, ok := delegationData.(map[string]interface{})
		if !ok {
			fmt.Println("Delegation data is not of type map[string]interface{}")
			continue
		}
	
		address, ok := delegation["address"].(string)
		if !ok {
			fmt.Println("Address field is not of type string")
			continue
		}
	
		value, ok := delegation["value"].(string)
		if !ok {
			fmt.Println("Value field is not of type string")
			continue
		}
	
		res.Delegations = append(res.Delegations, model.Delegation{
			Address: address,
			Value:   value,
		})

		// Get the name of the validator
		name, err := i.GetValidatorName(address)
		if err != nil {
			fmt.Println(err)
		}
		res.Delegations[len(res.Delegations)-1].Name = name
	}

	return res, nil
}

// GetValidatorName returns the name of a validator given its address
func (i *Icon) GetValidatorName(address string) (string, error) {
	// the parameter _tokenId is set to 0x2
	params := map[string]interface{}{
		"address": address, 
	}

	// create call object with params as nil
	callObject := transactions.CallBuilder("cx0000000000000000000000000000000000000000", "getPRep", params)

	// make the call
	response, err := i.client.Call(callObject)
	if err != nil {
		fmt.Println(err)
	}

	responseData, ok := response.(map[string]interface{})
	if !ok {
		fmt.Println("Response is not of type map[string]interface{}")
		return "", fmt.Errorf("response is not of type map[string]interface{}")
	}

	name, ok := responseData["name"].(string)
	if !ok {
		fmt.Println("Name field is not of type string")
		return "", fmt.Errorf("name field is not of type string")
	}

	return name, nil
}