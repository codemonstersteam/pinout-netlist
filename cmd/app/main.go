// Command app — точка входа pinout-netlist.
//
// Wiring (ticket-05): /health — liveness; POST /reports — срез 01 (ingest);
// остальные маршруты — 501 NOT_IMPLEMENTED, пока их срезы не реализованы
// (компонентные тесты идут RED по бизнес-причине; см. скилл component-tests).
package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pinout-netlist/internal/ingest"
	"pinout-netlist/internal/shared/config"
	"pinout-netlist/internal/shared/store"
)

// errorBody — форма ошибок контракта (openapi.yaml Error).
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	var b errorBody
	b.Error.Code = code
	b.Error.Message = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(b)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1) // fail fast на невалидном конфиге
	}

	st := store.New(cfg.StoreFile)
	ingestDeps := ingest.NewDeps(st)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Срез 01: POST /reports — приём отчёта валидатора (канон 1.1).
	mux.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST only")
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
		if err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REPORT", err.Error())
			return
		}
		accepted, err := ingest.ProcessIngest(ingest.IngestRequest{Body: body}, ingestDeps)
		if err != nil {
			writeError(w, ingest.HTTPStatus(err), err.Error(), "report rejected")
			return
		}
		writeJSON(w, http.StatusAccepted, accepted)
	})

	// Placeholder: все прочие маршруты → 501, пока их срезы не реализованы.
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
	})

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		<-quit
		log.Info("shutting down")
		_ = srv.Close()
	}()

	log.Info("listening", "addr", cfg.ListenAddr, "store_file", cfg.StoreFile)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("listen", "err", err)
		os.Exit(1)
	}
}
