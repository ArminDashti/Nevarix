// Command server runs nevarix-core: REST + gRPC for agents and the management panel.
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	agentauth "github.com/ArminDashti/nevarix-core/agent-connectors/auth"
	agentgrpc "github.com/ArminDashti/nevarix-core/agent-connectors/grpcapi"
	agentv1 "github.com/ArminDashti/nevarix-core/agent-connectors/pb/nevarix/agent/v1"
	agentrest "github.com/ArminDashti/nevarix-core/agent-connectors/restapi"
	"github.com/ArminDashti/nevarix-core/core/asn"
	"github.com/ArminDashti/nevarix-core/core/config"
	"github.com/ArminDashti/nevarix-core/core/dockerx"
	"github.com/ArminDashti/nevarix-core/core/metrics"
	"github.com/ArminDashti/nevarix-core/core/softether"
	"github.com/ArminDashti/nevarix-core/core/store"
	mgmtgrpc "github.com/ArminDashti/nevarix-core/management-connectors/grpcapi"
	managementv1 "github.com/ArminDashti/nevarix-core/management-connectors/pb/nevarix/management/v1"
	mgmtrest "github.com/ArminDashti/nevarix-core/management-connectors/restapi"
	"github.com/ArminDashti/nevarix-core/management-connectors/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.AgentIngestToken == "" {
		log.Fatal("AGENT_INGEST_TOKEN is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	st, err := store.Open(dbCtx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer st.Close()

	collector := metrics.NewCollector(cfg.MetricsHistoryPoints)
	dockerClient, err := dockerx.New()
	if err != nil {
		log.Printf("docker client unavailable: %v", err)
	}
	seClient := softether.New(cfg.SoftEtherContainer, cfg.SoftEtherPassword, cfg.SoftEtherHub, cfg.SoftEtherEnabled, asn.NewResolver())
	if cfg.SoftEtherEnabled {
		go softether.Sync(ctx, cfg.SoftEtherPollEvery, seClient, st)
	}

	svc := service.New(st)

	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger(), cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	engine.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	v1 := engine.Group("/api/v1")
	agentrest.New(st, cfg.AgentIngestToken, cfg.MetricsHistoryPoints).Register(v1)
	mgmtrest.New(svc, st, collector, dockerClient, seClient).Register(v1)

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(agentauth.UnaryInterceptor(cfg.AgentIngestToken)))
	agentv1.RegisterAgentServiceServer(grpcServer, agentgrpc.New(st, cfg.MetricsHistoryPoints))
	managementv1.RegisterManagementServiceServer(grpcServer, mgmtgrpc.New(svc, st))
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("nevarix-core REST on http://%s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()
	go func() {
		log.Printf("nevarix-core gRPC on %s", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
}
