package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const clientInfoURL = "https://ipinfo.io/json"

func FetchClientInfo() (*ClientInfo, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(clientInfoURL)
	if err != nil {
		return nil, fmt.Errorf("fetch client info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var info ClientInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &info, nil
}
