package restapi

import (
	"net/http"

	"github.com/ArminDashti/nevarix-core/management-connectors/service"
	"github.com/gin-gonic/gin"
)

func (p *Panel) getDevices(c *gin.Context) {
	devices, err := p.svc.ListNetDevices(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

func (p *Panel) getDevice(c *gin.Context) {
	d, err := p.svc.GetNetDevice(c.Request.Context(), c.Param("deviceId"))
	if err != nil {
		respondError(c, err)
		return
	}
	if d == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"device": d})
}

func (p *Panel) postDevice(c *gin.Context) {
	limitJSONBody(c)
	var body service.CreateNetDeviceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := p.svc.CreateNetDevice(c.Request.Context(), body)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (p *Panel) putDevice(c *gin.Context) {
	limitJSONBody(c)
	var body service.UpdateNetDeviceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := p.svc.UpdateNetDevice(c.Request.Context(), c.Param("deviceId"), body); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (p *Panel) deleteDevice(c *gin.Context) {
	if err := p.svc.DeleteNetDevice(c.Request.Context(), c.Param("deviceId")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (p *Panel) getDeviceInterfaces(c *gin.Context) {
	ifaces, err := p.svc.ListNetInterfaces(c.Request.Context(), c.Param("deviceId"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"interfaces": ifaces})
}

func (p *Panel) getDeviceHistory(c *gin.Context) {
	points, err := p.svc.NetDeviceHistory(c.Request.Context(), c.Param("deviceId"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

type snmpScanBody struct {
	CIDR      string `json:"cidr"`
	TimeoutMs int    `json:"timeoutMs"`
}

func (p *Panel) postDeviceScan(c *gin.Context) {
	limitJSONBody(c)
	var body snmpScanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scanned, hits, err := p.svc.ScanSNMP(c.Request.Context(), service.SnmpScanRequest{
		CIDR: body.CIDR, TimeoutMs: body.TimeoutMs,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"scanned": scanned, "hits": hits})
}
