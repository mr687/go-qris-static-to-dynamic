package main

import (
    "strings"
    "testing"
)

func TestParseTLV_Basic(t *testing.T) {
    // "00" tag, length 02, value "01"
    input := "000201"
    tlvs := parseTLV(input)
    if len(tlvs) != 1 {
        t.Fatalf("expected 1 tlv, got %d", len(tlvs))
    }
    if tlvs[0].Tag != "00" {
        t.Fatalf("expected tag 00, got %s", tlvs[0].Tag)
    }
    if tlvs[0].Value != "01" {
        t.Fatalf("expected value 01, got %s", tlvs[0].Value)
    }
    if tlvs[0].Length != 2 {
        t.Fatalf("expected length 2, got %d", tlvs[0].Length)
    }
}

func TestCalculateCRC16String_Format(t *testing.T) {
    sample := "000201010211"
    crc := calculateCRC16String(sample)
    if len(crc) != 4 {
        t.Fatalf("expected crc length 4, got %d", len(crc))
    }
    // should be uppercase hex digits
    if strings.ToUpper(crc) != crc {
        t.Fatalf("expected uppercase hex crc, got %s", crc)
    }
}
