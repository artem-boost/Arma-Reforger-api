package handlers

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"arma-reforger-api/models"

	"github.com/gin-gonic/gin"
)

// GetWorkshopTokenHandler форвардит POST-запрос к серверу, указанному в конфиге.
// Тело и заголовки запроса копируются, а ответ проксируется клиенту.
func GetWorkshopTokenHandler(c *gin.Context) {
	cfg := models.GetConfig()
	if cfg == nil || cfg.Workshop.URL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workshop URL not configured"})
		return
	}

	// Прочитаем тело исходного запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Создадим новый запрос к указанному в конфиге URL
	req, err := http.NewRequest("POST", cfg.Workshop.URL+"/GetWorkshopToken", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
		return
	}

	// Скопируем заголовки из оригинального запроса
	for k, vals := range c.Request.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	// Клиент с таймаутом
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to forward request"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read response from remote"})
		return
	}

	// Установим content-type из ответа и вернём тело и статус
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	c.Data(resp.StatusCode, contentType, respBody)
}

// GetUserTokenProxyHandler форвардит POST-запрос к серверу, указанному в конфиге.
// Тело и заголовки запроса копируются, а ответ проксируется клиенту.
func GetUserTokenProxyHandler(c *gin.Context) {
	cfg := models.GetConfig()
	if cfg == nil || cfg.Workshop.URL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workshop URL not configured"})
		return
	}

	// Прочитаем тело исходного запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Создадим новый запрос к указанному в конфиге URL
	req, err := http.NewRequest("POST", cfg.Workshop.URL+"/user/api/v2.0/game-identity/token", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
		return
	}

	// Скопируем заголовки из оригинального запроса
	for k, vals := range c.Request.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	// Клиент с таймаутом
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to forward request"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read response from remote"})
		return
	}

	// Установим content-type из ответа и вернём тело и статус
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	c.Data(resp.StatusCode, contentType, respBody)
}
