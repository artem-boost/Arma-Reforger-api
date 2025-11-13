package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

var DB *sql.DB
var globalConfig *Config

func InitDatabase(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	// Create indexes
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	usersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        steam_id TEXT UNIQUE NOT NULL,
        username TEXT NOT NULL,
        access_token TEXT NOT NULL,
        ticket TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`

	serversTable := `
    CREATE TABLE IF NOT EXISTS servers (
        id TEXT PRIMARY KEY,
        server_id TEXT UNIQUE NOT NULL,
        data TEXT NOT NULL,
        password TEXT,
        is_license BOOLEAN DEFAULT FALSE,
        player_count INTEGER DEFAULT 0,
        last_update DATETIME DEFAULT CURRENT_TIMESTAMP
    )`

	if _, err := DB.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if _, err := DB.Exec(serversTable); err != nil {
		return fmt.Errorf("failed to create servers table: %v", err)
	}

	return nil
}

func createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_steam_id ON users(steam_id)",
		"CREATE INDEX IF NOT EXISTS idx_users_access_token ON users(access_token)",
		"CREATE INDEX IF NOT EXISTS idx_users_ticket ON users(ticket)",
		"CREATE INDEX IF NOT EXISTS idx_servers_server_id ON servers(server_id)",
		"CREATE INDEX IF NOT EXISTS idx_servers_last_update ON servers(last_update)",
	}

	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}

// SetConfig устанавливает глобальную конфигурацию
func SetConfig(config *Config) {
	globalConfig = config
}

// GetConfig возвращает глобальную конфигурацию
func GetConfig() *Config {
	return globalConfig
}

// GetUserByTicket ищет пользователя по session ticket
func GetUserByTicket(ticket string) (*User, error) {
	query := "SELECT id, steam_id, username, access_token, ticket, created_at, updated_at FROM users WHERE ticket = ?"
	ticket = ticket[:64]
	user := &User{}
	err := DB.QueryRow(query, ticket).Scan(
		&user.ID, &user.SteamID, &user.Username, &user.AccessToken, &user.Ticket,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}

// GetUserByAccessToken ищет пользователя по access token
func GetUserByAccessToken(accessToken string) (*User, error) {
	query := "SELECT id, steam_id, username, access_token, ticket, created_at, updated_at FROM users WHERE access_token = ?"

	user := &User{}
	err := DB.QueryRow(query, accessToken).Scan(
		&user.ID, &user.SteamID, &user.Username, &user.AccessToken, &user.Ticket,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}

// UpdateUserTicket обновляет ticket пользователя
func UpdateUserTicket(accessToken, ticket string) error {
	query := "UPDATE users SET ticket = ?, updated_at = CURRENT_TIMESTAMP WHERE access_token = ?"
	_, err := DB.Exec(query, ticket, accessToken)
	return err
}

// CreateOrUpdateUser создает или обновляет пользователя
func CreateOrUpdateUser(user *User) error {
	query := `
    INSERT OR REPLACE INTO users (id, steam_id, username, access_token, ticket, updated_at)
    VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := DB.Exec(query, user.ID, user.SteamID, user.Username, user.AccessToken, user.Ticket)
	return err
}

// GetServersByIDs возвращает серверы по списку ID комнат
func GetServersByIDs(roomIDs []string) ([]Server, error) {
	if len(roomIDs) == 0 {
		return []Server{}, nil
	}

	// Создаем плейсхолдеры для SQL запроса
	placeholders := make([]string, len(roomIDs))
	args := make([]interface{}, len(roomIDs))
	for i, id := range roomIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT id, server_id, data, password, is_license, player_count, last_update 
        FROM servers 
        WHERE json_extract(data, '$.id') IN (%s)`,
		strings.Join(placeholders, ","))

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var server Server
		var dataJSON string

		err := rows.Scan(
			&server.ID, &server.ServerID, &dataJSON, &server.Password,
			&server.IsLicense, &server.PlayerCount, &server.LastUpdate,
		)
		if err != nil {
			return nil, err
		}

		// Преобразуем JSON строку в json.RawMessage
		server.Data = json.RawMessage(dataJSON)
		servers = append(servers, server)
	}

	return servers, nil
}

func CreateOrUpdateServer(server *Server) error {
	query := `
    INSERT OR REPLACE INTO servers (id, server_id, data, password, is_license, player_count, last_update)
    VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := DB.Exec(query, server.ID, server.ServerID, string(server.Data), server.Password,
		server.IsLicense, server.PlayerCount)
	return err
}

// GetAllServers возвращает все серверы
func GetAllServers() ([]Server, error) {
	query := "SELECT id, server_id, data, password, is_license, player_count, last_update FROM servers"

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var server Server
		var dataJSON string

		err := rows.Scan(
			&server.ID, &server.ServerID, &dataJSON, &server.Password,
			&server.IsLicense, &server.PlayerCount, &server.LastUpdate,
		)
		if err != nil {
			return nil, err
		}

		server.Data = json.RawMessage(dataJSON)
		servers = append(servers, server)
	}

	return servers, nil
}

// GetServerByRoomID возвращает сервер по ID комнаты
func GetServerByRoomID(roomID string) (*Server, error) {
	query := "SELECT id, server_id, data, password, is_license, player_count, last_update FROM servers WHERE json_extract(data, '$.id') = ?"

	var server Server
	var dataJSON string

	err := DB.QueryRow(query, roomID).Scan(
		&server.ID, &server.ServerID, &dataJSON, &server.Password,
		&server.IsLicense, &server.PlayerCount, &server.LastUpdate,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	server.Data = json.RawMessage(dataJSON)
	return &server, nil
}

// GetServerByID возвращает сервер по его ID
func GetServerByID(serverID string) (*Server, error) {
	query := "SELECT id, server_id, data, password, is_license, player_count, last_update FROM servers WHERE server_id = ?"

	var server Server
	var dataJSON string

	err := DB.QueryRow(query, serverID).Scan(
		&server.ID, &server.ServerID, &dataJSON, &server.Password,
		&server.IsLicense, &server.PlayerCount, &server.LastUpdate,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	server.Data = json.RawMessage(dataJSON)
	return &server, nil
}

// DeleteServer удаляет сервер по ID
func DeleteServer(serverID string) error {
	query := "DELETE FROM servers WHERE server_id = ?"
	_, err := DB.Exec(query, serverID)
	return err
}

// CleanupOldServers удаляет старые серверы
func CleanupOldServers() error {
	query := "DELETE FROM servers WHERE last_update < datetime('now', '-3 minutes')"
	_, err := DB.Exec(query)
	return err
}
