package main

import (
	"errors"
	"net"
	"strings"
	"time"
)

type Server struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	IPAddress   string    `json:"ip_address"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ServerInput struct {
	Name        string `json:"name"`
	IPAddress   string `json:"ip_address"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
}

func (r *ServerInput) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return errors.New("имя сервера не может быть пустым")
	}
	if len(r.Name) > 255 {
		return errors.New("имя сервера не должно превышать 255 символов")
	}

	r.IPAddress = strings.TrimSpace(r.IPAddress)
	if r.IPAddress == "" {
		return errors.New("IP-адрес обязателен для заполнения")
	}
	if parsedIP := net.ParseIP(r.IPAddress); parsedIP == nil {
		return errors.New("некорректный формат IP-адреса (должен быть валидный IPv4 или IPv6)")
	}

	r.Environment = strings.ToLower(strings.TrimSpace(r.Environment))
	switch r.Environment {
	case "production", "staging", "development":
	default:
		return errors.New("некорректное окружение: допустимы 'production', 'staging', 'development'")
	}

	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
	switch r.Status {
	case "active", "maintenance", "offline":
	default:
		return errors.New("некорректный статус: допустимы 'active', 'maintenance', 'offline'")
	}

	return nil
}

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}
