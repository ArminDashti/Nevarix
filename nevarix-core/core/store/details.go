package store

import "encoding/json"

type metricDetails struct {
	NICs   []nicPoint    `json:"nics"`
	Disks  []diskPoint   `json:"disks"`
	Docker []dockerPoint `json:"docker"`
}

type nicPoint struct {
	Name         string   `json:"name"`
	RxBps        float64  `json:"rxBps"`
	TxBps        float64  `json:"txBps"`
	TemperatureC *float64 `json:"temperatureC,omitempty"`
}

type diskPoint struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	ReadBps      float64  `json:"readBps"`
	WriteBps     float64  `json:"writeBps"`
	TemperatureC *float64 `json:"temperatureC,omitempty"`
}

type dockerPoint struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemBytes    uint64  `json:"memBytes"`
	RxBps       float64 `json:"rxBps"`
	TxBps       float64 `json:"txBps"`
	ReadBps     float64 `json:"readBps"`
	WriteBps    float64 `json:"writeBps"`
}

func MetricDetailsJSON(system json.RawMessage) []byte {
	if len(system) == 0 {
		return []byte("{}")
	}
	var r struct {
		Storage struct {
			SSDs []diskPoint `json:"ssds"`
			HDDs []diskPoint `json:"hdds"`
		} `json:"storage"`
		Network struct {
			NICs []nicPoint `json:"nics"`
		} `json:"network"`
		Docker struct {
			Containers []dockerPoint `json:"containers"`
		} `json:"docker"`
	}
	if err := json.Unmarshal(system, &r); err != nil {
		return []byte("{}")
	}
	d := metricDetails{
		NICs:   r.Network.NICs,
		Disks:  append(append([]diskPoint{}, r.Storage.SSDs...), r.Storage.HDDs...),
		Docker: r.Docker.Containers,
	}
	if d.NICs == nil {
		d.NICs = []nicPoint{}
	}
	if d.Disks == nil {
		d.Disks = []diskPoint{}
	}
	if d.Docker == nil {
		d.Docker = []dockerPoint{}
	}
	b, err := json.Marshal(d)
	if err != nil {
		return []byte("{}")
	}
	return b
}
