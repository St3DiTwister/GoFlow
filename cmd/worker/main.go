package main

import (
	"GoFlow/internal/gen/admin"
	"GoFlow/internal/kafka"
	"GoFlow/internal/metrics"
	"GoFlow/internal/model"
	"GoFlow/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	config      *Config
	chStorage   *storage.ClickHouseStorage
	adminClient admin.AdminServiceClient
	grpcConn    *grpc.ClientConn
	consumer    *kafka.Consumer
	validSites  map[string]bool
	sitesMu     sync.RWMutex
}

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	CHAddr       string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("service", "goflow-worker"),
		slog.String("env", "development"),
	)
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn("env_file_not_found", "using", "system_env")
	}

	metricsPort := os.Getenv("WORKER_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9001"
	}

	adminAddr := os.Getenv("ADMIN_GRPC_ADDR")
	if adminAddr == "" {
		adminAddr = "localhost:50051"
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		slog.Info("metrics_server_started", "port", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			slog.Error("metrics_server_failed", "error", err)
		}
	}()

	app := &App{validSites: make(map[string]bool)}
	if err := app.initStorages(ctx, adminAddr); err != nil {
		slog.Error("storage_init_failed", "error", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	eventsChan := make(chan model.Event, 1000)

	app.runCacheUpdater(ctx, &wg)
	app.runClickHouseBatcher(ctx, &wg, eventsChan)

	slog.Info("worker_started", "topic", os.Getenv("KAFKA_TOPIC"))

	app.startConsumption(ctx, eventsChan)

	slog.Info("shutdown_initiated")
	close(eventsChan)
	app.cleanup()

	waitCtx, timeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeout()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("worker_stopped_cleanly")
	case <-waitCtx.Done():
		slog.Error("shutdown_timeout", "msg", "some_data_might_be_lost")
	}
}

func (a *App) initStorages(ctx context.Context, adminAddr string) error {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// ClickHouse
	var err error
	a.chStorage, err = storage.NewClickHouseStorage(
		fmt.Sprintf("localhost:%s", os.Getenv("CH_PORT_TCP")),
		os.Getenv("CH_USER"), os.Getenv("CH_PASSWORD"), os.Getenv("CH_DB"),
	)

	if err != nil {
		return fmt.Errorf("clickhouse: %w", err)
	}

	conn, err := grpc.NewClient(adminAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("did not connect to admin service: %v", err)
	}
	a.adminClient = admin.NewAdminServiceClient(conn)
	a.grpcConn = conn

	// Kafka
	a.consumer = kafka.NewConsumer([]string{os.Getenv("KAFKA_BROKER")}, os.Getenv("KAFKA_TOPIC"), "event-processor-v1")

	return a.updateSitesCache(initCtx)
}

func (a *App) runCacheUpdater(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updatesCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err := a.updateSitesCache(updatesCtx)
				if err != nil {
					slog.Error("cache_update_failed", "error", err)
				}
				cancel()
			}
		}
	}()
}

func (a *App) updateSitesCache(ctx context.Context) error {
	resp, err := a.adminClient.GetValidSites(ctx, &admin.GetValidSitesRequest{})
	if err != nil {
		return fmt.Errorf("grpc_get_valid_sites: %v", err)
	}

	newSites := make(map[string]bool)
	for _, id := range resp.SiteIds {
		newSites[id] = true
	}

	a.sitesMu.Lock()
	a.validSites = newSites
	a.sitesMu.Unlock()

	slog.Info("cache_updated_via_grpc", "count", len(newSites))
	return nil
}

func (a *App) runClickHouseBatcher(_ context.Context, wg *sync.WaitGroup, ch <-chan model.Event) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		batch := make([]model.Event, 0, 1000)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					a.finalFlush(batch)
					return
				}
				batch = append(batch, event)
				if len(batch) >= 1000 {
					a.finalFlush(batch)
					batch = batch[:0]
					ticker.Reset(5 * time.Second)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					a.finalFlush(batch)
					batch = batch[:0]
				}
			}
		}
	}()
}

func (a *App) startConsumption(ctx context.Context, eventsChan chan<- model.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			msg, err := a.consumer.ReadMessage(readCtx)
			cancel()

			if err != nil {
				continue
			}

			var event model.Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("json_unmarshal_error", "error", err, "payload", string(msg.Value))
				continue
			}

			a.sitesMu.RLock()
			isValid := a.validSites[event.SiteID]
			a.sitesMu.RUnlock()

			if isValid {
				metrics.ProcessedEvents.Inc()
				eventsChan <- event
			} else {
				metrics.InvalidEvents.Inc()
				slog.Debug("invalid_site_id", "site_id", event.SiteID)
			}
		}
	}
}

func (a *App) finalFlush(batch []model.Event) {
	if len(batch) == 0 {
		return
	}

	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	if err := a.chStorage.InsertEvents(flushCtx, batch); err != nil {
		slog.Error("clickhouse_insert_failed", "error", err, "batch_size", len(batch))
	} else {
		metrics.ClickHouseInsertDuration.Observe(time.Since(start).Seconds())
		slog.Info("batch_flushed", "count", len(batch))
	}
}

func (a *App) cleanup() {
	if err := a.consumer.Close(); err != nil {
		slog.Error("kafka_close_failed", "error", err)
	}
	if err := a.chStorage.Close(); err != nil {
		slog.Error("clickhouse_close_failed", "error", err)
	}
	if err := a.grpcConn.Close(); err != nil {
		slog.Error("grpc_close_failed", "error", err)
	}
}
