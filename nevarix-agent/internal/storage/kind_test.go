package storage

import "testing"

func TestKindFromRotational(t *testing.T) {
	if KindFromRotational("0") != "ssd" || KindFromRotational("1") != "hdd" {
		t.Fatal("rotational mapping")
	}
}

func TestKindFromMedia(t *testing.T) {
	if KindFromMedia("SSD") != "ssd" || KindFromMedia("HDD") != "hdd" {
		t.Fatal("media mapping")
	}
}

func TestSkipDevice(t *testing.T) {
	if !SkipDevice("loop0") || SkipDevice("nvme0n1") || SkipDevice("sda") {
		t.Fatal("skip filter")
	}
}
