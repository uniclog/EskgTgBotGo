package service

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type License struct {
	Name    string `json:"Name"`
	Key     string `json:"Key"`
	OutDate string `json:"Out Date"`
}

func GetLogs(owner, repo, runID, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%s/logs", owner, repo, runID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Sprintf("Ошибка создания запроса: %v", err), err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Ошибка выполнения запроса: %v", err), err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Не удалось получить логи. Статус: %s", resp.Status), err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Ошибка чтения тела ответа: %v", err), err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Sprintf("Ошибка открытия zip-архива: %v", err), err
	}

	var license []License
	var lic = License{}

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "0_GenerateKey.txt") {
			continue
		}
		zippedFile, err := file.Open()
		if err != nil {
			log.Printf("Ошибка открытия файла внутри zip: %v", err)
		}
		scanner := bufio.NewScanner(zippedFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "License Name") {
				lic.Name = clearLine(line)
			}
			if strings.Contains(line, "License Key") {
				lic.Key = clearLine(line)
			}
			if strings.Contains(line, "License Out Date") {
				lic.OutDate = clearLine(line)
				license = append(license, lic)
				lic = License{}
			}
		}
		_ = zippedFile.Close()
	}

	licenseJ, _ := json.MarshalIndent(license, "", "  ")
	log.Printf(string(licenseJ))

	return string(licenseJ), nil
}

func clearLine(input string) string {
	re := regexp.MustCompile(`^.*?Z\s+.*?:\s*`)
	return re.ReplaceAllString(input, "")
}
