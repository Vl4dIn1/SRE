package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

type AlertItem struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

type WebhookPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []AlertItem       `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
}

type NotificationRecord struct {
	ReceivedAt  time.Time
	Status      string
	AlertName   string
	Severity    string
	Service     string
	Summary     string
	Description string
}

var (
	historyLock   sync.Mutex
	notifications []NotificationRecord
)

const htmlPage = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Центр уведомлений Alertmanager (StatusHub SRE)</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="5">
</head>
<body class="bg-slate-900 text-slate-100 min-h-screen p-6 font-sans">
    <div class="max-w-5xl mx-auto">
        <header class="flex items-center justify-between border-b border-slate-700 pb-5 mb-8">
            <div>
                <h1 class="text-3xl font-extrabold tracking-tight text-white flex items-center gap-3">
                    <span class="inline-block p-2 bg-rose-600 rounded-lg shadow-lg">🔔</span>
                    Центр уведомлений Alertmanager
                </h1>
                <p class="text-slate-400 text-sm mt-1">StatusHub Server Inventory Monitoring & Alerting</p>
            </div>
            <div class="flex items-center gap-3 text-xs">
                <span class="inline-flex items-center px-3 py-1 rounded-full font-medium bg-emerald-950 text-emerald-400 border border-emerald-800">
                    <span class="w-2 h-2 mr-2 bg-emerald-400 rounded-full animate-ping"></span>
                    Webhook Live: 5001
                </span>
                <span class="text-slate-400">Автообновление: 5с</span>
            </div>
        </header>

        <div class="space-y-4">
            {{if not .}}
            <div class="bg-slate-800 border border-slate-700 rounded-xl p-12 text-center text-slate-400">
                <p class="text-xl">📭 Уведомлений пока нет</p>
                <p class="text-sm text-slate-500 mt-2">Alertmanager отправит вебхук при переходе любого алерта в статус FIRING.</p>
            </div>
            {{else}}
            {{range .}}
            <div class="rounded-xl border p-5 shadow-lg transition-all {{if eq .Status "firing"}}bg-gradient-to-r from-rose-950/40 to-slate-800 border-rose-600/60{{else}}bg-gradient-to-r from-emerald-950/30 to-slate-800 border-emerald-600/60{{end}}">
                <div class="flex items-start justify-between">
                    <div class="flex items-center gap-3">
                        <span class="px-2.5 py-1 rounded text-xs font-bold uppercase tracking-wider {{if eq .Status "firing"}}bg-rose-600 text-white shadow-rose-900/50 shadow-md{{else}}bg-emerald-600 text-white{{end}}">
                            {{.Status}}
                        </span>
                        <h2 class="text-xl font-bold text-white tracking-tight">{{.AlertName}}</h2>
                        <span class="px-2 py-0.5 rounded text-xs font-semibold {{if eq .Severity "critical"}}bg-red-500/20 text-red-400 border border-red-500/30{{else}}bg-amber-500/20 text-amber-400 border border-amber-500/30{{end}}">
                            severity: {{.Severity}}
                        </span>
                        {{if .Service}}
                        <span class="px-2 py-0.5 rounded text-xs font-semibold bg-sky-500/20 text-sky-400 border border-sky-500/30">
                            service: {{.Service}}
                        </span>
                        {{end}}
                    </div>
                    <time class="text-xs font-mono text-slate-400">{{.ReceivedAt.Format "15:04:05 02.01.2006"}}</time>
                </div>
                <div class="mt-3 pl-1 border-l-2 {{if eq .Status "firing"}}border-rose-500{{else}}border-emerald-500{{end}}">
                    <p class="text-slate-200 font-semibold text-base">{{.Summary}}</p>
                    <p class="text-slate-400 text-sm mt-1 leading-relaxed">{{.Description}}</p>
                </div>
            </div>
            {{end}}
            {{end}}
        </div>
    </div>
</body>
</html>`

var tmpl = template.Must(template.New("index").Parse(htmlPage))

func main() {
	http.HandleFunc("POST /webhook", handleWebhook)
	http.HandleFunc("GET /", handleIndex)

	port := "5001"
	log.Printf("🚀 Webhook-приёмник Alertmanager запущен на :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	historyLock.Lock()
	defer historyLock.Unlock()

	for _, alert := range payload.Alerts {
		rec := NotificationRecord{
			ReceivedAt:  time.Now(),
			Status:      alert.Status,
			AlertName:   alert.Labels["alertname"],
			Severity:    alert.Labels["severity"],
			Service:     alert.Labels["service"],
			Summary:     alert.Annotations["summary"],
			Description: alert.Annotations["description"],
		}
		if rec.Summary == "" {
			rec.Summary = payload.CommonAnnotations["summary"]
		}
		if rec.Description == "" {
			rec.Description = payload.CommonAnnotations["description"]
		}

		notifications = append([]NotificationRecord{rec}, notifications...)
		if len(notifications) > 100 {
			notifications = notifications[:100]
		}

		fmt.Printf("\n═══════════════════════════════════════════════════════════════════\n")
		fmt.Printf("🔔 [ALERTMANAGER NOTIFICATION] %s\n", stringsToUpper(rec.Status))
		fmt.Printf("📌 Алерт:      %s\n", rec.AlertName)
		fmt.Printf("⚠️  Важность:   %s\n", rec.Severity)
		if rec.Service != "" {
			fmt.Printf("📦 Сервис:     %s\n", rec.Service)
		}
		fmt.Printf("📝 Заголовок:  %s\n", rec.Summary)
		fmt.Printf("📄 Описание:   %s\n", rec.Description)
		fmt.Printf("⏰ Время:      %s\n", rec.ReceivedAt.Format("15:04:05 02-01-2006"))
		fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	historyLock.Lock()
	data := make([]NotificationRecord, len(notifications))
	copy(data, notifications)
	historyLock.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

func stringsToUpper(s string) string {
	switch s {
	case "firing":
		return "🔥 FIRING"
	case "resolved":
		return "✅ RESOLVED"
	default:
		return s
	}
}
