package handlers

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"arma-reforger-api/models"
	"arma-reforger-api/utils"
)

func AuthHandler(c *gin.Context) {
	var authReq models.AuthRequest
	if err := c.ShouldBindJSON(&authReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	bufferFromBase64, err := base64.StdEncoding.DecodeString(authReq.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 token"})
		return
	}
	hexString := hex.EncodeToString(bufferFromBase64)
	// Get Steam ticket data
	steamAuthURL := "https://api.steampowered.com/ISteamUserAuth/AuthenticateUserTicket/v1/"
	params := map[string]string{
		"key":    models.GetConfig().Steam.APIKey,
		"appid":  "480",
		"ticket": hexString,
	}

	authResp, err := utils.MakeGetRequest(steamAuthURL, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Steam"})
		return
	}

	var steamAuth models.SteamAuthResponse
	if err := json.Unmarshal(authResp, &steamAuth); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Steam response"})
		return
	}
	if steamAuth.Response.Params.SteamID == "" {
		log.Print("Steam return response with SteamID is empty")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SteamID is empty"})
		return
	}
	steamID := steamAuth.Response.Params.SteamID

	// Get Steam user info
	steamUserURL := "https://community.steam-api.com/ISteamUser/GetPlayerSummaries/v2/"
	userParams := map[string]string{
		"key":      models.GetConfig().Steam.APIKey,
		"steamids": steamID,
	}

	userResp, err := utils.MakeGetRequest(steamUserURL, userParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Steam"})
		return
	}

	var steamUser models.SteamUserResponse
	if err := json.Unmarshal(userResp, &steamUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Steam user response"})
		return
	}
	if len(steamUser.Response.Players) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Steam API return no players"})
		return
	}
	steamUsername := steamUser.Response.Players[0].PersonaName
	userID := uuid.New().String()
	accessToken := "eyJhbGciOiJSUzUxMiJ9.eyJpYXQiOjE3MjM5NDA3NDUsImV4cCI6MTcyMzk0NDM0NSwiaXNzIjoiZ2kiLCJhdWQiOiJnaSwgY2xpZW50LCBiaS1hY2NvdW50IiwiZ2lkIjoiYmNkM2NkMDctZjg3ZC00YzUwLTk3ZmItMzcyNWU5NGUzYTcxIiwiZ21lIjoicmVmb3JnZXIiLCJwbHQiOiJzdGVhbSJ9.INGYyPfKS2bkGk1nWLnydzczwHtHCycAUE5QRMHrL0f3nAIA3cv6uXVwHOUpqdEgDqdqo49YCTBE6BHam8MbWHQysilTX04e-Z2XXWX6YePIukQ6fjyH0xw1C_KKXzTOekbmlU-KCZ9dLi3D8vVC-4fkWwrL3czxpCclbwRxYQPOTmoTy5G-Fv3-U4edKET3a5-RyVMRsD5p0K_6wba3l6j8cET0SXH-5P46yxxyp1mUu76SdLT2nDDmEYdIgNWkWpXO-ONyxd0CJr_M3RQaTSIMF2r5A4gyMMpzlvF5kmnhOkiO0p1i1-1WAG21yrMrz6xM0DjAPLJF" +
		utils.GenerateRandomString(10, true)

	user := &models.User{
		ID:          userID,
		SteamID:     steamID,
		Username:    steamUsername,
		AccessToken: accessToken,
	}

	if err := models.CreateOrUpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	data := gin.H{
		"identityId":     userID,
		"accessToken":    accessToken,
		"accessTokenExp": 1912663541,
		"identity": gin.H{
			"id":   userID,
			"game": "reforger",
			"links": []gin.H{
				{
					"platform":   "steam",
					"platformId": steamID,
					"createdAt":  utils.GetCurrentTimeISO(),
				},
				{
					"platform":   "bi-account",
					"platformId": "51dcf826-17db-4f87-91d5-4e95cdb853cf",
					"createdAt":  "2024-03-18T09:18:12",
				},
			},
			"linkHistory": []interface{}{},
			"createdAt":   "2024-01-02T20:05:43",
			"updatedAt":   "2024-03-18T08:18:12",
		},
	}
	go utils.AuthOnOtherAPI(userID, authReq.Token)
	c.JSON(http.StatusOK, data)
}

func AcceptPlayerHandler(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	sessionTicket, ok := body["sessionTicket"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionTicket is required"})
		return
	}

	user, err := models.GetUserByTicket(sessionTicket[:64])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if user == nil {
		// Forward to Bohemia if user not found
		// Implementation for forwarding would go here
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	data := gin.H{
		"userProfile": gin.H{
			"userId":      user.ID,
			"username":    user.Username,
			"renameCount": -1,
			"currencies": gin.H{
				"HardCurrency": 0,
				"SoftCurrency": 0,
			},
			"countryCode":        "UA",
			"overallPlayTime":    66230,
			"tester":             false,
			"isDeveloperAccount": false,
			"rentedServers": gin.H{
				"entries":      []interface{}{},
				"visitedGames": []interface{}{},
			},
		},
		"character": gin.H{
			"id":      user.ID,
			"name":    user.Username,
			"version": 1711140949069791232,
			"data":    `{"m_aStats":[0.0,53337.69140625,43835.9453125,131749.46875,98400.2734375,14878.484375,0.0,50760.828125,14.0,106.0,3366.0,78.0,7.0,2.0,66.0,31465.36328125,19135.662109375,0.0,0.0,2.0,2.0,0.0,27604.794921875,19.0,13.0,4.0,0.0,0.0,0.0,2.0,0.0,6520919040.0,1.0,0.0,0.0,0.0,0.0,0.0,0.0]}`,
		},
		"sessionTicket": sessionTicket,
		"secret":        "8f5e73fcdfead3b2e79bae2fef52ed2b19ffa0272ba76459cbd534937bb497d09e5145e85ae93009689a9e4946d3af4f548d8830b93c24dc891519017875a954a53e46d2fee91952",
		"platformIdentities": gin.H{
			"biAccountId": "51dcf826-17db-4f87-91d5-4e95cdb853cf",
			"steamId":     user.SteamID,
		},
		"gameClientType":   "PLATFORM_PC",
		"platformUsername": user.Username,
	}

	c.JSON(http.StatusOK, data)
}

func JoinRoomHandler(c *gin.Context) {
	var joinData models.RoomJoinRequest
	if err := c.ShouldBindJSON(&joinData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	server, err := models.GetServerByID(joinData.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if server == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if server.Api_name == "local" {
		// Handle local server join
		if err := handleLocalJoin(c, server, &joinData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	} else {
		// Forward to remote API and return raw response
		if err := handleRemoteJoin(c, &joinData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward join request", "details": err.Error()})
		}
	}
}

// handleLocalJoin processes join requests for servers hosted on this API
func handleLocalJoin(c *gin.Context, server *models.Server, joinData *models.RoomJoinRequest) error {
	var serverData map[string]interface{}
	if err := json.Unmarshal(server.Data, &serverData); err != nil {
		return fmt.Errorf("invalid server data: %w", err)
	}

	serverAddress, ok := serverData["hostAddress"].(string)
	if !ok {
		return fmt.Errorf("invalid server address")
	}

	sessionTicket := utils.GenerateRandomString(64, true)
	if err := models.UpdateUserTicket(joinData.AccessToken, sessionTicket); err != nil {
		return fmt.Errorf("failed to update user ticket: %w", err)
	}

	data := gin.H{
		"sessionTicket": sessionTicket,
		"secret":        "8f5e73fcdfead3b2e79bae2fef52ed2b19ffa0272ba76459cbd534937bb497d09e5145e85ae93009689a9e4946d3af4f548d8830b93c24dc891519017875a954a53e46d2fee91952",
		"address":       serverAddress,
		"inviteToken":   "dVABIFt7UE1lX2hQEUwQWEZVUgocJzECXmZcfnBCUBR/Y2IpAnNWIV9FMA==",
		"joinResult":    "Join",
	}

	c.JSON(http.StatusOK, data)
	return nil
}

// handleRemoteJoin forwards join requests to external APIs
func handleRemoteJoin(c *gin.Context, joinData *models.RoomJoinRequest) error {
	// Re-encode the original request body
	reqBody, err := json.Marshal(joinData)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	resp, err := utils.JoinOnOtherAPI(joinData.RoomID, reqBody)
	if err != nil {
		return err
	}

	// Return raw bytes from remote API
	c.Data(http.StatusOK, "application/json", resp)
	return nil
}
