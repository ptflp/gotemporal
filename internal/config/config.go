package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr          string
	DatabaseURL       string
	TemporalHostPort  string
	TemporalNamespace string
	TaskQueue         string
	// Workflow step durations for demo purposes.
	PaymentDelay  time.Duration
	ShippingDelay time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:          getEnv("APP_HTTP_ADDR", ":3000"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://temporal:temporal@localhost:5432/app?sslmode=disable"),
		TemporalHostPort:  getEnv("TEMPORAL_HOSTPORT", "localhost:7233"),
		TemporalNamespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		TaskQueue:         getEnv("TASK_QUEUE", "orders"),
		PaymentDelay:      getDurationEnv("PAYMENT_DELAY_SECONDS", 3*time.Second),
		ShippingDelay:     getDurationEnv("SHIPPING_DELAY_SECONDS", 3*time.Second),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		sec, err := strconv.Atoi(v)
		if err == nil {
			return time.Duration(sec) * time.Second
		}
	}
	return def
}
