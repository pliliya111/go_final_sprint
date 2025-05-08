package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pliliya111/go_final_sprint/internal/model"
)

var (
	serverURL = "http://localhost:8080"
)

func FetchTask() (*model.Task, error) {
	resp, err := http.Get(fmt.Sprintf("%s/internal/task", serverURL))
	if err != nil {
		return nil, fmt.Errorf("error fetching task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch task: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	log.Printf("Response body: %s", string(body))

	var response struct {
		Task model.Task `json:"task"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error decoding task: %v", err)
	}

	return &response.Task, nil
}

func SubmitTaskResult(taskID string, result interface{}) error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"id":     taskID,
		"result": result,
	})
	if err != nil {
		return fmt.Errorf("error marshaling result: %v", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/internal/task", serverURL), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error submitting result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to submit result: %s", resp.Status)
	}

	return nil
}
