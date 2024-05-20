package model

// ProposalDetails represents the structure of each proposal.
type ProposalDetails struct {
	AbstainVoters       string `json:"abstain_voters"`
	AbstainedVotes      string `json:"abstained_votes"`
	ApproveVoters       string `json:"approve_voters"`
	ApprovedVotes       string `json:"approved_votes"`
	IPFSHash            string `json:"ipfs_hash"`
	PercentageCompleted string `json:"percentage_completed"`
	ProjectTitle        string `json:"project_title"`
	RejectVoters        string `json:"reject_voters"`
	RejectedVotes       string `json:"rejected_votes"`
	Status              string `json:"status"`
	Token               string `json:"token"`
	TotalBudget         string `json:"total_budget"`
	TotalVoters         string `json:"total_voters"`
	TotalVotes          string `json:"total_votes"`
}

// RPCResponse represents the structure of the overall RPC response.
type ProposalDetailsResponse struct {
	Count int               `json:"count"`
	Data  []ProposalDetails `json:"data"`
}

type CPSPreps struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	Delegated string `json:"delegated"`
}
