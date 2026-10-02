package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

type App struct {
	db *sql.DB
}

func main() {
	initLogger()

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgrespassword@localhost:5432/server_inventory?sslmode=disable")

	slog.Info("Запуск сервиса Server Inventory", "port", port)

	initCtx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	db, err := InitDB(initCtx, dbURL)
	if err != nil {
		slog.Error("Критическая ошибка подключения к БД", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	app := &App{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/servers", app.handleGetServers)
	mux.HandleFunc("GET /api/v1/servers/{id}", app.handleGetServerByID)
	mux.HandleFunc("POST /api/v1/servers", app.handleCreateServer)
	mux.HandleFunc("PUT /api/v1/servers/{id}", app.handleUpdateServer)
	mux.HandleFunc("DELETE /api/v1/servers/{id}", app.handleDeleteServer)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("HTTP-сервер слушает входящие соединения", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Ошибка работы HTTP-сервера", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Получен сигнал завершения работы, начало Graceful Shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Принудительное завершение сервера из-за ошибки", "error", err)
		_ = server.Close()
	}
	slog.Info("HTTP-сервер успешно остановлен")
}

func (a *App) handleGetServers(w http.ResponseWriter, r *http.Request) {
	env := r.URL.Query().Get("environment")
	search := r.URL.Query().Get("search")

	servers, err := GetAllServers(r.Context(), a.db, env, search)
	if err != nil {
		slog.Error("Ошибка чтения серверов из БД", "error", err)
		sendError(w, http.StatusInternalServerError, "Не удалось получить список серверов")
		return
	}
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Data: servers})
}

func (a *App) handleGetServerByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Некорректный ID сервера")
		return
	}

	server, err := GetServerByID(r.Context(), a.db, id)
	if err != nil {
		slog.Error("Ошибка поиска сервера", "id", id, "error", err)
		sendError(w, http.StatusInternalServerError, "Ошибка базы данных")
		return
	}
	if server == nil {
		sendError(w, http.StatusNotFound, "Сервер с указанным ID не найден")
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{Success: true, Data: server})
}

func (a *App) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var input ServerInput
	if err := decodeJSON(w, r, &input); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	server, err := CreateServer(r.Context(), a.db, input)
	if err != nil {
		slog.Error("Ошибка сохранения сервера в БД", "error", err)
		sendError(w, http.StatusInternalServerError, "Не удалось создать сервер")
		return
	}

	slog.Info("Создан новый сервер", "id", server.ID, "name", server.Name)
	sendJSON(w, http.StatusCreated, APIResponse{Success: true, Data: server})
}

func (a *App) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Некорректный ID сервера")
		return
	}

	var input ServerInput
	if err := decodeJSON(w, r, &input); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	server, err := UpdateServer(r.Context(), a.db, id, input)
	if err != nil {
		slog.Error("Ошибка обновления сервера в БД", "id", id, "error", err)
		sendError(w, http.StatusInternalServerError, "Ошибка обновления данных")
		return
	}
	if server == nil {
		sendError(w, http.StatusNotFound, "Сервер с указанным ID не найден")
		return
	}

	slog.Info("Сервер обновлен", "id", server.ID, "name", server.Name)
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Data: server})
}

func (a *App) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Некорректный ID сервера")
		return
	}

	deleted, err := DeleteServer(r.Context(), a.db, id)
	if err != nil {
		slog.Error("Ошибка удаления сервера", "id", id, "error", err)
		sendError(w, http.StatusInternalServerError, "Ошибка удаления записи")
		return
	}
	if !deleted {
		sendError(w, http.StatusNotFound, "Сервер с указанным ID не найден")
		return
	}

	slog.Info("Сервер удален", "id", id)
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Data: id})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{ Validate() error }) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.New("некорректный JSON в теле запроса")
	}
	return v.Validate()
}

func parseID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, msg string) {
	sendJSON(w, status, APIResponse{Success: false, Error: msg})
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func initLogger() {
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, nil)
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
