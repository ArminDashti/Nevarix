package restapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
