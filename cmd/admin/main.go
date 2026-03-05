package main

import (
	"GoFlow/internal/gen/admin"
	"GoFlow/internal/service"
	"GoFlow/internal/storage"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("service", "admin-service"),
	)
	slog.SetDefault(logger)

	godotenv.Load()

	pgConn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s",
		os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_PORT"), os.Getenv("PG_DB"))

	pg, err := storage.NewPostgresStorage(ctx, pgConn)
	if err != nil {
		slog.Error("failed_to_connect_db", "error", err)
		os.Exit(1)
	}

	grpcPort := os.Getenv("ADMIN_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	httpPort := os.Getenv("ADMIN_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8082"
	}

	adminSrv := service.NewAdminServer(pg)
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			slog.Error("grpc_listen_failed", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		admin.RegisterAdminServiceServer(grpcServer, adminSrv)

		slog.Info("grpc_server_started", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("grpc_serve_failed", "error", err)
		}
	}()

	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()

		r.POST("/register", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Вызываем gRPC метод напрямую через структуру сервера
			resp, err := adminSrv.RegisterSite(c.Request.Context(), &admin.RegisterSiteRequest{
				Name: req.Name,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register site"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"site_id": resp.SiteId,
				"api_key": resp.ApiKey,
			})
		})

		slog.Info("http_admin_api_started", "port", httpPort)
		if err := r.Run(":" + httpPort); err != nil {
			slog.Error("http_serve_failed", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting_down_admin_service")
}
