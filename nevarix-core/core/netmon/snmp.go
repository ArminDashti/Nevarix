package netmon

import (
	"fmt"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/gosnmp/gosnmp"
)

const (
	oidSysDescr   = "1.3.6.1.2.1.1.1.0"
	oidSysUpTime  = "1.3.6.1.2.1.1.3.0"
	oidSysObject  = "1.3.6.1.2.1.1.2.0"
	oidHrProcessor = "1.3.6.1.2.1.25.3.3.1.2"
	oidMikrotikCPU = "1.3.6.1.4.1.14988.1.1.3.14.0"
	oidCiscoCPU5m  = "1.3.6.1.4.1.9.9.109.1.1.1.1.5"

	oidIfDescr     = "1.3.6.1.2.1.2.2.1.2"
	oidIfOper      = "1.3.6.1.2.1.2.2.1.8"
	oidIfSpeed     = "1.3.6.1.2.1.2.2.1.5"
	oidIfInOctets  = "1.3.6.1.2.1.2.2.1.10"
	oidIfOutOctets = "1.3.6.1.2.1.2.2.1.16"
)

// ReachableSNMP dials the host and reads sysDescr to confirm SNMP v2c works.
func ReachableSNMP(target, community string) (string, error) {
	client := newClient(target, community)
	if err := client.Connect(); err != nil {
		return "", err
	}
	defer client.Conn.Close()
	descr, _, _, err := fetchSystem(client)
	return descr, err
}

func newClient(target, community string) *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   1,
	}
}

type priorOctetsFn func(deviceID string, ifIndex int) (uint64, uint64, time.Time, bool)

func PollDevice(device model.NetDevice, prior priorOctetsFn) (model.NetPollSample, error) {
	client := newClient(device.MgmtIP, device.SnmpCommunity)
	if err := client.Connect(); err != nil {
		return model.NetPollSample{}, err
	}
	defer client.Conn.Close()

	now := time.Now().UTC()
	sysDescr, sysObject, uptime, err := fetchSystem(client)
	if err != nil {
		return model.NetPollSample{}, err
	}
	vendor := IdentifyVendor(sysDescr, sysObject)
	cpu := fetchCPU(client, vendor)
	mem := fetchMemoryUsedPercent(client)

	ifaces, rxTotal, txTotal := fetchInterfaces(client, device.ID, now, prior)

	return model.NetPollSample{
		DeviceID:          device.ID,
		CPUPercent:        round2(cpu),
		MemoryUsedPercent: round2(mem),
		UptimeSeconds:     uptime,
		BandwidthRxBps:    round2(rxTotal),
		BandwidthTxBps:    round2(txTotal),
		CollectedAt:       now,
		Interfaces:        ifaces,
		Vendor:            vendor,
		SysDescr:          sysDescr,
		SysObjectID:       sysObject,
	}, nil
}

func ProbeSysDescr(target, community string) (string, string, error) {
	client := newClient(target, community)
	if err := client.Connect(); err != nil {
		return "", "", err
	}
	defer client.Conn.Close()
	sysDescr, sysObject, _, err := fetchSystem(client)
	if err != nil {
		return "", "", err
	}
	return sysDescr, IdentifyVendor(sysDescr, sysObject), nil
}

func fetchSystem(client *gosnmp.GoSNMP) (string, string, uint64, error) {
	pkt, err := client.Get([]string{oidSysDescr, oidSysUpTime, oidSysObject})
	if err != nil {
		return "", "", 0, err
	}
	if pkt.Error != gosnmp.NoError || len(pkt.Variables) < 3 {
		return "", "", 0, fmt.Errorf("snmp system get failed")
	}
	descr := pduString(pkt.Variables[0])
	obj := pduString(pkt.Variables[2])
	uptime := uint64(0)
	if v, ok := pkt.Variables[1].Value.(uint32); ok {
		uptime = uint64(v) / 100
	}
	return descr, obj, uptime, nil
}

func fetchCPU(client *gosnmp.GoSNMP, vendor string) float64 {
	switch vendor {
	case "mikrotik":
		if v := scalarFloat(client, oidMikrotikCPU); v > 0 {
			return v
		}
	case "cisco":
		if v := scalarFloat(client, oidCiscoCPU5m); v > 0 {
			return v
		}
	}
	return avgWalkInt(client, oidHrProcessor)
}

func fetchMemoryUsedPercent(client *gosnmp.GoSNMP) float64 {
	descrPkt, err := bulkWalk(client, oidHrStorageDescr())
	if err != nil {
		return 0
	}
	sizePkt, err := bulkWalk(client, "1.3.6.1.2.1.25.2.3.1.5")
	if err != nil {
		return 0
	}
	usedPkt, err := bulkWalk(client, "1.3.6.1.2.1.25.2.3.1.6")
	if err != nil {
		return 0
	}
	descr := walkIndexStringFromPDU(descrPkt)
	sizes := indexMapUint(sizePkt)
	used := indexMapUint(usedPkt)
	for idx, label := range descr {
		lower := strings.ToLower(label)
		if !strings.Contains(lower, "memory") && !strings.Contains(lower, "ram") {
			continue
		}
		sz := sizes[idx]
		if sz == 0 {
			continue
		}
		return float64(used[idx]) / float64(sz) * 100
	}
	for idx, sz := range sizes {
		if sz == 0 {
			continue
		}
		return float64(used[idx]) / float64(sz) * 100
	}
	return 0
}

func walkIndexStringFromPDU(pkt []gosnmp.SnmpPDU) map[int]string {
	out := map[int]string{}
	for _, v := range pkt {
		out[lastOIDIndex(v.Name)] = pduString(v)
	}
	return out
}

func oidHrStorageDescr() string {
	return "1.3.6.1.2.1.25.2.3.1.3"
}

func scalarFloat(client *gosnmp.GoSNMP, oid string) float64 {
	pkt, err := client.Get([]string{oid})
	if err != nil || len(pkt.Variables) == 0 {
		return 0
	}
	switch v := pkt.Variables[0].Value.(type) {
	case int:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

func avgWalkInt(client *gosnmp.GoSNMP, baseOID string) float64 {
	pkt, err := bulkWalk(client, baseOID)
	if err != nil || len(pkt) == 0 {
		return 0
	}
	var sum float64
	var n float64
	for _, v := range pkt {
		switch val := v.Value.(type) {
		case int:
			sum += float64(val)
			n++
		case uint:
			sum += float64(val)
			n++
		case uint32:
			sum += float64(val)
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

func fetchInterfaces(client *gosnmp.GoSNMP, deviceID string, now time.Time, prior priorOctetsFn) ([]model.NetInterfacePoll, float64, float64) {
	names := walkIndexString(client, oidIfDescr)
	oper := walkIndexInt(client, oidIfOper)
	speed := walkIndexUint64(client, oidIfSpeed)
	inOct := walkIndexUint64(client, oidIfInOctets)
	outOct := walkIndexUint64(client, oidIfOutOctets)

	var rxTotal, txTotal float64
	out := make([]model.NetInterfacePoll, 0, len(names))
	for idx, name := range names {
		if strings.HasPrefix(strings.ToLower(name), "lo") {
			continue
		}
		in := inOct[idx]
		outb := outOct[idx]
		var rx, tx float64
		if prevIn, prevOut, prevAt, ok := prior(deviceID, idx); ok && prevAt.Before(now) {
			elapsed := now.Sub(prevAt)
			rx = CounterBps(prevIn, in, elapsed)
			tx = CounterBps(prevOut, outb, elapsed)
		}
		rxTotal += rx
		txTotal += tx
		out = append(out, model.NetInterfacePoll{
			IfIndex:    idx,
			Name:       name,
			OperStatus: oper[idx],
			SpeedBps:   speed[idx],
			InOctets:   in,
			OutOctets:  outb,
			RxBps:      round2(rx),
			TxBps:      round2(tx),
		})
	}
	return out, rxTotal, txTotal
}

func pduString(v gosnmp.SnmpPDU) string {
	switch val := v.Value.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	default:
		return fmt.Sprint(val)
	}
}

func walkIndexString(client *gosnmp.GoSNMP, base string) map[int]string {
	pkt, err := bulkWalk(client, base)
	if err != nil {
		return map[int]string{}
	}
	out := map[int]string{}
	for _, v := range pkt {
		idx := lastOIDIndex(v.Name)
		out[idx] = pduString(v)
	}
	return out
}

func walkIndexInt(client *gosnmp.GoSNMP, base string) map[int]int {
	pkt, err := bulkWalk(client, base)
	if err != nil {
		return map[int]int{}
	}
	out := map[int]int{}
	for _, v := range pkt {
		idx := lastOIDIndex(v.Name)
		switch val := v.Value.(type) {
		case int:
			out[idx] = val
		case uint:
			out[idx] = int(val)
		case uint32:
			out[idx] = int(val)
		}
	}
	return out
}

func walkIndexUint64(client *gosnmp.GoSNMP, base string) map[int]uint64 {
	pkt, err := bulkWalk(client, base)
	if err != nil {
		return map[int]uint64{}
	}
	out := map[int]uint64{}
	for _, v := range pkt {
		idx := lastOIDIndex(v.Name)
		switch val := v.Value.(type) {
		case uint:
			out[idx] = uint64(val)
		case uint32:
			out[idx] = uint64(val)
		case uint64:
			out[idx] = val
		case int:
			if val >= 0 {
				out[idx] = uint64(val)
			}
		}
	}
	return out
}

func indexMapUint(pkt []gosnmp.SnmpPDU) map[int]uint64 {
	out := map[int]uint64{}
	for _, v := range pkt {
		idx := lastOIDIndex(v.Name)
		switch val := v.Value.(type) {
		case uint:
			out[idx] = uint64(val)
		case uint32:
			out[idx] = uint64(val)
		case uint64:
			out[idx] = val
		case int:
			if val >= 0 {
				out[idx] = uint64(val)
			}
		}
	}
	return out
}

func lastOIDIndex(oid string) int {
	parts := strings.Split(oid, ".")
	if len(parts) == 0 {
		return 0
	}
	last := parts[len(parts)-1]
	var idx int
	fmt.Sscanf(last, "%d", &idx)
	return idx
}
