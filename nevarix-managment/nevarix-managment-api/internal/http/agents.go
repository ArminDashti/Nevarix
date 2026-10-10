package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getSystem returns the most recently seen agent's system block (info, motherboard, firewall, temperatures).
func (s *Server) getSystem(c *gin.Context) {
	latest := s.latestAgentOrNil(c)
	if latest == nil {
		c.JSON(http.StatusOK, gin.H{"hostname": "", "system": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hostname": latest.Hostname, "system": latest.System})
}

// getAgents proxies the Hub's agent list.
func (s *Server) getAgents(c *gin.Context) {
	s.hubJSON(c, "/api/v1/agents", "agents", []any{})
}
