package goqris

import (
	"strings"
	"testing"
)

func TestToDynamic_AddsAmountAndCRC(t *testing.T) {
	base := "000201010211" // minimal with tag 00 and tag 01 = 11 (static)
	crc := calculateCRC16String(base)
	qris := base + crc

	out, err := ToDynamic(qris, DynamicOption{Amount: 100})
	if err != nil {
		t.Fatalf("ToDynamic returned error: %v", err)
	}

	// Expect tag 54 with value 100 -> "5403100"
	if !strings.Contains(out, "5403100") {
		t.Fatalf("expected output to contain amount TLV 5403100, got %s", out)
	}

	// Verify CRC at end is valid for the output without its CRC
	if len(out) < 4 {
		t.Fatalf("output too short to contain CRC: %s", out)
	}
	without := out[:len(out)-4]
	expected := calculateCRC16String(without)
	got := out[len(out)-4:]
	if expected != got {
		t.Fatalf("crc mismatch: expected %s got %s", expected, got)
	}
}

func BenchmarkParseTLV(b *testing.B) {
	sample := "0002010102115303605908Merchant60City620705ABC06304ABCD"
	// repeat to increase size
	long := strings.Repeat(sample, 50)
	b.ReportAllocs()
	for b.Loop() {
		_ = parseTLV(long)
	}
}

func BenchmarkToDynamic(b *testing.B) {
	base := "000201010211"
	base = base + calculateCRC16String(base)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = ToDynamic(base, DynamicOption{Amount: 1000})
	}
}

func TestE2E_QRISExample(t *testing.T) {
	qris := "00020101021126570011ID.DANA.WWW011893600915362751266202096275126620303UMI51440014ID.CO.QRIS.WWW0215ID10243193490550303UMI5204594553033605802ID5910dpangestuw6012Kab. Bandung6105402386304AC2C"

	// Sanity-parse the input and verify it's static and merchant name
	data := ParseQRISData(qris)
	if data.Method != QRISStatic {
		t.Fatalf("expected input to be static QRIS, got %v", data.Method)
	}
	if data.MerchantName != "dpangestuw" {
		t.Fatalf("unexpected merchant name: %q", data.MerchantName)
	}

	// Convert to dynamic with an amount and verify TLVs and CRC
	amount := 50000
	out, err := ToDynamic(qris, DynamicOption{Amount: amount})
	if err != nil {
		t.Fatalf("ToDynamic error: %v", err)
	}

	// Parse output and verify tag 01 changed to dynamic indicator (12)
	outTLVs := parseTLV(out)
	if tag01, ok := findTag(outTLVs, "01"); !ok || tag01.Value != "12" {
		t.Fatalf("expected tag 01 value '12' in output, got %#v", tag01)
	}

	// Verify amount (tag 54) was set correctly
	if tag54, ok := findTag(outTLVs, "54"); !ok {
		t.Fatalf("missing tag 54 in output")
	} else if strToInt(tag54.Value) != amount {
		t.Fatalf("tag54 amount mismatch: expected %d got %s", amount, tag54.Value)
	}

	// Validate CRC for produced output
	if err := validateChecksum(out); err != nil {
		t.Fatalf("output CRC invalid: %v", err)
	}
}

func TestToDynamic_WithFixedTip(t *testing.T) {
	qris := "00020101021126570011ID.DANA.WWW011893600915362751266202096275126620303UMI51440014ID.CO.QRIS.WWW0215ID10243193490550303UMI5204594553033605802ID5910dpangestuw6012Kab. Bandung6105402386304AC2C"

	out, err := ToDynamic(qris, DynamicOption{
		Amount: 25000,
		Tip: &DynamicTipOption{
			TipType:   TipFixed,
			TipAmount: 2000,
		},
	})
	if err != nil {
		t.Fatalf("ToDynamic error: %v", err)
	}

	outTLVs := parseTLV(out)
	if tag55, ok := findTag(outTLVs, "55"); !ok {
		t.Fatalf("missing tag 55 in output")
	} else if tag55.Value != "02" {
		t.Fatalf("expected tag 55 value 02 for fixed tip, got %s", tag55.Value)
	}

	if tag56, ok := findTag(outTLVs, "56"); !ok {
		t.Fatalf("missing tag 56 in output")
	} else if strToInt(tag56.Value) != 2000 {
		t.Fatalf("expected tag 56 value 2000, got %s", tag56.Value)
	}

	if tag54, ok := findTag(outTLVs, "54"); !ok {
		t.Fatalf("missing tag 54 in output")
	} else if strToInt(tag54.Value) != 25000 {
		t.Fatalf("tag54 amount mismatch: expected %d got %s", 25000, tag54.Value)
	}

	if err := validateChecksum(out); err != nil {
		t.Fatalf("output CRC invalid: %v", err)
	}
}
