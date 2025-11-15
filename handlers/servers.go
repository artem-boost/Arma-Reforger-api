package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"arma-reforger-api/models"
	"arma-reforger-api/utils"
)

func SearchServersHandler(c *gin.Context) {
	servers, err := models.GetAllServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}

	rooms := make([]map[string]interface{}, 0)
	var serverscount int
	for _, server := range servers {
		var data map[string]interface{}
		if err := json.Unmarshal(server.Data, &data); err == nil {
			rooms = append(rooms, data)
			serverscount += 1
		}
	}

	data := gin.H{
		"rooms":      rooms,
		"searchFrom": 0,
		"totalCount": serverscount,
	}

	c.JSON(http.StatusOK, data)
}

func GetRoomsByIDsHandler(c *gin.Context) {
	var request struct {
		RoomIDs []string `json:"roomIds"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(request.RoomIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"rooms":      []interface{}{},
			"searchFrom": 0,
			"totalCount": 0,
		})
		return
	}

	servers, err := models.GetServersByIDs(request.RoomIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}

	rooms := make([]map[string]interface{}, 0)
	for _, server := range servers {
		var data map[string]interface{}
		if err := json.Unmarshal(server.Data, &data); err == nil {
			rooms = append(rooms, data)
		}
	}

	data := gin.H{
		"rooms":      rooms,
		"searchFrom": 0,
		"totalCount": len(rooms),
	}

	c.JSON(http.StatusOK, data)
}
func RegisterUnmanagedServerHandler(c *gin.Context) {
	var body struct {
		Name                     string        `json:"name"`
		ScenarioID               string        `json:"scenarioId"`
		HostedScenarioModID      string        `json:"hostedScenarioModId"`
		PlayerCountLimit         int           `json:"playerCountLimit"`
		SupportedGameClientTypes []string      `json:"supportedGameClientTypes"`
		Mods                     []interface{} `json:"mods"`
		Tags                     []string      `json:"tags"`
		GameMode                 string        `json:"gameMode"`
		AccessToken              string        `json:"accessToken"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var serverID = uuid.New().String()

	hostServerID := uuid.New().String()

	data := gin.H{
		"dsConfig": gin.H{
			"providerServerId":  serverID,
			"dedicatedServerId": hostServerID,
			"region":            "",
			"game": gin.H{
				"name":                     body.Name,
				"scenarioId":               body.ScenarioID,
				"hostedScenarioModId":      body.HostedScenarioModID,
				"playerCountLimit":         body.PlayerCountLimit,
				"gameNumber":               0,
				"autoJoinable":             false,
				"visible":                  true,
				"supportedGameClientTypes": body.SupportedGameClientTypes,
				"gameInstanceFiles": gin.H{
					"fileReferences": []interface{}{},
				},
				"mods":     body.Mods,
				"tags":     body.Tags,
				"gameMode": body.GameMode,
			},
		},
		"ownerToken": body.AccessToken,
	}

	// Сохраняем сервер в базе данных с пустыми данными
	serverData, _ := json.Marshal(gin.H{})
	server := &models.Server{
		ID:          uuid.New().String(),
		ServerID:    hostServerID,
		Data:        serverData,
		Password:    "",
		IsLicense:   false,
		PlayerCount: 0,
	}

	if err := models.CreateOrUpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save server"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func RegisterServerHandler(c *gin.Context) {
	var body models.ServerRegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	serverID := uuid.New().String()
	hostServerID := uuid.New().String()

	// Create server data
	serverData := map[string]interface{}{
		"id":                       hostServerID,
		"scenarioId":               body.ScenarioID,
		"name":                     body.Name,
		"scenarioVersion":          "",
		"scenarioName":             body.ScenarioName,
		"region":                   "n/a",
		"gameVersion":              body.GameVersion,
		"hostType":                 "CommunityDs",
		"dedicated":                true,
		"official":                 true,
		"joinable":                 true,
		"visible":                  true,
		"passwordProtected":        body.Password != "",
		"created":                  1712841158268,
		"updated":                  1712841158268,
		"hostAddress":              body.HostAddress,
		"hostUserId":               body.DedicatedServerID,
		"playerCountLimit":         body.PlayerCountLimit,
		"playerCount":              0,
		"autoJoinable":             body.AutoJoinable,
		"directJoinCode":           "0622875052",
		"supportedGameClientTypes": body.SupportedGameClientTypes,
		"dsLaunchTimestamp":        1712841157982,
		"dsProviderServerId":       hostServerID,
		"mods":                     body.Mods,
		"battlEye":                 false,
		"favorite":                 false,
		"gameMode":                 body.GameMode,
		"pingSiteId":               "frankfurt",
		"platformName":             "Windows",
		"runtimeStats": map[string]interface{}{
			"needRestart": false,
		},
		"sessionId": "20240411131235-0000207476a1",
	}

	dataJSON, err := json.Marshal(serverData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal server data"})
		return
	}

	server := &models.Server{
		ID:          uuid.New().String(),
		ServerID:    serverID,
		Data:        dataJSON,
		Password:    body.Password,
		IsLicense:   false,
		PlayerCount: 0,
	}

	if err := models.CreateOrUpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save server"})
		return
	}

	response := gin.H{
		"roomId": hostServerID,
		"mpRoom": serverData,
	}

	c.JSON(http.StatusOK, response)
}

func RemoveServerHandler(c *gin.Context) {
	var body struct {
		DedicatedServerID string `json:"dedicatedServerId"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := models.DeleteServer(body.DedicatedServerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func HeartBeatHandler(c *gin.Context) {
	var body struct {
		ID      string        `json:"id"`
		Players []interface{} `json:"players"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	server, err := models.GetServerByID(body.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if server == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "NotFound Server"})
		return
	}

	// Update player count
	server.PlayerCount = len(body.Players)
	if err := models.CreateOrUpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	data := gin.H{
		"currentWorldVersion": "BanSettings",
		"gameEvents":          []interface{}{},
		"status":              "OK",
	}

	c.JSON(http.StatusOK, data)
}

func VerifyPasswordHandler(c *gin.Context) {
	var rawData string
	if err := c.ShouldBindJSON(&rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var body struct {
		RoomID   string `json:"roomId"`
		Password string `json:"password"`
	}

	if err := utils.ParseJSONBody([]byte(rawData), &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request data"})
		return
	}

	server, err := models.GetServerByRoomID(body.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if server == nil || server.Password != body.Password {
		errorResponse := gin.H{
			"code":      403,
			"errorType": "PasswordMismatch",
			"apiCode":   "PasswordMismatch",
			"message":   "Password mismatch",
			"uid":       "0ec0a582-46bc-4748-9743-585f757e06d0",
		}
		c.JSON(403, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// Stub handlers for unimplemented endpoints
func ListActiveBansHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"activeBans": gin.H{
			"entries":    []interface{}{},
			"totalCount": 0,
			"page": gin.H{
				"offset": 0,
				"limit":  10,
			},
		},
	})
}
func APIStatusOk(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func RemovePlayerHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func CreateBanHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func RemoveBansHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
