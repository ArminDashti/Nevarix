package restapi

import (
	"net/http"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/gin-gonic/gin"
)

// latestAgent returns the most recently reported agent snapshot, or nil when none exists or the read fails.
func (p *Panel) latestAgent(c *gin.Context) *model.AgentSnapshot {
	latest, err := p.store.LatestAgent(c.Request.Context())
	if err != nil {
		return nil
	}
	return latest
}

// agentHistory returns one agent's history, or nil when the read fails.
func (p *Panel) agentHistory(c *gin.Context, agentID string) []model.MetricPoint {
	points, err := p.store.AgentHistory(c.Request.Context(), agentID)
	if err != nil {
		return nil
	}
	return points
}

func (p *Panel) getAgents(c *gin.Context) {
	agents, err := p.store.ListAgents(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// getLatestAgent returns {"agent": null} when no agent has reported yet.
func (p *Panel) getLatestAgent(c *gin.Context) {
	latest, err := p.store.LatestAgent(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"agent": latest})
}

func (p *Panel) getAgentHistory(c *gin.Context) {
	points, err := p.store.AgentHistory(c.Request.Context(), c.Param("agentId"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (p *Panel) getSystem(c *gin.Context) {
	latest := p.latestAgent(c)
	if latest == nil {
		c.JSON(http.StatusOK, gin.H{"hostname": "", "system": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hostname": latest.Hostname, "system": latest.System})
}

func (p *Panel) getDashboard(c *gin.Context) {
	if latest := p.latestAgent(c); latest != nil {
		diskUsed, diskTotal := sumDisks(latest.Disks)
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

	snap := p.metrics.Snapshot()
	diskUsed, diskTotal := sumDisks(snap.Disks)
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

func (p *Panel) getCPU(c *gin.Context) {
	if latest := p.latestAgent(c); latest != nil {
		points := p.agentHistory(c, latest.AgentID)
		series := make([]model.Point, 0, len(points))
		for _, pt := range points {
			series = append(series, model.Point{Timestamp: pt.Timestamp, Value: pt.CPUPercent})
		}
		c.JSON(http.StatusOK, gin.H{
			"cpuPercent":  latest.CPUPercent,
			"cores":       latest.Cores,
			"series":      series,
			"collectedAt": latest.CollectedAt.UTC(),
		})
		return
	}
	snap := p.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"cpuPercent":  snap.CPUPercent,
		"cores":       snap.Cores,
		"series":      p.metrics.CPUHistory(),
		"collectedAt": snap.CollectedAt.UTC(),
	})
}

func (p *Panel) getMemory(c *gin.Context) {
	if latest := p.latestAgent(c); latest != nil {
		points := p.agentHistory(c, latest.AgentID)
		series := make([]model.Point, 0, len(points))
		for _, pt := range points {
			series = append(series, model.Point{Timestamp: pt.Timestamp, Value: pt.MemPercent})
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
	snap := p.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"memoryTotalBytes":     snap.MemoryTotalBytes,
		"memoryUsedBytes":      snap.MemoryUsedBytes,
		"memoryAvailableBytes": snap.MemoryAvailableBytes,
		"memoryUsedPercent":    snap.MemoryUsedPercent,
		"series":               p.metrics.MemoryHistory(),
		"collectedAt":          snap.CollectedAt.UTC(),
	})
}

func (p *Panel) getDisk(c *gin.Context) {
	if latest := p.latestAgent(c); latest != nil {
		c.JSON(http.StatusOK, gin.H{
			"disks":       latest.Disks,
			"collectedAt": latest.CollectedAt.UTC(),
		})
		return
	}
	snap := p.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"disks":       snap.Disks,
		"collectedAt": snap.CollectedAt.UTC(),
	})
}

func (p *Panel) getBandwidth(c *gin.Context) {
	if latest := p.latestAgent(c); latest != nil {
		points := p.agentHistory(c, latest.AgentID)
		series := make([]model.BandwidthPoint, 0, len(points))
		for _, pt := range points {
			series = append(series, model.BandwidthPoint{Timestamp: pt.Timestamp, RxBps: pt.RxBps, TxBps: pt.TxBps})
		}
		c.JSON(http.StatusOK, gin.H{
			"bandwidthRxBps": latest.BandwidthRxBps,
			"bandwidthTxBps": latest.BandwidthTxBps,
			"series":         series,
			"collectedAt":    latest.CollectedAt.UTC(),
		})
		return
	}
	snap := p.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"bandwidthRxBps": snap.BandwidthRxBps,
		"bandwidthTxBps": snap.BandwidthTxBps,
		"series":         p.metrics.BandwidthHistory(),
		"collectedAt":    snap.CollectedAt.UTC(),
	})
}

func sumDisks(disks []model.DiskSample) (used, total uint64) {
	for _, d := range disks {
		used += d.UsedBytes
		total += d.TotalBytes
	}
	return used, total
}
