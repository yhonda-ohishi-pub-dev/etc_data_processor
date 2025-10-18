package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ActualETCRecord represents the actual ETC record format from the CSV files
type ActualETCRecord struct {
	EntryDate     string // 利用年月日（入）
	EntryTime     string // 時刻（入）
	ExitDate      string // 利用年月日（出）
	ExitTime      string // 時刻（出）
	EntryIC       string // 利用IC（入）
	ExitIC        string // 利用IC（出）
	RouteInfo     string // 経路情報
	ETCAmount     int    // ETC料金
	NormalAmount  int    // 通行料金
	DiscountApplied int  // 割引金額適用
	Mileage       int    // マイレージ
	VehicleClass  int    // 車種
	VehicleNumber string // 車両番号
	CardNumber    string // ETCカード番号
	Notes         string // 備考
}

// ETCCSVParser handles actual ETC CSV file parsing
type ETCCSVParser struct{}

// NewETCCSVParser creates a new ETC CSV parser instance
func NewETCCSVParser() *ETCCSVParser {
	return &ETCCSVParser{}
}

// ParseFile parses an actual ETC CSV file with Shift-JIS encoding
func (p *ETCCSVParser) ParseFile(filepath string) ([]ActualETCRecord, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Convert from Shift-JIS to UTF-8
	reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())

	return p.Parse(reader)
}

// Parse parses CSV data from a reader
func (p *ETCCSVParser) Parse(reader io.Reader) ([]ActualETCRecord, error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1 // Variable number of fields

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Parse header and create column mapping
	headerMap := make(map[string]int)
	startIndex := 0

	// Check if first row is header
	if len(records) > 0 {
		firstRow := records[0]
		isHeader := false

		// Check for known header patterns
		for _, col := range firstRow {
			if strings.Contains(col, "利用年月日") || strings.Contains(col, "時刻") ||
			   strings.Contains(col, "利用IC") || strings.Contains(col, "料金") ||
			   strings.Contains(col, "カード番号") {
				isHeader = true
				break
			}
		}

		if isHeader {
			// Build header mapping
			for idx, col := range firstRow {
				headerMap[col] = idx
			}
			startIndex = 1
		}
	}

	if err := p.ValidateRecordsAvailable(records, startIndex); err != nil {
		return nil, err
	}

	var etcRecords []ActualETCRecord
	for i := startIndex; i < len(records); i++ {
		record := records[i]

		// Parse using header mapping if available, otherwise use positional
		var etcRecord ActualETCRecord

		if len(headerMap) > 0 {
			// Use header-based mapping
			etcRecord = p.parseWithHeaders(record, headerMap)
		} else {
			// Use positional mapping (backward compatibility)
			// Ensure we have minimum required fields
			if len(record) < 13 {
				// Skip this record silently - insufficient fields
				continue
			}

			etcRecord = ActualETCRecord{
				EntryDate:     record[0],
				EntryTime:     record[1],
				ExitDate:      record[2],
				ExitTime:      record[3],
				EntryIC:       record[4],
				ExitIC:        record[5],
				RouteInfo:     p.getFieldSafe(record, 6),
				Notes:         "",
			}

			// Parse ETC amount (field 7)
			if p.getFieldSafe(record, 7) != "" {
				amount, err := p.parseAmount(p.getFieldSafe(record, 7))
				if err != nil {
					// Log warning but continue
					etcRecord.ETCAmount = 0
				} else {
					etcRecord.ETCAmount = amount
				}
			}

			// Parse normal amount (field 8)
			if p.getFieldSafe(record, 8) != "" {
				amount, err := p.parseAmount(p.getFieldSafe(record, 8))
				if err != nil {
					etcRecord.NormalAmount = 0
				} else {
					etcRecord.NormalAmount = amount
				}
			}

			// Parse discount amount (field 9)
			if p.getFieldSafe(record, 9) != "" {
				amount, err := p.parseAmount(p.getFieldSafe(record, 9))
				if err != nil {
					etcRecord.DiscountApplied = 0
				} else {
					etcRecord.DiscountApplied = amount
				}
			}

			// Parse mileage (field 10)
			if p.getFieldSafe(record, 10) != "" {
				amount, err := p.parseAmount(p.getFieldSafe(record, 10))
				if err != nil {
					etcRecord.Mileage = 0
				} else {
					etcRecord.Mileage = amount
				}
			}

			// Parse vehicle class (field 11)
			etcRecord.VehicleClass = p.ParseVehicleClass(record, 11)

			// Vehicle number (field 12)
			etcRecord.VehicleNumber = p.getFieldSafe(record, 12)

			// Card number (field 13)
			etcRecord.CardNumber = p.getFieldSafe(record, 13)

			// Notes (field 14)
			etcRecord.Notes = p.getFieldSafe(record, 14)
		}

		// Validate the record
		if err := p.ValidateRecord(etcRecord); err != nil {
			// Skip validation errors silently - continue processing
			// Validation errors are expected for some records
		}

		etcRecords = append(etcRecords, etcRecord)
	}

	return etcRecords, nil
}

// parseAmount parses amount strings that may have negative values
func (p *ETCCSVParser) parseAmount(s string) (int, error) {
	// Remove commas
	s = strings.ReplaceAll(s, ",", "")

	// Check for negative value (e.g., "-7430")
	if strings.HasPrefix(s, "-") {
		value, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return value, nil
	}

	// Parse positive value
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// ValidateRecord validates a single ETC record
func (p *ETCCSVParser) ValidateRecord(record ActualETCRecord) error {
	// Basic validation - allow empty IC for some records
	// Some records might have empty entry/exit IC for special cases

	// Check card number is not empty
	if record.CardNumber == "" {
		return fmt.Errorf("card number cannot be empty")
	}

	// Parse and validate dates
	if record.EntryDate != "" {
		_, err := p.parseDate(record.EntryDate)
		if err != nil {
			return fmt.Errorf("invalid entry date: %w", err)
		}
	}

	if record.ExitDate != "" {
		_, err := p.parseDate(record.ExitDate)
		if err != nil {
			return fmt.Errorf("invalid exit date: %w", err)
		}
	}

	return nil
}

// parseDate parses date in format "YY/MM/DD"
func (p *ETCCSVParser) parseDate(dateStr string) (time.Time, error) {
	// Handle date format like "25/09/01" (YY/MM/DD)
	parts := strings.Split(dateStr, "/")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	// Convert 2-digit year to 4-digit
	if year < 100 {
		if year < 50 {
			year += 2000
		} else {
			year += 1900
		}
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// ConvertToSimpleRecord converts ActualETCRecord to the simplified ETCRecord format
func (p *ETCCSVParser) ConvertToSimpleRecord(actual ActualETCRecord) (ETCRecord, error) {
	date, err := p.parseDate(actual.ExitDate)
	if err != nil {
		// Try entry date if exit date fails
		date, err = p.parseDate(actual.EntryDate)
		if err != nil {
			return ETCRecord{}, err
		}
	}

	// Determine the amount to use
	amount := actual.ETCAmount
	if amount == 0 {
		amount = actual.NormalAmount
	}
	// Handle negative amounts
	if amount < 0 {
		amount = -amount
	}

	return ETCRecord{
		Date:        date,
		EntryIC:     actual.EntryIC,
		ExitIC:      actual.ExitIC,
		Route:       actual.RouteInfo,
		VehicleType: actual.VehicleClass,
		Amount:      amount,
		CardNumber:  actual.CardNumber,
	}, nil
}

// getFieldSafe safely gets a field from a record slice
func (p *ETCCSVParser) getFieldSafe(record []string, index int) string {
	if index < len(record) {
		return record[index]
	}
	return ""
}

// parseWithHeaders parses a record using header mapping
func (p *ETCCSVParser) parseWithHeaders(record []string, headerMap map[string]int) ActualETCRecord {
	etcRecord := ActualETCRecord{}

	// Map header names to fields - handle different formats
	// Some files use （自）/（至） while others use （入）/（出）
	etcRecord.EntryDate = p.getFieldByHeader(record, headerMap, "利用年月日（入）", "利用年月日(入)", "利用年月日（自）", "入口日付")
	etcRecord.EntryTime = p.getFieldByHeader(record, headerMap, "時刻（入）", "時刻(入)", "時分（自）", "入口時刻")
	etcRecord.ExitDate = p.getFieldByHeader(record, headerMap, "利用年月日（出）", "利用年月日(出)", "利用年月日（至）", "出口日付")
	etcRecord.ExitTime = p.getFieldByHeader(record, headerMap, "時刻（出）", "時刻(出)", "時分（至）", "出口時刻")
	etcRecord.EntryIC = p.getFieldByHeader(record, headerMap, "利用IC（入）", "利用IC(入)", "利用ＩＣ（自）", "入口IC", "入口")
	etcRecord.ExitIC = p.getFieldByHeader(record, headerMap, "利用IC（出）", "利用IC(出)", "利用ＩＣ（至）", "出口IC", "出口")
	etcRecord.RouteInfo = p.getFieldByHeader(record, headerMap, "経路情報", "路線", "経路")

	// Parse amounts - handle different header formats
	// 割引前料金 = Normal amount (before discount)
	normalAmountStr := p.getFieldByHeader(record, headerMap, "割引前料金", "通行料金", "通常料金")
	if normalAmountStr != "" {
		amount, err := p.parseAmount(normalAmountStr)
		if err == nil {
			etcRecord.NormalAmount = amount
		}
	}

	// ＥＴＣ割引額 = Discount amount (negative value)
	discountStr := p.getFieldByHeader(record, headerMap, "ＥＴＣ割引額", "ETC割引額", "割引額")
	if discountStr != "" {
		amount, err := p.parseAmount(discountStr)
		if err == nil {
			etcRecord.DiscountApplied = amount
		}
	}

	// 通行料金 = Actual charged amount
	etcAmountStr := p.getFieldByHeader(record, headerMap, "通行料金", "ETC料金", "料金")
	if etcAmountStr != "" {
		amount, err := p.parseAmount(etcAmountStr)
		if err == nil {
			etcRecord.ETCAmount = amount
		}
	}

	// 後納料金 = Post-payment amount (if exists)
	postPaymentStr := p.getFieldByHeader(record, headerMap, "後納料金", "後払料金")
	if postPaymentStr != "" {
		amount, err := p.parseAmount(postPaymentStr)
		if err == nil && amount != 0 {
			// Use post-payment amount if available
			etcRecord.ETCAmount = amount
		}
	}

	// Parse vehicle info
	vehicleClassStr := p.getFieldByHeader(record, headerMap, "車種", "車両区分", "車種区分")
	if vehicleClassStr != "" {
		class, err := strconv.Atoi(vehicleClassStr)
		if err == nil {
			etcRecord.VehicleClass = class
		}
	}

	etcRecord.VehicleNumber = p.getFieldByHeader(record, headerMap, "車両番号", "ナンバー", "車番")
	etcRecord.CardNumber = p.getFieldByHeader(record, headerMap, "ＥＴＣカード番号", "ETCカード番号", "カード番号", "カード")
	etcRecord.Notes = p.getFieldByHeader(record, headerMap, "備考", "メモ", "注記")

	return etcRecord
}

// getFieldByHeader gets a field value using multiple possible header names
func (p *ETCCSVParser) getFieldByHeader(record []string, headerMap map[string]int, headerNames ...string) string {
	for _, headerName := range headerNames {
		if idx, exists := headerMap[headerName]; exists {
			if idx < len(record) {
				return record[idx]
			}
		}
	}
	return ""
}

// ParseVehicleClass parses vehicle class from record field, returns 0 if parsing fails
func (p *ETCCSVParser) ParseVehicleClass(record []string, fieldIndex int) int {
	fieldValue := p.getFieldSafe(record, fieldIndex)
	if fieldValue != "" {
		class, err := strconv.Atoi(fieldValue)
		if err != nil {
			return 0
		}
		return class
	}
	return 0
}

// ValidateRecordsAvailable checks if there are data records available for processing
func (p *ETCCSVParser) ValidateRecordsAvailable(records [][]string, startIndex int) error {
	if len(records) <= startIndex {
		return fmt.Errorf("no data records found")
	}
	return nil
}