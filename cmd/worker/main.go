package main

import (
	"GoFlow/internal/kafka"
	"GoFlow/internal/model"
	"GoFlow/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	config     *Config
	chStorage  *storage.ClickHouseStorage
	pgStorage  *storage.PostgresStorage
	consumer   *kafka.Consumer
	validSites map[string]bool
	sitesMu    sync.RWMutex
}

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	CHAddr       string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	app := &App{validSites: make(map[string]bool)}
	if err := app.initStorages(ctx); err != nil {
		log.Fatal("Init Storages Error:", err)
	}

	var wg sync.WaitGroup
	eventsChan := make(chan model.Event, 1000)

	app.runCacheUpdater(ctx, &wg)

	app.runClickHouseBatcher(ctx, &wg, eventsChan)

	app.startConsumption(ctx, eventsChan)

	fmt.Println("\nShutting down gracefully...")
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
		fmt.Println("Worker stopped cleanly.")
	case <-waitCtx.Done():
		fmt.Println("Shutdown timed out, forcing exit.")
	}
}

func (a *App) initStorages(ctx context.Context) error {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// ClickHouse
	var err error
	a.chStorage, err = storage.NewClickHouseStorage(
		fmt.Sprintf("localhost:%s", os.Getenv("CH_PORT_TCP")),
		os.Getenv("CH_USER"), os.Getenv("CH_PASSWORD"), os.Getenv("CH_DB"),
	)

	if err != nil {
		return err
	}

	// Postgres
	pgConn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s",
		os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_PORT"), os.Getenv("PG_DB"))
	a.pgStorage, err = storage.NewPostgresStorage(initCtx, pgConn)
	if err != nil {
		return err
	}

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
					log.Printf("Failed to update cache: %v", err)
				}
				cancel()
			}
		}
	}()
}

func (a *App) updateSitesCache(ctx context.Context) error {
	ids, err := a.pgStorage.GetValidSiteIDs(ctx)
	if err != nil {
		log.Printf("Cache update error: %v", err)
		return err
	}
	a.sitesMu.Lock()
	a.validSites = ids
	a.sitesMu.Unlock()
	log.Printf("Cache updated: %d sites", len(ids))
	return nil
}

func (a *App) runClickHouseBatcher(ctx context.Context, wg *sync.WaitGroup, ch <-chan model.Event) {
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
	fmt.Println("Worker started. Reading Kafka...")
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
				continue
			}

			a.sitesMu.RLock()
			isValid := a.validSites[event.SiteID]
			a.sitesMu.RUnlock()

			if isValid {
				eventsChan <- event
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

	if err := a.chStorage.InsertEvents(flushCtx, batch); err != nil {
		log.Printf("Flush error: %v", err)
	}
	log.Printf("Flushed %d events to ClickHouse", len(batch))
}

func (a *App) cleanup() {
	if err := a.consumer.Close(); err != nil {
		log.Printf("Kafka close error: %v", err)
	}
	a.pgStorage.Close()
	if err := a.chStorage.Close(); err != nil {
		log.Printf("CH close error: %v", err)
	}
}
