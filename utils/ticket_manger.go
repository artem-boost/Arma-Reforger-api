package utils

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"arma-reforger-api/models"
)

type TicketManager struct {
	config *models.Config
	ticket string
	token  string
}

func NewTicketManager(config *models.Config) *TicketManager {
	return &TicketManager{
		config: config,
	}
}

func (tm *TicketManager) SetTicket(newTicket string) {
	tm.ticket = newTicket
}

func (tm *TicketManager) SetToken(newToken string) string {
	tm.token = newToken
	return tm.token
}

func (tm *TicketManager) GetToken() string {
	return tm.token
}

func (tm *TicketManager) GetTicketFromServer(host string, port int) (string, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(tm.config.Workshop.KEY))
	if err != nil {
		return "", fmt.Errorf("failed to send key: %v", err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read ticket: %v", err)
	}

	ticketHex := string(buffer[:n])
	log.Printf("Received ticket (hex): %s", ticketHex)
	return ticketHex, nil
}

func hexToBase64(hexString string) (string, error) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %v", err)
	}
	return string(data), nil
}

// Добавленная функция makeHTTPRequest
func makeHTTPRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (tm *TicketManager) GetAccessToken() (string, error) {
	url := "https://api-ar-id.bistudio.com/game-identity/api/v1.1/identities/reforger/auth?include=profile"

	headers := map[string]string{
		"Content-Type": "application/json",
		"user-agent":   "Arma Reforger/1.6.0.68 (Client; Windows)",
	}

	requestData := map[string]interface{}{
		"platform": "steam",
		"token":    tm.ticket,
		"platformOpts": map[string]interface{}{
			"appId": "1874880",
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := makeHTTPRequest(ctx, "POST", url, jsonData, headers)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	accessToken, ok := result["accessToken"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessToken, nil
}

func (tm *TicketManager) FetchTicketPeriodically() {
	if !tm.config.API.ActivateWorkshop {
		log.Println("Workshop is deactivated, set static token")
		tm.SetToken("eyJhbGciOiJSUzUxMiJ9.eyJpYXQiOjE3MjM5NDA3NDUsImV4cCI6MTcyMzk0NDM0NSwiaXNzIjoiZ2kiLCJhdWQiOiJnaSwgY2xpZW50LCBiaS1hY2NvdW50IiwiZ2lkIjoiYmNkM2NkMDctZjg3ZC00YzUwLTk3ZmItMzcyNWU5NGUzYTcxIiwiZ21lIjoicmVmb3JnZXIiLCJwbHQiOiJzdGVhbSJ9.INGYyPfKS2bkGk1nWLnydzczwHtHCycAUE5QRMHrL0f3nAIA3cv6uXVwHOUpqdEgDqdqo49YCTBE6BHam8MbWHQysilTX04e-Z2XXWX6YePIukQ6fjyH0xw1C_KKXzTOekbmlU-KCZ9dLi3D8vVC-4fkWwrL3czxpCclbwRxYQPOTmoTy5G-Fv3-U4edKET3a5-RyVMRsD5p0K_6wba3l6j8cET0SXH-5P46yxxyp1mUu76SdLT2nDDmEYdIgNWkWpXO-ONyxd0CJr_M3RQaTSIMF2r5A4gyMMpzlvF5kmnhOkiO0p1i1-1WAG21yrMrz6xM0DjAPLJFAAAAAAAAAA")
		return
	}

	log.Println("Workshop is activated")
	log.Println("Start auth ticket")

	tm.fetchTicket()

	ticker := time.NewTicker(30 * 50 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tm.fetchTicket()
	}
}

func (tm *TicketManager) fetchTicket() {
	ticketHex, err := tm.GetTicketFromServer(tm.config.Workshop.IP, tm.config.Workshop.PORT)
	if err != nil {
		log.Printf("Failed to get ticket: %v", err)
		return
	}

	ticketBase64, err := hexToBase64(ticketHex)
	if err != nil {
		log.Printf("Failed to convert ticket to base64: %v", err)
		return
	}

	tm.SetTicket(ticketBase64)
	log.Printf("Received ticket (base64): %s", ticketBase64)

	token, err := tm.GetAccessToken()
	if err != nil {
		log.Printf("Failed to get access token: %v", err)
		return
	}

	log.Printf("Received token: %s", token)
	tm.SetToken(token)
}
