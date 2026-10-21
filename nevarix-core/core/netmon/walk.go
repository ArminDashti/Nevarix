package netmon

import "github.com/gosnmp/gosnmp"

func bulkWalk(client *gosnmp.GoSNMP, rootOID string) ([]gosnmp.SnmpPDU, error) {
	var out []gosnmp.SnmpPDU
	err := client.BulkWalk(rootOID, func(pdu gosnmp.SnmpPDU) error {
		out = append(out, pdu)
		return nil
	})
	return out, err
}
