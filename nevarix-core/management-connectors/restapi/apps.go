package restapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (p *Panel) getApps(c *gin.Context) {
	apps, err := p.store.ListApps(c.Request.Context(), c.Query("agentId"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

func (p *Panel) getEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	events, err := p.store.ListEvents(c.Request.Context(), c.Query("agentId"), limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (p *Panel) getActions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	actions, err := p.store.ListActions(c.Request.Context(), limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

// audit records a successful mutating dashboard operation.
func (p *Panel) audit(c *gin.Context, kind, detail string) {
	_ = p.store.RecordAction(c.Request.Context(), kind, detail)
}
