package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Event struct {
	Type    string `json:"type"`
	Repo    Repo   `json:"repo"`
	Payload struct {
		Action      string `json:"action"`
		Ref         string `json:"ref"`
		RefType     string `json:"ref_type"`
		PushID      int64  `json:"push_id"`
		Size        int    `json:"size"`
		IssueTitle  string `json:"issue,title"`
		PullRequest struct {
			Title string `json:"title"`
		} `json:"pull_request"`
	} `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	Name string `json:"name"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: github-activity <username>")
		os.Exit(1)
	}

	username := os.Args[1]
	events, err := fetchUserEvents(username)
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
		os.Exit(1)
	}

	displayEvents(events)
}

func fetchUserEvents(username string) ([]Event, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "GitHub-Activity-CLI")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user '%s' not found", username)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var events []Event
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&events); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return events, nil
}

func displayEvents(events []Event) {
	if len(events) == 0 {
		fmt.Println("No recent activity found")
		return
	}

	for _, event := range events {
		description := formatEvent(event)
		if description != "" {
			fmt.Printf("- %s\n", description)
		}
	}
}

func formatEvent(event Event) string {
	switch event.Type {
	case "PushEvent":
		return fmt.Sprintf("Pushed %d commits to %s", 
			event.Payload.Size, 
			event.Repo.Name)
	
	case "CreateEvent":
		return fmt.Sprintf("Created %s '%s' in %s", 
			event.Payload.RefType, 
			event.Payload.Ref, 
			event.Repo.Name)
	
	case "IssuesEvent":
		return fmt.Sprintf("%s issue in %s", 
			capitalize(event.Payload.Action), 
			event.Repo.Name)
	
	case "PullRequestEvent":
		return fmt.Sprintf("%s pull request '%s' in %s", 
			capitalize(event.Payload.Action), 
			event.Payload.PullRequest.Title, 
			event.Repo.Name)
	
	case "WatchEvent":
		return fmt.Sprintf("Starred %s", event.Repo.Name)
	
	case "ForkEvent":
		return fmt.Sprintf("Forked %s", event.Repo.Name)
	
	default:
		return ""
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return fmt.Sprintf("%c%s", s[0]-32, s[1:])
}