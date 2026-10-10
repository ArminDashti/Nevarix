package httpserver

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/config"
	"github.com/ArminDashti/nevarix-api/internal/dockerx"
	"github.com/ArminDashti/nevarix-api/internal/hubclient"
	"github.com/ArminDashti/nevarix-api/internal/metrics"
	"github.com/ArminDashti/nevarix-api/internal/softether"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg       config.Config
	metrics   *metrics.Collector
	docker    *dockerx.Client
	softether *softether.Client
	hub       *hubclient.Client
}

func New(cfg config.Config, m *metrics.Collector, d *dockerx.Client, se *softether.Client, hub *hubclient.Client) *Server {
	return &Server{cfg: cfg, metrics: m, docker: d, softether: se, hub: hub}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.GET("/agents", s.getAgents)
		v1.GET("/system", s.getSystem)
		v1.GET("/dashboard", s.getDashboard)
		v1.GET("/cpu", s.getCPU)
		v1.GET("/memory", s.getMemory)
		v1.GET("/disk", s.getDisk)
		v1.GET("/bandwidth", s.getBandwidth)
		v1.GET("/docker/images", s.getDockerImages)
		v1.GET("/docker/containers", s.getDockerContainers)
		v1.GET("/softether/sessions", s.getSoftEtherSessions)
		v1.GET("/softether/users", s.getSoftEtherUsers)
		v1.GET("/softether/users/:username/sessions", s.getSoftEtherUserSessions)
	}
	return r
}

// latestAgentOrNil returns the Hub's latest agent snapshot, or nil when Hub has none or is unreachable.
func (s *Server) latestAgentOrNil(c *gin.Context) *hubclient.AgentLatest {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	latest, err := s.hub.LatestAgent(ctx)
	if err != nil {
		return nil
	}
	return latest
}

func (s *Server) getDashboard(c *gin.Context) {
	if latest := s.latestAgentOrNil(c); latest != nil {
		diskUsed := uint64(0)
		diskTotal := uint64(0)
		for _, d := range latest.Disks {
			diskUsed += d.UsedBytes
			diskTotal += d.TotalBytes
		}
		c.JSON(http.StatusOK, gin.H{
			"cpuPercent":        latest.CPUPercent,
			"memoryUsedBytes":   latest.MemoryUsedBytes,
			"memoryTotalBytes":  latest.MemoryTotalBytes,
			"memoryUsedPercent": latest.MemoryUsedPercent,
			"diskUsedBytes":     diskUsed,
			"diskTotalBytes":    diskTotal,
			"bandwidthRxBps":    latest.BandwidthRxBps,
			"bandwidthTxBps":    latest.BandwidthTxBps,
			"uptimeSeconds":     latest.UptimeSeconds,
			"collectedAt":       latest.CollectedAt.UTC(),
			"agentId":           latest.AgentID,
			"hostname":          latest.Hostname,
		})
		return
	}

	snap := s.metrics.Snapshot()
	diskUsed := uint64(0)
	diskTotal := uint64(0)
	for _, d := range snap.Disks {
		diskUsed += d.UsedBytes
		diskTotal += d.TotalBytes
	}
	c.JSON(http.StatusOK, gin.H{
		"cpuPercent":        snap.CPUPercent,
		"memoryUsedBytes":   snap.MemoryUsedBytes,
		"memoryTotalBytes":  snap.MemoryTotalBytes,
		"memoryUsedPercent": snap.MemoryUsedPercent,
		"diskUsedBytes":     diskUsed,
		"diskTotalBytes":    diskTotal,
		"bandwidthRxBps":    snap.BandwidthRxBps,
		"bandwidthTxBps":    snap.BandwidthTxBps,
		"uptimeSeconds":     snap.UptimeSeconds,
		"collectedAt":       snap.CollectedAt.UTC(),
	})
}

// agentHistoryOrNil returns the Hub history for an agent, or nil when unavailable.
func (s *Server) agentHistoryOrNil(c *gin.Context, agentID string) []hubclient.MetricPoint {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	points, err := s.hub.AgentHistory(ctx, agentID)
	if err != nil {
		return nil
	}
	return points
}

func (s *Server) getCPU(c *gin.Context) {
	if latest := s.latestAgentOrNil(c); latest != nil {
		points := s.agentHistoryOrNil(c, latest.AgentID)
		series := make([]metrics.Point, 0, len(points))
		for _, p := range points {
			series = append(series, metrics.Point{Timestamp: p.Timestamp, Value: p.CPUPercent})
		}
		c.JSON(http.StatusOK, gin.H{
			"cpuPercent":  latest.CPUPercent,
			"cores":       latest.Cores,
			"series":      series,
			"collectedAt": latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"cpuPercent":  snap.CPUPercent,
		"cores":       snap.Cores,
		"series":      s.metrics.CPUHistory(),
		"collectedAt": snap.CollectedAt.UTC(),
	})
}

func (s *Server) getMemory(c *gin.Context) {
	if latest := s.latestAgentOrNil(c); latest != nil {
		points := s.agentHistoryOrNil(c, latest.AgentID)
		series := make([]metrics.Point, 0, len(points))
		for _, p := range points {
			series = append(series, metrics.Point{Timestamp: p.Timestamp, Value: p.MemPercent})
		}
		c.JSON(http.StatusOK, gin.H{
			"memoryTotalBytes":     latest.MemoryTotalBytes,
			"memoryUsedBytes":      latest.MemoryUsedBytes,
			"memoryAvailableBytes": latest.MemoryAvailableBytes,
			"memoryUsedPercent":    latest.MemoryUsedPercent,
			"series":               series,
			"collectedAt":          latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"memoryTotalBytes":     snap.MemoryTotalBytes,
		"memoryUsedBytes":      snap.MemoryUsedBytes,
		"memoryAvailableBytes": snap.MemoryAvailableBytes,
		"memoryUsedPercent":    snap.MemoryUsedPercent,
		"series":               s.metrics.MemoryHistory(),
		"collectedAt":          snap.CollectedAt.UTC(),
	})
}

func (s *Server) getDisk(c *gin.Context) {
	if latest := s.latestAgentOrNil(c); latest != nil {
		c.JSON(http.StatusOK, gin.H{
			"disks":       latest.Disks,
			"collectedAt": latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"disks":       snap.Disks,
		"collectedAt": snap.CollectedAt.UTC(),
	})
}

func (s *Server) getBandwidth(c *gin.Context) {
	if latest := s.latestAgentOrNil(c); latest != nil {
		points := s.agentHistoryOrNil(c, latest.AgentID)
		series := make([]metrics.BandwidthPoint, 0, len(points))
		for _, p := range points {
			series = append(series, metrics.BandwidthPoint{Timestamp: p.Timestamp, RxBps: p.RxBps, TxBps: p.TxBps})
		}
		c.JSON(http.StatusOK, gin.H{
			"bandwidthRxBps": latest.BandwidthRxBps,
			"bandwidthTxBps": latest.BandwidthTxBps,
			"series":         series,
			"collectedAt":    latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"bandwidthRxBps": snap.BandwidthRxBps,
		"bandwidthTxBps": snap.BandwidthTxBps,
		"series":         s.metrics.BandwidthHistory(),
		"collectedAt":    snap.CollectedAt.UTC(),
	})
}

func (s *Server) getDockerImages(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"images": []any{}, "error": "docker unavailable"})
		return
	}
	imgs, err := s.docker.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"images": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"images": imgs})
}

func (s *Server) getDockerContainers(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"stacks": []any{}, "error": "docker unavailable"})
		return
	}
	stacks, err := s.docker.ListContainersByStack(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"stacks": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stacks": stacks})
}

func (s *Server) getSoftEtherSessions(c *gin.Context) {
	sessions, err := s.softether.ListOnlineSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// hubJSON proxies a Hub read straight through; Hub owns the SoftEther history.
func (s *Server) hubJSON(c *gin.Context, path string, key string, fallback any) {
	raw, err := s.hub.GetJSON(c.Request.Context(), path, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{key: fallback, "error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}

func (s *Server) getSoftEtherUsers(c *gin.Context) {
	s.hubJSON(c, "/api/v1/softether/users", "users", []any{})
}

func (s *Server) getSoftEtherUserSessions(c *gin.Context) {
	s.hubJSON(c, "/api/v1/softether/users/"+url.PathEscape(c.Param("username"))+"/sessions", "sessions", []any{})
}
