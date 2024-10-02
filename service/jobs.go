package service

import (
	"EskgTgBotGo/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type WorkflowRun struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Jobs       []Job  `json:"jobs"`
}

type Job struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

func TriggerWorkflow(count string, config *config.Config) string {
	owner := config.Git.Owner
	repo := config.Git.Repo
	workflowID := config.Git.WorkflowID
	branch := config.Git.Branch
	token := config.Git.Token
	keyCount := count

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches", owner, repo, workflowID)
	data := map[string]interface{}{
		"ref": branch,
		"inputs": map[string]string{
			"key": keyCount,
		},
	}
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		data := fmt.Sprintf("Ошибка создания запроса: %v", err)
		log.Printf(data)
		return data
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		data := fmt.Sprintf("Ошибка выполнения запроса: %v", err)
		log.Printf(data)
		return data
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		data := fmt.Sprintf("Ошибка выполнения workflow: %s", string(body))
		log.Printf(data)
		return data
	}
	return "Запрос на генерацию отправлен"
}

func GetLatestWorkflowRunID(owner, repo, token string) (int64, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Ошибка получения списка workflow run'ов: %s", string(body))
		return 0, fmt.Errorf("ошибка получения списка workflow run'ов: %s", string(body))
	}

	var result struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Ошибка декодирования ответа: %v", err)
		return 0, err
	}

	if len(result.WorkflowRuns) > 0 {
		fmt.Printf("WorkflowRunID: %d\n", result.WorkflowRuns[0].ID)
		return result.WorkflowRuns[0].ID, nil
	}

	return 0, fmt.Errorf("нет запущенных workflow run'ов")
}

func checkWorkflowStatus(owner, repo, runID, token string) (*WorkflowRun, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%s", owner, repo, runID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Ошибка получения статуса: %s", string(body))
		return nil, fmt.Errorf("ошибка получения списка workflow run'ов: %s", string(body))
	}

	var run WorkflowRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		log.Printf("Ошибка декодирования ответа: %v", err)
		return nil, err
	}

	return &run, nil
}

func GetGitWorkflowResult(config *config.Config) string {
	owner := config.Git.Owner
	repo := config.Git.Repo
	token := config.Git.Token

	var data string

	runIDValue, err := GetLatestWorkflowRunID(owner, repo, token)
	if err != nil {
		return fmt.Sprint(err)
	}
	runID := fmt.Sprint(runIDValue)
	for {
		run, err := checkWorkflowStatus(owner, repo, runID, token)
		if err != nil {
			return fmt.Sprintf("Ошибка получения статуса workflow: %v", err)
		}
		if run.Status == "completed" {
			data = fmt.Sprintf("Workflow завершен. Результат: %s\n", run.Conclusion)
			if run.Conclusion != "cancelled" {
				val, _ := GetLogs(owner, repo, runID, token)
				data = val
			}
			break
		}
		time.Sleep(10 * time.Second)
	}

	return data
}
