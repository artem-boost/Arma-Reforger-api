package utils

import (
	"arma-reforger-api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var ServersAPI []models.ServerAPI

func SetServersAPI(servers []models.ServerAPI) {
	ServersAPI = servers
}
func RequestForAPI(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	var bodyreq = `{"directJoinCode":"","hostAddress":"","order":"PlayerCount","scenarioId":"","includePing":1,"text":"","ascendent":false,"gameClientFilter":"AnyCompatible","accessToken":"eyJhbGciOiJSUzUxMiJ9.eyJpYXQiOjE3MjM5NDA3NDUsImV4cCI6MTcyMzk0NDM0NSwiaXNzIjoiZ2kiLCJhdWQiOiJnaSwgY2xpZW50LCBiaS1hY2NvdW50IiwiZ2lkIjoiYmNkM2NkMDctZjg3ZC00YzUwLTk3ZmItMzcyNWU5NGUzYTcxIiwiZ21lIjoicmVmb3JnZXIiLCJwbHQiOiJzdGVhbSJ9.INGYyPfKS2bkGk1nWLnydzczwHtHCycAUE5QRMHrL0f3nAIA3cv6uXVwHOUpqdEgDqdqo49YCTBE6BHam8MbWHQysilTX04e-Z2XXWX6YePIukQ6fjyH0xw1C_KKXzTOekbmlU-KCZ9dLi3D8vVC-4fkWwrL3czxpCclbwRxYQPOTmoTy5G-Fv3-U4edKET3a5-RyVMRsD5p0K_6wba3l6j8cET0SXH-5P46yxxyp1mUu76SdLT2nDDmEYdIgNWkWpXO-ONyxd0CJr_M3RQaTSIMF2r5A4gyMMpzlvF5kmnhOkiO0p1i1-1WAG21yrMrz6xM0DjAPLJF6shw5h4sdu","clientVersion":"1.6.0","platformId":"ReforgerSteam","gameClientType":"PLATFORM_PC","lightweight":true,"from":0,"limit":50,"pingValues":[{"pingSiteId":"tokyo","value":0.0},{"pingSiteId":"los_angeles","value":0.0},{"pingSiteId":"miami","value":0.0},{"pingSiteId":"new_york","value":0.0},{"pingSiteId":"singapore","value":0.0},{"pingSiteId":"frankfurt","value":0.0},{"pingSiteId":"london","value":0.0},{"pingSiteId":"sydney","value":0.0}]}`
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(bodyreq))
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Arma Reforger/1.6.0.54 (Client; Windows)")
	req.Header.Add("Content-Type", "application/json")

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
func GetRooms() []models.Server {
	var Rooms []models.Server

	var roomsdata struct {
		Rooms []map[string]interface{} `json:"rooms"`
	}

	for _, serverapi := range ServersAPI {
		if !serverapi.Enable {
			continue
		}
		response, err := RequestForAPI(serverapi.URL + "/game-api/api/v1.0/lobby/rooms/search")
		if err != nil {
			continue
		}
		json.Unmarshal(response, &roomsdata)
		for _, room := range roomsdata.Rooms {
			var server models.Server
			server.ID = uuid.New().String()
			server.ServerID = room["id"].(string)
			server.IsLicense = true
			server.LastUpdate = time.Now()
			server.PlayerCount = int(room["playerCount"].(float64))
			server.Api_name = serverapi.Name
			data, _ := json.Marshal(room)
			server.Data = json.RawMessage(data)
			Rooms = append(Rooms, server)
		}

	}
	return Rooms
}
func CreateRooms(rooms []models.Server) {
	for _, room := range rooms {
		models.CreateOrUpdateServer(&room)
	}
}
