package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"pr-mail/app/dto"
)

// func FetchPRDetails(url string) (*dto.GitHubPRResponse, error) {
func FetchPRDetails(owner, repo string, prNumber int) (*dto.GitHubPRResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// gitHub Token from .env
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN not set in environment")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	//
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//Log response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("Raw GitHub API Response:\n", string(bodyBytes))

	// Unmarshal into struct
	var data dto.GitHubPRResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	return &data, nil
}
