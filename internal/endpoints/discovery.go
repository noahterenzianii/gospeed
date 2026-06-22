package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func FetchServers(url string, timeout time.Duration) ([]Server, error) {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %d", resp.StatusCode)
	}

	var servers []Server
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, err
	}
	return servers, nil
}
