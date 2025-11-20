package utils

import (
	"arma-reforger-api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var ServersAPI []models.ServerAPI
var APIclient = &http.Client{Timeout: 30 * time.Second}

func SetServersAPI(servers []models.ServerAPI) {
	ServersAPI = servers
}
func SearchRequestForAPI(url string) ([]byte, error) {
	var bodyreq = `{"directJoinCode":"","hostAddress":"","order":"PlayerCount","scenarioId":"","includePing":1,"text":"","ascendent":false,"gameClientFilter":"AnyCompatible","accessToken":"eyJhbGciOiJSUzUxMiJ9.eyJpYXQiOjE3MjM5NDA3NDUsImV4cCI6MTcyMzk0NDM0NSwiaXNzIjoiZ2kiLCJhdWQiOiJnaSwgY2xpZW50LCBiaS1hY2NvdW50IiwiZ2lkIjoiYmNkM2NkMDctZjg3ZC00YzUwLTk3ZmItMzcyNWU5NGUzYTcxIiwiZ21lIjoicmVmb3JnZXIiLCJwbHQiOiJzdGVhbSJ9.INGYyPfKS2bkGk1nWLnydzczwHtHCycAUE5QRMHrL0f3nAIA3cv6uXVwHOUpqdEgDqdqo49YCTBE6BHam8MbWHQysilTX04e-Z2XXWX6YePIukQ6fjyH0xw1C_KKXzTOekbmlU-KCZ9dLi3D8vVC-4fkWwrL3czxpCclbwRxYQPOTmoTy5G-Fv3-U4edKET3a5-RyVMRsD5p0K_6wba3l6j8cET0SXH-5P46yxxyp1mUu76SdLT2nDDmEYdIgNWkWpXO-ONyxd0CJr_M3RQaTSIMF2r5A4gyMMpzlvF5kmnhOkiO0p1i1-1WAG21yrMrz6xM0DjAPLJF6shw5h4sdu","clientVersion":"1.6.0","platformId":"ReforgerSteam","gameClientType":"PLATFORM_PC","lightweight":true,"from":0,"limit":50,"pingValues":[{"pingSiteId":"tokyo","value":0.0},{"pingSiteId":"los_angeles","value":0.0},{"pingSiteId":"miami","value":0.0},{"pingSiteId":"new_york","value":0.0},{"pingSiteId":"singapore","value":0.0},{"pingSiteId":"frankfurt","value":0.0},{"pingSiteId":"london","value":0.0},{"pingSiteId":"sydney","value":0.0}]}`
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(bodyreq))
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Arma Reforger/1.6.0.54 (Client; Windows)")
	req.Header.Add("Content-Type", "application/json")

	resp, err := APIclient.Do(req)
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

func AuthRequest(url string, Steamtoken string) (string, error) {
	var AuthRequestBody models.AuthRequest
	AuthRequestBody.Platform = "Arma Reforger PC"
	AuthRequestBody.Token = Steamtoken
	AuthRequestBody.PlatformOpts.AppID = "480"
	body, err := json.Marshal(&AuthRequestBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "Arma Reforger/1.6.0.54 (Client; Windows)")
	req.Header.Add("Content-Type", "application/json")

	resp, err := APIclient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try to extract access token from various possible response shapes
	var parsed map[string]interface{}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse auth response: %w", err)
	}

	// common location for access token
	if at, ok := parsed["accessToken"].(string); ok && at != "" {
		return at, nil
	}
	// If token not found, return whole response as error for debugging
	return "", fmt.Errorf("access token not found in response: %s", string(respBody))

}

func AuthOnOtherAPI(userID, Steamtoken string) {
	for _, serverapi := range ServersAPI {
		go func(apiURL, apiName string) {
			token, err := AuthRequest(apiURL+"/game-identity/api/v1.1/identities/reforger/auth?include=profile", Steamtoken)
			if err != nil {
				log.Printf("Auth on %s failed: %v\n", apiURL, err)
				return
			}

			// Save received token to DB associated with userID and API name
			if err := models.CreateOrUpdateAccessToken(userID, token, apiName); err != nil {
				log.Printf("Failed to save access token for user %s on %s: %v\n", userID, apiName, err)
				return
			}
		}(serverapi.URL, serverapi.Name)
	}
}
func JoinOnOtherAPI(roomid string, reqBody []byte) ([]byte, error) {
	// Parse incoming request to extract accessToken
	var joinReq models.RoomJoinRequest
	if err := json.Unmarshal(reqBody, &joinReq); err != nil {
		return nil, fmt.Errorf("failed to parse join request: %w", err)
	}

	if joinReq.AccessToken == "" {
		return nil, fmt.Errorf("missing accessToken in join request")
	}

	// Get userID from local user by accessToken
	user, err := models.GetUserByAccessToken(joinReq.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user by accessToken: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found for accessToken")
	}

	// Resolve target API
	apiName, err := models.GetAPINameByRoom(roomid)
	if err != nil {
		return nil, fmt.Errorf("failed to get API name for room %s: %w", roomid, err)
	}

	var targetAPI *models.ServerAPI
	for i := range ServersAPI {
		if ServersAPI[i].Name == apiName {
			targetAPI = &ServersAPI[i]
			break
		}
	}
	if targetAPI == nil {
		return nil, fmt.Errorf("no API configuration found for %s", apiName)
	}

	// Get remote accessToken for this user on the target API
	remoteToken, err := models.GetAccessTokenByUserAndAPI(user.ID, apiName)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote access token: %w", err)
	}
	if remoteToken == "" {
		return nil, fmt.Errorf("no remote access token found for user %s on API %s", user.ID, apiName)
	}

	// Replace accessToken in the request
	joinReq.AccessToken = remoteToken
	newBody, err := json.Marshal(joinReq)
	if err != nil {
		return nil, fmt.Errorf("failed to rebuild request with remote token: %w", err)
	}

	// Forward to remote API
	req, err := http.NewRequest("POST", targetAPI.URL+"/game-api/api/v1.0/lobby/rooms/join", bytes.NewReader(newBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "Arma Reforger/1.6.0.68 (Client; Windows)")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := APIclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote API returned status %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func GetRooms() []models.Server {
	var allRooms []models.Server

	// Create a channel to collect results from goroutines
	type apiResult struct {
		rooms []models.Server
		err   error
	}
	resultCh := make(chan apiResult, len(ServersAPI))

	// Launch a goroutine for each enabled API
	for _, serverapi := range ServersAPI {
		if !serverapi.Enable {
			continue
		}
		go func(api models.ServerAPI) {
			var result apiResult
			defer func() { resultCh <- result }()

			response, err := SearchRequestForAPI(api.URL + "/game-api/api/v1.0/lobby/rooms/search")
			if err != nil {
				result.err = fmt.Errorf("API %s failed: %w", api.Name, err)
				return
			}

			var roomsdata struct {
				Rooms []map[string]interface{} `json:"rooms"`
			}
			if err := json.Unmarshal(response, &roomsdata); err != nil {
				result.err = fmt.Errorf("API %s JSON parse error: %w", api.Name, err)
				return
			}

			var servers []models.Server
			for _, room := range roomsdata.Rooms {
				server := models.Server{
					ID:          uuid.New().String(),
					ServerID:    room["id"].(string),
					IsLicense:   true,
					LastUpdate:  time.Now(),
					PlayerCount: int(room["playerCount"].(float64)),
					Api_name:    api.Name,
				}
				data, _ := json.Marshal(room)
				server.Data = json.RawMessage(data)
				servers = append(servers, server)
			}
			result.rooms = servers
		}(serverapi)
	}

	// Collect results from all goroutines
	for range ServersAPI {
		res := <-resultCh
		if res.err == nil {
			allRooms = append(allRooms, res.rooms...)
		} else {
			// Log error or handle as needed
			fmt.Printf("Warning: %v\n", res.err)
		}
	}

	return allRooms
}
func CreateRooms(rooms []models.Server) {
	for _, room := range rooms {
		models.CreateOrUpdateServer(&room)
	}
}
