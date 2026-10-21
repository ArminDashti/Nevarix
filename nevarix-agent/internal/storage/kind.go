package storage

import "strings"

func KindFromRotational(v string) string {
	switch strings.TrimSpace(v) {
	case "0":
		return "ssd"
	case "1":
		return "hdd"
	default:
		return "unknown"
	}
}

func KindFromMedia(media string) string {
	m := strings.ToLower(strings.TrimSpace(media))
	switch m {
	case "ssd", "4", "solid state drive":
		return "ssd"
	case "hdd", "3", "smr":
		return "hdd"
	}
	if strings.Contains(m, "ssd") || strings.Contains(m, "nvme") || strings.Contains(m, "solid") {
		return "ssd"
	}
	if strings.Contains(m, "hdd") || strings.Contains(m, "hard") {
		return "hdd"
	}
	return "unknown"
}

func SkipDevice(name string) bool {
	n := strings.ToLower(name)
	for _, p := range []string{"loop", "ram", "sr", "fd", "zram", "md", "dm-"} {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func rate(cur, prev uint64, sec float64) float64 {
	if sec <= 0 || cur < prev {
		return 0
	}
	return float64(int((float64(cur-prev)/sec)*100+0.5)) / 100
}
