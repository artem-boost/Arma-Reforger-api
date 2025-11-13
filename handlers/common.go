package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "pong"})
}

func NewsHandler(c *gin.Context) {
	news := gin.H{
		"items": []gin.H{
			{
				"date":     "08 Августа 2024",
				"excerpt":  "Начало тестирования нового API",
				"category": "Development",
				"slug":     "update-august-08-2024",
				"title":    "1.2.0.102 Update",
				"coverImage": gin.H{
					"src": "https://cms-cdn.bistudio.com/cms-static--reforger/images/08f81027-c102-4a4f-b9e7-fe12e8e6e8c2-NEWS%201280x720.jpg",
				},
				"fullUrl": "https://youtu.be/dQw4w9WgXcQ?si=zY34xDJ__8psHZ3i",
			},
		},
	}
	c.JSON(http.StatusOK, news)
}

func BlockListHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"blockList": gin.H{
			"entries":    []interface{}{},
			"totalCount": 0,
			"page": gin.H{
				"offset": 0,
				"limit":  16,
			},
		},
	})
}

func ListPlayersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connectedPlayers": []interface{}{},
		"queuePlayers":     []interface{}{},
	})
}

func SendTdEventsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func GetPingSitesHandler(c *gin.Context) {
	pingSites := gin.H{
		"pingSites": []gin.H{
			{
				"id":        "frankfurt",
				"address":   "ping-location-de.nitrado.net",
				"ipAddress": "31.214.130.69",
				"location": gin.H{
					"latitude":  51.29930114746094,
					"longitude": 9.491000175476074,
				},
				"mappedRegions": []string{"eu-ffm"},
			},
			// Add other ping sites as needed
		},
	}
	c.JSON(http.StatusOK, pingSites)
}

func GetLobbyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"rooms": []interface{}{}})
}

func WorldHandler(c *gin.Context) {
	world := gin.H{
		"version": "BanSettings",
		"gameData": gin.H{
			"BanSettings": gin.H{
				"m_sDesc": "Ban Logic Cleansweep",
				"m_BanSettings": gin.H{
					"m_fScoreThreshold":          10.0,
					"m_fScoreDecreasePerMinute":  0.2,
					"m_fScoreMultiplier":         0.2,
					"m_fAccelerationMin":         1.0,
					"m_fAccelerationMax":         6.0,
					"m_fBanEvaluationLight":      0.8,
					"m_fBanEvaluationHeavy":      1.0,
					"m_fCrimePtFriendKill":       1.0,
					"m_fCrimePtTeamKill":         0.7,
					"m_fQualityTimeTemp":         1.0,
					"m_bVotingSuggestionEnabled": 0,
				},
			},
		},
	}
	c.JSON(http.StatusOK, world)
}

func DummyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func SessionLoginHandler(c *gin.Context) {
	data := gin.H{
		"userProfile": gin.H{
			"userId":      "1bde4705-34fd-489d-a7fe-93f3a2f5aefc",
			"username":    "",
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
		"worldVersion":             "BanSettings",
		"ipAddress":                "91.219.235.155",
		"pendingMicroTransactions": []interface{}{},
		"compatibleGameVersions":   []string{"1.1.0.42"},
		"notifications":            []interface{}{},
		"sessionId":                "4105b12d-6873-4f8e-9dd3-36d638fdc455",
	}
	c.JSON(http.StatusOK, data)
}
