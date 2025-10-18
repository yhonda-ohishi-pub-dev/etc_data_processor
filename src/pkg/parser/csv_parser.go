package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// ETCRecord represents a single ETC toll record
type ETCRecord struct {
	Date        time.Time
	EntryIC     string
	ExitIC      string
	Route       string
	VehicleType int
	Amount      int
	CardNumber  string
}

// CSVParser handles CSV file parsing
type CSVParser struct{}

// NewCSVParser creates a new CSV parser instance
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// Parse parses CSV data from a reader
func (p *CSVParser) Parse(reader io.Reader) ([]ETCRecord, error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = 7

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Skip header if it exists
	startIndex := 0
	if len(records) > 0 && records[0][0] == "日付" {
		startIndex = 1
	}

	if len(records) <= startIndex {
		return nil, fmt.Errorf("no data records found")
	}

	return p.ProcessRecords(records, startIndex)
}

// ProcessRecords processes CSV records starting from the given index
func (p *CSVParser) ProcessRecords(records [][]string, startIndex int) ([]ETCRecord, error) {
	var etcRecords []ETCRecord
	for i := startIndex; i < len(records); i++ {
		record := records[i]

		// Parse date
		date, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			return nil, fmt.Errorf("invalid date format at line %d: %w", i+1, err)
		}

		// Parse amount
		amount, err := strconv.Atoi(record[5])
		if err != nil {
			return nil, fmt.Errorf("invalid amount at line %d: %w", i+1, err)
		}

		// Parse vehicle type
		vehicleType, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("invalid vehicle type at line %d: %w", i+1, err)
		}

		etcRecord := ETCRecord{
			Date:        date,
			EntryIC:     record[1],
			ExitIC:      record[2],
			Route:       record[3],
			VehicleType: vehicleType,
			Amount:      amount,
			CardNumber:  record[6],
		}

		// Validate the record
		if err := p.ValidateRecord(etcRecord); err != nil {
			return nil, fmt.Errorf("validation error at line %d: %w", i+1, err)
		}

		etcRecords = append(etcRecords, etcRecord)
	}

	return etcRecords, nil
}

// ParseFile parses a CSV file from the filesystem
func (p *CSVParser) ParseFile(filepath string) ([]ETCRecord, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// ValidateRecord validates a single ETC record
func (p *CSVParser) ValidateRecord(record ETCRecord) error {
	// Check for empty required fields
	if record.EntryIC == "" {
		return fmt.Errorf("entry IC cannot be empty")
	}
	if record.ExitIC == "" {
		return fmt.Errorf("exit IC cannot be empty")
	}
	if record.Route == "" {
		return fmt.Errorf("route cannot be empty")
	}
	if record.VehicleType == 0 {
		return fmt.Errorf("vehicle type cannot be zero")
	}
	if record.CardNumber == "" {
		return fmt.Errorf("card number cannot be empty")
	}

	// Check amount is non-negative
	if record.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	// Check date is not in the future
	if record.Date.After(time.Now()) {
		return fmt.Errorf("date cannot be in the future")
	}

	// Check date is reasonable (not too old)
	minDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if record.Date.Before(minDate) {
		return fmt.Errorf("date is too old (before year 2000)")
	}

	return nil
}

// ParseStats contains parsing statistics
type ParseStats struct {
	TotalLines     int
	ParsedRecords  int
	SkippedRecords int
	Errors         []string
}