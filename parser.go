package goqris

import (
	"fmt"
	"strconv"
)

// TLV (Tag-Length-Value) from QRIS specification
type TLV struct {
	Tag      string
	Name     string
	Length   int
	Value    string
	Children []*TLV
}

type QRISMethod string

const (
	QRISStatic  QRISMethod = "static"
	QRISDynamic QRISMethod = "dynamic"
)

type TipIndicator string

func (t TipIndicator) IsValid() bool {
	switch t {
	case TipPrompt, TipFixed, TipPercentage:
		return true
	default:
		return false
	}
}

const (
	TipPrompt     TipIndicator = "prompt"
	TipFixed      TipIndicator = "fixed"
	TipPercentage TipIndicator = "percentage"
)

// QRISData represents human readable data extracted from a QRIS code
type QRISData struct {
	Version              string
	Method               QRISMethod
	Merchants            []MerchantInfo
	MerchantCategoryCode string
	MerchantName         string
	MerchantCity         string
	PostalCode           string
	CurrencyCode         string
	Amount               int
	TipIndicator         TipIndicator
	TipFixedAmount       int
	TipPercentage        int
	CountryCode          string
	CRC                  string
	AdditionalData       []*TLV
	TLVs                 []*TLV
}

// MerchantInfo represents the merchant information extracted from a QRIS code
type MerchantInfo struct {
	Tag              string
	UniqueID         string
	MerchantID       string
	MerchantCriteria string
	Fields           []*TLV
}

// TAGS maps QRIS tags to their human-readable names
var TAGS = map[string]string{
	"00": "Payload Format Indicator",
	"01": "Point of Initiation Method",
	"02": "Visa",
	"03": "Mastercard",
	"04": "Mastercard",
	"15": "Visa",
	"26": "Merchant Account Information",
	"27": "Merchant Account Information",
	"28": "Merchant Account Information",
	"29": "Merchant Account Information",
	"30": "Merchant Account Information",
	"31": "Merchant Account Information",
	"32": "Merchant Account Information",
	"33": "Merchant Account Information",
	"34": "Merchant Account Information",
	"35": "Merchant Account Information",
	"36": "Merchant Account Information",
	"37": "Merchant Account Information",
	"38": "Merchant Account Information",
	"39": "Merchant Account Information",
	"40": "Merchant Account Information",
	"41": "Merchant Account Information",
	"42": "Merchant Account Information",
	"43": "Merchant Account Information",
	"44": "Merchant Account Information",
	"45": "Merchant Account Information",
	"46": "Merchant Account Information",
	"47": "Merchant Account Information",
	"48": "Merchant Account Information",
	"49": "Merchant Account Information",
	"50": "Merchant Account Information",
	"51": "Merchant Account Information",
	"52": "Merchant Category Code",
	"53": "Transaction Currency",
	"54": "Transaction Amount",
	"55": "Tip or Convenience Indicator",
	"56": "Value of Convenience Fee (Fixed)",
	"57": "Value of Convenience Fee (%)",
	"58": "Country Code",
	"59": "Merchant Name",
	"60": "Merchant City",
	"61": "Postal Code",
	"62": "Additional Data Field",
	"63": "CRC",
}

var nestedTags map[string]struct{}

func init() {
	nestedTags = make(map[string]struct{}, 27)
	for i := 26; i <= 51; i++ {
		tag := fmt.Sprintf("%02d", i)
		nestedTags[tag] = struct{}{}
	}
	nestedTags["62"] = struct{}{}
}

func parseTLV(qris string) []*TLV {
	tlvs := make([]*TLV, 0)

	qrisLength := len(qris)
	blockLength := 4
	partLength := 2

	pos := 0

	for pos < qrisLength {
		if pos+blockLength > qrisLength {
			break
		}

		tag := qris[pos : pos+partLength]
		tagLength := strToInt(qris[pos+partLength : pos+blockLength])

		if tagLength == 0 || pos+blockLength+tagLength > qrisLength {
			break
		}

		tagValue := qris[pos+blockLength : pos+blockLength+tagLength]
		tagName := fmt.Sprintf("Unknown (%s)", tag)
		if name, exists := TAGS[tag]; exists {
			tagName = name
		}

		tlv := TLV{
			Tag:      tag,
			Name:     tagName,
			Length:   tagLength,
			Value:    tagValue,
			Children: nil,
		}

		if _, isNested := nestedTags[tag]; isNested {
			tlv.Children = parseTLV(tagValue)
		}

		tlvs = append(tlvs, &tlv)
		pos += blockLength + tagLength
	}

	return tlvs
}

func findTag(tlvs []*TLV, tag string) (*TLV, bool) {
	for _, tlv := range tlvs {
		if tlv.Tag == tag {
			return tlv, true
		}
	}
	return nil, false
}

func ParseQRISData(qris string) QRISData {
	tlvs := parseTLV(qris)

	qrisData := QRISData{
		Version:      "01",
		Method:       QRISStatic,
		CurrencyCode: "360", // Default to IDR
		CountryCode:  "ID",  // Default to Indonesia
		TLVs:         tlvs,
	}

	if tag00, exists := findTag(tlvs, "00"); exists {
		qrisData.Version = tag00.Value
	}

	if tag01, exists := findTag(tlvs, "01"); exists && tag01.Value == "12" {
		qrisData.Method = QRISDynamic
	}

	if tag52, exists := findTag(tlvs, "52"); exists {
		qrisData.MerchantCategoryCode = tag52.Value
	}

	if tag53, exists := findTag(tlvs, "53"); exists {
		qrisData.CurrencyCode = tag53.Value
	}

	if tag54, exists := findTag(tlvs, "54"); exists {
		qrisData.Amount = strToInt(tag54.Value)
	}

	if tag55, exists := findTag(tlvs, "55"); exists {
		switch tag55.Value {
		case "01":
			qrisData.TipIndicator = TipPrompt
		case "02":
			qrisData.TipIndicator = TipFixed
		case "03":
			qrisData.TipIndicator = TipPercentage
		}
	}

	if tag56, exists := findTag(tlvs, "56"); exists {
		qrisData.TipFixedAmount = strToInt(tag56.Value)
	}

	if tag57, exists := findTag(tlvs, "57"); exists {
		qrisData.TipPercentage = strToInt(tag57.Value)
	}

	if tag58, exists := findTag(tlvs, "58"); exists {
		qrisData.CountryCode = tag58.Value
	}

	if tag59, exists := findTag(tlvs, "59"); exists {
		qrisData.MerchantName = tag59.Value
	}

	if tag60, exists := findTag(tlvs, "60"); exists {
		qrisData.MerchantCity = tag60.Value
	}

	if tag61, exists := findTag(tlvs, "61"); exists {
		qrisData.PostalCode = tag61.Value
	}

	if tag62, exists := findTag(tlvs, "62"); exists {
		qrisData.AdditionalData = tag62.Children
	}

	if tag63, exists := findTag(tlvs, "63"); exists {
		qrisData.CRC = tag63.Value
	}

	var merchants []MerchantInfo
	// extract merchant information from tag 26-51
	for _, tlv := range tlvs {
		tag := strToInt(tlv.Tag)
		if len(tlv.Children) == 0 || tag < 26 || tag > 51 {
			continue
		}

		merchant := MerchantInfo{
			Tag:    tlv.Tag,
			Fields: tlv.Children,
		}

		if tag00, exists := findTag(tlv.Children, "00"); exists {
			merchant.UniqueID = tag00.Value
		}

		if tag01, exists := findTag(tlv.Children, "01"); exists {
			merchant.MerchantID = tag01.Value
		} else if tag02, exists := findTag(tlv.Children, "02"); exists {
			merchant.MerchantID = tag02.Value
		}

		if tag03, exists := findTag(tlv.Children, "03"); exists {
			merchant.MerchantCriteria = tag03.Value
		}

		merchants = append(merchants, merchant)
	}

	qrisData.Merchants = merchants

	return qrisData
}

func strToInt(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
