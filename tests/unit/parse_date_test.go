package unit

import (
	"testing"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

// Test parseDate function directly - this is a private method, so we'll test through ValidateRecord
func TestETCCSVParser_ParseDate_Validation(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name        string
		entryDate   string
		exitDate    string
		shouldError bool
		description string
	}{
		{
			name:        "valid 2-digit year less than 50",
			entryDate:   "25/09/01", // Should become 2025
			exitDate:    "25/09/01",
			shouldError: false,
			description: "2-digit year < 50 should become 20xx",
		},
		{
			name:        "valid 2-digit year greater than 50",
			entryDate:   "85/09/01", // Should become 1985
			exitDate:    "85/09/01",
			shouldError: false,
			description: "2-digit year >= 50 should become 19xx",
		},
		{
			name:        "valid 4-digit year",
			entryDate:   "2025/09/01", // Should stay 2025
			exitDate:    "2025/09/01",
			shouldError: false,
			description: "4-digit year should remain unchanged",
		},
		{
			name:        "invalid date format - not 3 parts",
			entryDate:   "25/09", // Missing day
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Date with less than 3 parts should error",
		},
		{
			name:        "invalid date format - too many parts",
			entryDate:   "25/09/01/extra", // Too many parts
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Date with more than 3 parts should error",
		},
		{
			name:        "invalid year - non-numeric",
			entryDate:   "xx/09/01",
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Non-numeric year should error",
		},
		{
			name:        "invalid month - non-numeric",
			entryDate:   "25/xx/01",
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Non-numeric month should error",
		},
		{
			name:        "invalid day - non-numeric",
			entryDate:   "25/09/xx",
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Non-numeric day should error",
		},
		{
			name:        "year exactly 50",
			entryDate:   "50/09/01", // Should become 1950
			exitDate:    "50/09/01",
			shouldError: false,
			description: "Year exactly 50 should become 1950",
		},
		{
			name:        "year exactly 49",
			entryDate:   "49/09/01", // Should become 2049
			exitDate:    "49/09/01",
			shouldError: false,
			description: "Year exactly 49 should become 2049",
		},
		{
			name:        "empty date string",
			entryDate:   "",
			exitDate:    "25/09/01",
			shouldError: false,
			description: "Empty date is allowed in ValidateRecord",
		},
		{
			name:        "single slash",
			entryDate:   "/",
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Single slash should error",
		},
		{
			name:        "only slashes",
			entryDate:   "//",
			exitDate:    "25/09/01",
			shouldError: true,
			description: "Only slashes should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test record with the date to test
			record := parser.ActualETCRecord{
				EntryDate:   tt.entryDate,
				ExitDate:    tt.exitDate,
				CardNumber:  "1234567890", // Required field
			}

			err := p.ValidateRecord(record)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil. %s", tt.name, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v. %s", tt.name, err, tt.description)
				}
			}
		})
	}
}

// Test parseDate function edge cases through ConvertToSimpleRecord
func TestETCCSVParser_ParseDate_Conversion(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name        string
		entryDate   string
		exitDate    string
		shouldError bool
		expectedYear int
		description string
	}{
		{
			name:        "conversion with valid exit date",
			entryDate:   "invalid", // This will fail
			exitDate:    "25/09/01", // This should work
			shouldError: false,
			expectedYear: 2025,
			description: "Should use exit date when entry date is invalid",
		},
		{
			name:        "conversion with both dates invalid",
			entryDate:   "invalid",
			exitDate:    "also/invalid",
			shouldError: true,
			description: "Should error when both dates are invalid",
		},
		{
			name:        "year boundary testing - 2000",
			entryDate:   "00/01/01", // Should become 2000
			exitDate:    "00/01/01",
			shouldError: false,
			expectedYear: 2000,
			description: "Year 00 should become 2000",
		},
		{
			name:        "year boundary testing - 1999",
			entryDate:   "99/01/01", // Should become 1999
			exitDate:    "99/01/01",
			shouldError: false,
			expectedYear: 1999,
			description: "Year 99 should become 1999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := parser.ActualETCRecord{
				EntryDate:   tt.entryDate,
				ExitDate:    tt.exitDate,
				CardNumber:  "1234567890",
				ETCAmount:   1000,
			}

			result, err := p.ConvertToSimpleRecord(record)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil. %s", tt.name, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v. %s", tt.name, err, tt.description)
				} else if tt.expectedYear > 0 {
					if result.Date.Year() != tt.expectedYear {
						t.Errorf("Expected year %d for %s, got %d. %s", tt.expectedYear, tt.name, result.Date.Year(), tt.description)
					}
				}
			}
		})
	}
}