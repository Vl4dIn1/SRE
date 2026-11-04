package main

import (
	"testing"
)

func TestServerInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ServerInput
		wantErr bool
	}{
		{
			name: "Успешная валидация корректного запроса",
			input: ServerInput{
				Name:        "web-prod-01",
				IPAddress:   "192.168.1.100",
				Environment: "production",
				Status:      "active",
			},
			wantErr: false,
		},
		{
			name: "Ошибка при пустом имени сервера",
			input: ServerInput{
				Name:        "   ",
				IPAddress:   "192.168.1.100",
				Environment: "production",
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "Ошибка при некорректном IP-адресе",
			input: ServerInput{
				Name:        "db-node",
				IPAddress:   "999.999.999.999",
				Environment: "staging",
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "Ошибка при неизвестном окружении",
			input: ServerInput{
				Name:        "test-node",
				IPAddress:   "10.0.0.1",
				Environment: "invalid_env",
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "Ошибка при неизвестном статусе",
			input: ServerInput{
				Name:        "test-node",
				IPAddress:   "10.0.0.1",
				Environment: "development",
				Status:      "sleeping",
			},
			wantErr: true,
		},
		{
			name: "Успешная валидация с IPv6 адресом",
			input: ServerInput{
				Name:        "ipv6-node",
				IPAddress:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				Environment: "development",
				Status:      "maintenance",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDatabaseURL(t *testing.T) {
	t.Run("Чтение из DATABASE_URL если задан", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://custom:pass@custom-host:5432/custom_db?sslmode=require")
		url := getDatabaseURL()
		if url != "postgres://custom:pass@custom-host:5432/custom_db?sslmode=require" {
			t.Errorf("Ожидался custom url, получено %s", url)
		}
	})

	t.Run("Формирование из раздельных переменных DB_*", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("DB_USER", "myuser")
		t.Setenv("DB_PASSWORD", "mypass")
		t.Setenv("DB_HOST", "myhost")
		t.Setenv("DB_PORT", "5433")
		t.Setenv("DB_NAME", "mydb")
		t.Setenv("DB_SSLMODE", "disable")

		url := getDatabaseURL()
		expected := "postgres://myuser:mypass@myhost:5433/mydb?sslmode=disable"
		if url != expected {
			t.Errorf("Ожидался %s, получено %s", expected, url)
		}
	})
}

