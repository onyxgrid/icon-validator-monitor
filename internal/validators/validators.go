package validators

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paulrouge/icon-validator-monitor/internal/model"
)

// ValidatorInfo returns the information of a validator given its IP address
func ValidatorInfo(ip string) (*model.PrepResponse, error) {
	u := "https://tracker.icon.community/api/v1/governance/preps/" + ip

	// get the response
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var res []model.PrepResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("validator not found")
	}

	return &res[0], nil
}