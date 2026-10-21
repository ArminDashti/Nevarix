package restapi

import (
	"net/http"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/management-connectors/service"
	"github.com/gin-gonic/gin"
)

func (p *Panel) getEndpoints(c *gin.Context) {
	endpoints, err := p.svc.ListEndpoints(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

type addEndpointBody struct {
	IP          string `json:"ip"`
	Hostname    string `json:"hostname"`
	AgentPort   int    `json:"agentPort"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	VerifyAgent bool   `json:"verifyAgent"`
}

func (p *Panel) postEndpoint(c *gin.Context) {
	limitJSONBody(c)
	var body addEndpointBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := p.svc.AddEndpoint(c.Request.Context(), service.AddEndpointInput{
		IP: body.IP, Hostname: body.Hostname, AgentPort: body.AgentPort,
		Username: body.Username, Password: body.Password, VerifyAgent: body.VerifyAgent,
	}); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

type scanBody struct {
	CIDR      string `json:"cidr"`
	Port      int    `json:"port"`
	TimeoutMs int    `json:"timeoutMs"`
}

func (p *Panel) postScan(c *gin.Context) {
	limitJSONBody(c)
	var body scanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := p.svc.Scan(c.Request.Context(), service.ScanRequest{CIDR: body.CIDR, Port: body.Port, TimeoutMs: body.TimeoutMs})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"scanned": res.Scanned, "hits": res.Hits})
}

func (p *Panel) getSettings(c *gin.Context) {
	st, err := p.svc.GetSettings(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

func (p *Panel) putSettings(c *gin.Context) {
	limitJSONBody(c)
	var st model.Settings
	if err := c.ShouldBindJSON(&st); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved, err := p.svc.UpdateSettings(c.Request.Context(), st)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, saved)
}
