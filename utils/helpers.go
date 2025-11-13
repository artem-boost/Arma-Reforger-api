package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func GenerateRandomString(length int, onlyNumbers bool) string {
	var characters string
	if onlyNumbers {
		characters = "0123456789"
	} else {
		characters = "abcdefghijklmnopqrstuvwxyz0123456789"
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	for i, b := range bytes {
		bytes[i] = characters[b%byte(len(characters))]
	}
	return string(bytes)
}

func GenerateUUID() string {
	return uuid.New().String()
}

func ParseJSONBody(body []byte, v interface{}) error {
	// Handle cases where body might be a JSON string or direct JSON
	var rawData interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return err
	}

	// If it's a string, try to parse it as JSON
	if str, ok := rawData.(string); ok {
		return json.Unmarshal([]byte(str), v)
	}

	// Otherwise, unmarshal directly
	data, err := json.Marshal(rawData)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
func MakeGetRequest(url string, params map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetCurrentTimeISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05")
}
