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
