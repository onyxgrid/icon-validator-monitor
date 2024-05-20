package icon

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/eyeonicon/go-icon-sdk/transactions"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
)

var CPSContract = "cx9f4ab72f854d3ccdc59aa6f2c3e2215dd62e879f"
var EOIWALLET = "hx2e7db537ca3ff73336bee1bab4cf733a94ae769b"



// 1 get cps registered validators - getPReps()
func (i *Icon) GetPreps() ([]model.CPSPreps, error) {
	callObject := transactions.CallBuilder(CPSContract, "getPReps", nil)
	response, err := i.client.Call(callObject)
	if err != nil {
		return nil, err
	}

	res := response.([]interface{})
	var preps []model.CPSPreps
	for _, r := range res {
		rm := r.(map[string]interface{})
		p := model.CPSPreps{
			Address:   safeString(rm, "address"),
			Name:      safeString(rm, "name"),
			Delegated: safeString(rm, "delegated"),
		}
		preps = append(preps, p)
	}

	return preps, nil
}

// GetProposalDetails return the details of all pending proposals
// we actually not using this function atm
func (i *Icon) GetProposalDetails(a, t string) (model.ProposalDetailsResponse, error) {
	if t != "progress_reports" && t != "proposal" {
		return model.ProposalDetailsResponse{}, fmt.Errorf("invalid type, must be either progress_reports or proposal")
	}

	var pd model.ProposalDetailsResponse

	idx := 0
	params := map[string]interface{}{
		"status":        "_pending",
		"walletAddress": a,
		"startIndex":    fmt.Sprintf("0x%x", idx),
	}

	callObject := transactions.CallBuilder(CPSContract, "getProposalDetails", params)
	response, err := i.client.Call(callObject)
	if err != nil {
		return pd, err
	}

	res := response.(map[string]interface{})
	ch := res["count"].(string)[2:]
	cd, err := strconv.ParseInt(ch, 16, 64)
	if err != nil {
		return pd, err
	}

	pd.Count = int(cd)

	data := res["data"].([]interface{})
	for _, d := range data {
		dm := d.(map[string]interface{})
		p := model.ProposalDetails{
			AbstainVoters:       safeString(dm, "abstain_voters"),
			AbstainedVotes:      safeString(dm, "abstained_votes"),
			ApproveVoters:       safeString(dm, "approve_voters"),
			ApprovedVotes:       safeString(dm, "approved_votes"),
			IPFSHash:            safeString(dm, "ipfs_hash"),
			PercentageCompleted: safeString(dm, "percentage_completed"),
			ProjectTitle:        safeString(dm, "project_title"),
			RejectVoters:        safeString(dm, "reject_voters"),
			RejectedVotes:       safeString(dm, "rejected_votes"),
			Status:              safeString(dm, "status"),
			Token:               safeString(dm, "token"),
			TotalBudget:         safeString(dm, "total_budget"),
			TotalVoters:         safeString(dm, "total_voters"),
			TotalVotes:          safeString(dm, "total_votes"),
		}
		pd.Data = append(pd.Data, p)
	}

	return pd, err
}

func (i *Icon) CheckPriorityVoting(a string) (bool, error) {
	params := map[string]interface{}{
		"_prep": a,
	}

	callObject := transactions.CallBuilder(CPSContract, "checkPriorityVoting", params)
	response, err := i.client.Call(callObject)
	if err != nil {
		return false, err
	}

	res := response.(string)
	if res == "0x1" {
		return true, nil
	}

	return false, nil
}

// getRemainingProject returns the outstanding votes per validator
func (i *Icon) GetRemainingProject(a, t string) ([]interface{}, error) {
	if t != "progress_reports" && t != "proposal" {
		return nil, fmt.Errorf("invalid type, must be either progress_reports or proposal")
	}

	params := map[string]interface{}{
		"walletAddress": a,
		"projectType":   t,
	}

	callObject := transactions.CallBuilder(CPSContract, "getRemainingProject", params)
	response, err := i.client.Call(callObject)
	if err != nil {
		return nil, err
	}

	res := response.([]interface{})
	return res, nil
}

func (i *Icon) GetRemainingTimePeriod() (time.Duration, error) {
	// getPeriodStatus
	callObject := transactions.CallBuilder(CPSContract, "getPeriodStatus", nil)
	response, err := i.client.Call(callObject)

	if err != nil {
		return -1, err
	}

	res := response.(map[string]interface{})
	remainingTime := res["remaining_time"].(string)[2:]
	remainingTimeInt, err := strconv.ParseInt(remainingTime, 16, 64)
	if err != nil {
		return -1, err
	}

	t := time.Duration(remainingTimeInt) * time.Second
	return t, nil
}

// safeString safely retrieves a string value from a map, given a key
func safeString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	log.Printf("Warning: key %s is not of type string", key)
	return ""
}
