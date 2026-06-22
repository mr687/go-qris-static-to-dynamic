package goqris

import (
	"cmp"
	"errors"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrDynamicAmountRequired = errors.New("dynamic amount is required for dynamic QRIS")
	ErrQRISNotStatic         = errors.New("invalid, QRIS is not static")
	ErrQRISInvalidChecksum   = errors.New("invalid QRIS checksum")
	ErrDynamicTipTypeInvalid = errors.New("invalid tip type, must be 'fixed', or 'percentage'")
)

type DynamicTipOption struct {
	TipType   TipIndicator
	TipAmount int
}

type DynamicOption struct {
	Amount int
	Tip    *DynamicTipOption
}

func ToDynamic(qrisStatic string, option DynamicOption) (string, error) {
	if option.Amount <= 0 {
		return "", ErrDynamicAmountRequired
	}

	if option.Tip != nil && !option.Tip.TipType.IsValid() {
		return "", ErrDynamicTipTypeInvalid
	}

	// validate static QRIS
	if err := ValidateCRC(qrisStatic); err != nil {
		return "", err
	}

	tlvs := parseTLV(qrisStatic)

	// check if QRIS is static
	if tag01, exists := findTag(tlvs, "01"); exists {
		if tag01.Value != "11" {
			return "", ErrQRISNotStatic
		}
		tag01.Value = "12" // change static to dyanmic
	}

	// create or update tag 54 (transaction amount)
	if tag54, exists := findTag(tlvs, "54"); exists {
		tag54.Value = intToStr(option.Amount)
	} else {
		tag54 := createTLV("54", intToStr(option.Amount))
		tlvs = append(tlvs, &tag54)
	}

	if option.Tip != nil {
		// create tag 55 (tip or convenience fee)
		// and create tag 56 (tip or convenience fee amount) if tip type is fixed
		// or create tag 57 (tip or convenience fee percentage) if tip type is percentage
		switch option.Tip.TipType {
		case TipFixed:
			tag55 := createTLV("55", "02")
			tag56 := createTLV("56", intToStr(option.Tip.TipAmount))
			tlvs = append(tlvs, &tag55)
			tlvs = append(tlvs, &tag56)
		case TipPercentage:
			tag55 := createTLV("55", "03")
			tag57 := createTLV("57", intToStr(option.Tip.TipAmount))
			tlvs = append(tlvs, &tag55)
			tlvs = append(tlvs, &tag57)
		}
	}

	// sort TLVs by tag
	sortTLVs(tlvs)

	// reset CRC16 tag (tag 63) and will be recalculated later
	if tag63, exists := findTag(tlvs, "63"); exists {
		tag63.Value = ""
	}

	input := tlvsToString(tlvs)
	output := input + calculateCRC16String(input)

	return output, nil
}

func tlvsToString(tlvs []*TLV) string {
	var sb strings.Builder

	for _, el := range tlvs {
		valStr := el.Value

		// If children exist, recursively build their TLV string instead
		if len(el.Children) > 0 {
			valStr = tlvsToString(el.Children)
		}

		// Format Length as a 2-character, zero-padded string (e.g., "05", "12")
		// Use strconv to avoid fmt allocations in hot paths.
		length := max(el.Length, 0)
		lenStr := strconv.Itoa(length)
		if length < 10 {
			sb.WriteString(el.Tag)
			sb.WriteByte('0')
			sb.WriteString(lenStr)
		} else {
			sb.WriteString(el.Tag)
			sb.WriteString(lenStr)
		}
		sb.WriteString(valStr)
	}

	return sb.String()
}

func sortTLVs(tlvs []*TLV) {
	slices.SortFunc(tlvs, func(a, b *TLV) int {
		return cmp.Compare(a.Tag, b.Tag)
	})
}

func createTLV(tag string, value string) TLV {
	tagName, exists := TAGS[tag]
	if !exists {
		tagName = "Unknown (" + tag + ")"
	}
	return TLV{
		Tag:    tag,
		Length: len(value),
		Name:   tagName,
		Value:  value,
	}
}

func ValidateStatic(qris string) error {
	if err := ValidateCRC(qris); err != nil {
		return err
	}

	tlvs := parseTLV(qris)
	tag01, exists := findTag(tlvs, "01")
	if !exists || tag01.Value != "11" {
		return ErrQRISNotStatic
	}
	return nil
}

func ValidateCRC(qris string) error {
	if len(qris) < 4 {
		return ErrQRISInvalidChecksum
	}

	data := qris[:len(qris)-4]
	checksum := qris[len(qris)-4:]
	calculatedChecksum := calculateCRC16String(data)
	if checksum != calculatedChecksum {
		return ErrQRISInvalidChecksum
	}

	return nil
}

func calculateCRC16Bytes(data []byte) string {
	const polynomial uint16 = 0x1021
	var crc uint16 = 0xFFFF

	for _, b := range data {
		crc ^= uint16(b) << 8
		for range 8 {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc <<= 1
			}
		}
	}

	// uppercase hex, 4 chars, padded with zeros
	h := strconv.FormatUint(uint64(crc)&0xFFFF, 16)
	// ensure 4 characters
	for len(h) < 4 {
		h = "0" + h
	}
	return strings.ToUpper(h)
}

func calculateCRC16String(s string) string {
	return calculateCRC16Bytes([]byte(s))
}

func intToStr(v int) string {
	return strconv.Itoa(v)
}
