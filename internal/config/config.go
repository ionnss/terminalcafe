package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Email
	EmailFrom     string
	EmailPassword string
	EmailTo       string
	SMTPHost      string
	SMTPPort      string

	// MercadoPago
	MPAccessToken string

	// Correios
	CorreiosCode     string
	CorreiosPassword string
	StoreCEP         string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("erro ao carregar .env: %v", err)
	}

	config := &Config{
		EmailFrom:        os.Getenv("CAFE_EMAIL"),
		EmailPassword:    os.Getenv("CAFE_EMAIL_PASSWORD"),
		EmailTo:          os.Getenv("CAFE_NOTIFICATION_EMAIL"),
		SMTPHost:         os.Getenv("CAFE_SMTP_HOST"),
		SMTPPort:         os.Getenv("CAFE_SMTP_PORT"),
		MPAccessToken:    os.Getenv("MP_ACCESS_TOKEN"),
		CorreiosCode:     os.Getenv("CORREIOS_CODE"),
		CorreiosPassword: os.Getenv("CORREIOS_PASSWORD"),
		StoreCEP:         os.Getenv("STORE_CEP"),
	}

	// Validação básica
	if config.EmailFrom == "" || config.EmailPassword == "" || config.EmailTo == "" {
		return nil, fmt.Errorf("configurações de email não definidas")
	}

	if config.SMTPHost == "" || config.SMTPPort == "" {
		return nil, fmt.Errorf("configurações de SMTP não definidas")
	}

	if config.MPAccessToken == "" {
		return nil, fmt.Errorf("token do MercadoPago não definido")
	}

	return config, nil
}
