package restapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SoftEther and Docker are local integrations: the panel reads them live, and SoftEther history comes from the store.

func (p *Panel) getDockerImages(c *gin.Context) {
	if p.docker == nil {
		c.JSON(http.StatusOK, gin.H{"images": []any{}, "error": "docker unavailable"})
		return
	}
	imgs, err := p.docker.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"images": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"images": imgs})
}

func (p *Panel) getDockerContainers(c *gin.Context) {
	if p.docker == nil {
		c.JSON(http.StatusOK, gin.H{"stacks": []any{}, "error": "docker unavailable"})
		return
	}
	stacks, err := p.docker.ListContainersByStack(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"stacks": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stacks": stacks})
}

func (p *Panel) getSoftEtherSessions(c *gin.Context) {
	sessions, err := p.softether.ListOnlineSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (p *Panel) getSoftEtherUsers(c *gin.Context) {
	users, err := p.store.ListUserStats(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (p *Panel) getSoftEtherUserSessions(c *gin.Context) {
	username := c.Param("username")
	sessions, err := p.store.ListUserSessions(c.Request.Context(), username)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": username, "sessions": sessions})
}
