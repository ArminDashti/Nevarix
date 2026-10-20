// Package restapi serves the management panel API over REST.
package restapi

import (
	"errors"
	"net/http"

	"github.com/ArminDashti/nevarix-core/core/dockerx"
	"github.com/ArminDashti/nevarix-core/core/metrics"
	"github.com/ArminDashti/nevarix-core/core/store"
	"github.com/ArminDashti/nevarix-core/management-connectors/service"
	"github.com/gin-gonic/gin"
)

const maxJSONBodyBytes = 1 << 16

type Panel struct {
	svc     *service.Management
	store   *store.Store
	metrics *metrics.Collector
	docker  *dockerx.Client
}

func New(svc *service.Management, st *store.Store, m *metrics.Collector, d *dockerx.Client) *Panel {
	return &Panel{svc: svc, store: st, metrics: m, docker: d}
}

// Register mounts the panel routes on the given group (e.g. /api/v1).
func (p *Panel) Register(v1 *gin.RouterGroup) {
	v1.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1.GET("/agents", p.getAgents)
	v1.GET("/agents/latest", p.getLatestAgent)
	v1.GET("/agents/:agentId/history", p.getAgentHistory)
	v1.GET("/system", p.getSystem)
	v1.GET("/dashboard", p.getDashboard)
	v1.GET("/cpu", p.getCPU)
	v1.GET("/memory", p.getMemory)
	v1.GET("/disk", p.getDisk)
	v1.GET("/bandwidth", p.getBandwidth)

	v1.GET("/docker/images", p.getDockerImages)
	v1.GET("/docker/containers", p.getDockerContainers)

	v1.GET("/endpoints", p.getEndpoints)
	v1.POST("/endpoints", p.postEndpoint)
	v1.POST("/scan", p.postScan)
	v1.GET("/settings", p.getSettings)
	v1.PUT("/settings", p.putSettings)
}

// respondError maps validation failures to 400 and everything else to 500.
func respondError(c *gin.Context, err error) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func limitJSONBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBodyBytes)
}
