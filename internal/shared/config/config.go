// Package config загружает конфигурацию сервиса из файла (config.yaml) с
// override из env. В коде НЕТ хардкодов — всё приходит отсюда. Fail fast на
// невалидном конфиге (см. NFR задач). Добавляй поля по мере роста сервиса.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config — конфигурация сервиса.
type Config struct {
	ListenAddr string `yaml:"listen_addr"`
	// StoreFile — путь JSON-снапшота хранилища (пусто = только in-memory).
	StoreFile string `yaml:"store_file"`
}

// Load читает config.yaml (путь из CONFIG_FILE, дефолт "config.yaml") и
// применяет env-override (PORT / LISTEN_ADDR). Отсутствие файла — не ошибка
// (используются дефолты + env). Невалидный YAML → ошибка (fail fast).
func Load() (Config, error) {
	cfg := Config{ListenAddr: ":8080"} // дефолт

	path := getenv("CONFIG_FILE", "config.yaml")
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	if p := os.Getenv("PORT"); p != "" {
		cfg.ListenAddr = ":" + p
	}
	if a := os.Getenv("LISTEN_ADDR"); a != "" {
		cfg.ListenAddr = a
	}
	if f := os.Getenv("STORE_FILE"); f != "" {
		cfg.StoreFile = f
	}
	if cfg.ListenAddr == "" {
		return Config{}, fmt.Errorf("listen_addr пуст")
	}
	return cfg, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
