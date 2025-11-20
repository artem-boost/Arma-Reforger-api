package models

import (
	"encoding/json"
	"time"
)

type Config struct {
	API struct {
		PORT             int  `json:"PORT"`
		ActivateWorkshop bool `json:"ActivateWorkshop"`
	} `json:"API"`
	Workshop struct {
		IP   string `json:"IP"`
		PORT int    `json:"PORT"`
		KEY  string `json:"KEY"`
	} `json:"Workshop"`
	DB struct {
		Path string `json:"Path"`
	} `json:"DB"`
	Steam struct {
		APIKey string `json:"APIKey"`
	} `json:"Steam"`
	ServersAPI []ServerAPI `json:"ServersAPI"`
}
type ServerAPI struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
}
type UserToken struct {
	AccessToken string    `json:"accessToken"`
	ApiName     string    `json:"apiName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type User struct {
	ID          string      `json:"id"`
	SteamID     string      `json:"steamid"`
	Username    string      `json:"username"`
	AccessToken string      `json:"accessToken"`
	Ticket      string      `json:"ticket"`
	Tokens      []UserToken `json:"tokens,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Server struct {
	ID          string          `json:"id"`
	ServerID    string          `json:"serverID"`
	Data        json.RawMessage `json:"data"`
	Password    string          `json:"password"`
	IsLicense   bool            `json:"isLicense"`
	LastUpdate  time.Time       `json:"lastUpdate"`
	PlayerCount int             `json:"playerCount"`
	Api_name    string          `json:"api_name"`
}

type AuthRequest struct {
	Platform     string `json:"platform"`
	Token        string `json:"token"`
	PlatformOpts struct {
		AppID string `json:"appId"`
	} `json:"platformOpts"`
}

type RoomJoinRequest struct {
	AccessToken      string `json:"accessToken"`
	ClientVersion    string `json:"clientVersion"`
	PlatformID       string `json:"platformId"`
	RoomID           string `json:"roomId"`
	PlatformUsername string `json:"platformUsername"`
	GameClientType   string `json:"gameClientType"`
}

type SteamAuthResponse struct {
	Response struct {
		Params struct {
			SteamID string `json:"steamid"`
		} `json:"params"`
	} `json:"response"`
}

type SteamUserResponse struct {
	Response struct {
		Players []struct {
			PersonaName string `json:"personaname"`
		} `json:"players"`
	} `json:"response"`
}

type WorkshopListRequest struct {
	AssetIDs []string `json:"assetIds"`
}

type ServerRegisterRequest struct {
	Name                     string        `json:"name"`
	ScenarioID               string        `json:"scenarioId"`
	ScenarioName             string        `json:"scenarioName"`
	HostedScenarioModID      string        `json:"hostedScenarioModId"`
	PlayerCountLimit         int           `json:"playerCountLimit"`
	SupportedGameClientTypes []string      `json:"supportedGameClientTypes"`
	Mods                     []interface{} `json:"mods"`
	Tags                     []string      `json:"tags"`
	GameMode                 string        `json:"gameMode"`
	AccessToken              string        `json:"accessToken"`
	HostAddress              string        `json:"hostAddress"`
	GameVersion              string        `json:"gameVersion"`
	AutoJoinable             bool          `json:"autoJoinable"`
	Password                 string        `json:"password"`
	DedicatedServerID        string        `json:"dedicatedServerId"`
}
