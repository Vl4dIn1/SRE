package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)
var targetURL = "http://localhost:8080/api/v1/servers"

func main() {
	client := &http.Client{Timeout: 3 * time.Second}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("🚀 Пробер запущен! Шлём запросы к", targetURL)

	step := 0

	for range ticker.C {
		step++

		if step%5 == 0{
			proberPOST(client)
		} else {
			proberGET(client)
		}
	}
}

func proberGET(client *http.Client){
	start := time.Now()

	resp, err := client.Get(targetURL)
	duration := time.Since(start).Round(time.Millisecond)

	if err != nil {
		fmt.Printf("[%s] ❌ Сетевая ошибка: %v (%v)\n", time.Now().Format("15:04:05"), err, duration)
        return
	}

	defer resp.Body.Close()
	
	fmt.Printf("[%s] 🌐 GET %s -> Status: %d (%v)\n", 
        time.Now().Format("15:04:05"), targetURL, resp.StatusCode, duration)
}

func proberPOST(client *http.Client) {
	jsonData := []byte(`{
	"name":"probe-serv",
	"ip_address":"10.0.0.99",
	"environment":"production",
	"status":"active"}`,
	)

	start := time.Now()

	resp,err := client.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
	duration := time.Since(start).Round(time.Millisecond)

	if err != nil {
		fmt.Printf("[%s] ❌ POST Ошибка: %v (%v)\n", time.Now().Format("15:04:05"), err, duration)
		return
	}

	defer resp.Body.Close()
	fmt.Printf("[%s] 💾 POST %s -> Status: %d (%v)\n", 
		time.Now().Format("15:04:05"), targetURL, resp.StatusCode, duration)
}