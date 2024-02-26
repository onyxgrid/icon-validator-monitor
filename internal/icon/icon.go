package icon

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strconv"

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
	params := map[string]interface{}{
		"address": address, 
	}

	callObject := transactions.CallBuilder("cx0000000000000000000000000000000000000000", "getDelegation", params)
	response, err := i.client.Call(callObject)
	if err != nil {
		fmt.Println(err)
	}

	responseData, ok := response.(map[string]interface{})
	if !ok {
		return model.DelegationResponse{}, fmt.Errorf("response is not of type map[string]interface{}")
	}
	
	// Extracting and assigning values to the struct fields
	res := model.DelegationResponse{
		TotalDelegated: fmt.Sprintf("%v", responseData["totalDelegated"]),
		VotingPower:    fmt.Sprintf("%v", responseData["votingPower"]),
	}
	
	delegations, ok := responseData["delegations"].([]interface{})
	if !ok {
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

		// Convert the value to a big.Int
		valueBigInt, ok := new(big.Int).SetString(value, 0)
		if !ok {
			fmt.Println("Failed to convert value to big.Int")
			continue
		}
	
		res.Delegations = append(res.Delegations, model.Delegation{
			Address: address,
			Value:   valueBigInt,
		})

		name, err := i.GetValidatorName(address)
		if err != nil {
			fmt.Println(err)
		}
		res.Delegations[len(res.Delegations)-1].Name = name
	}

	return res, nil
}


// GetBonds returns the bonds of a wallet
func (i *Icon) GetBonds(address string) (model.BondResponse, error) {
	params := map[string]interface{}{
		"address": address, 
	}

	callObject := transactions.CallBuilder("cx0000000000000000000000000000000000000000", "getBond", params)
	response, err := i.client.Call(callObject)
	if err != nil {
		fmt.Println(err)
	}

	responseData, ok := response.(map[string]interface{})

	if !ok {
		return model.BondResponse{}, fmt.Errorf("response is not of type map[string]interface{}")
	}

	// Extracting and assigning values to the struct fields
	res := model.BondResponse{
		TotalBonded: fmt.Sprintf("%v", responseData["totalBonded"]),
		VotingPower: fmt.Sprintf("%v", responseData["votingPower"]),
	}

	bonds, ok := responseData["bonds"].([]interface{})
	if !ok {
		return model.BondResponse{}, fmt.Errorf("bonds field is not of type []interface{}")
	}

	for _, bondData := range bonds {
		bond, ok := bondData.(map[string]interface{})
		if !ok {
			fmt.Println("Bond data is not of type map[string]interface{}")
			continue
		}

		address, ok := bond["address"].(string)
		if !ok {
			fmt.Println("Address field is not of type string")
			continue
		}

		value, ok := bond["value"].(string)
		if !ok {
			fmt.Println("Value field is not of string")
			continue
		}

		// Convert the value to a big.Int
		valueBigInt, ok := new(big.Int).SetString(value, 0)
		if !ok {
			fmt.Println("Failed to convert value to big.Int")
			continue
		}

		res.Bonds = append(res.Bonds, model.Bond{
			Address: address,
			Value:   valueBigInt,
		})

		name, err := i.GetValidatorName(address)
		if err != nil {
			fmt.Println(err)
		}
		res.Bonds[len(res.Bonds)-1].Name = name
	}

	return res, nil
}

// GetValidatorName returns the name of a validator given its address
func (i *Icon) GetValidatorName(address string) (string, error) {
	params := map[string]interface{}{
		"address": address, 
	}

	callObject := transactions.CallBuilder("cx0000000000000000000000000000000000000000", "getPRep", params)
	response, err := i.client.Call(callObject)
	if err != nil {
		fmt.Println(err)
	}

	responseData, ok := response.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("response is not of type map[string]interface{}")
	}

	name, ok := responseData["name"].(string)
	if !ok {
		return "", fmt.Errorf("name field is not of type string")
	}

	return name, nil
}

// GetAllValidators returns the list of all validators
func (i *Icon) GetAllValidators() ([]model.ValidatorInfo, error){
	u := "https://tracker.icon.community/api/v1/governance/preps"

	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	totalCount := response.Header.Get("X-Total-Count")
	c, err := strconv.Atoi(totalCount)
	if err != nil {
		fmt.Println(err)
	}
	// not sure if we need count to paginate, or if we can just get all the validators at once
	_ = c

	var vs []model.ValidatorInfo
	err = json.NewDecoder(response.Body).Decode(&vs)
	if err != nil {
		return nil, err
	}

	return vs, nil
}