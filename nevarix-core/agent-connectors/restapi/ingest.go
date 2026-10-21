// Package restapi serves the agent write path over REST.
package restapi

import (
	"net/http"

	"github.com/ArminDashti/nevarix-core/agent-connectors/auth"
	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/core/store"
	"github.com/gin-gonic/gin"
)

const maxBodyBytes = 4 << 20

type Ingest struct {
	store         *store.Store
	token         string
	historyPoints int
}

func New(st *store.Store, token string, historyPoints int) *Ingest {
	return &Ingest{store: st, token: token, historyPoints: historyPoints}
}

// Register mounts the agent routes on the given group (e.g. /api/v1).
func (i *Ingest) Register(v1 *gin.RouterGroup) {
	v1.POST("/ingest/metrics", i.requireToken, i.postIngestMetrics)
}

func (i *Ingest) requireToken(c *gin.Context) {
	if !auth.BearerMatches(c.GetHeader("Authorization"), i.token) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
}

// postIngestMetrics answers 200 only after the snapshot is committed, so the agent can delete its outbox row on any 2xx.
func (i *Ingest) postIngestMetrics(c *gin.Context) {
	var snap model.AgentSnapshot
	if err := c.ShouldBindJSON(&snap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if snap.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}
	if err := i.store.IngestAgent(c.Request.Context(), snap, i.historyPoints); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
