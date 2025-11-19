package unit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

// Test ValidateRecordsAvailable function directly
func TestETCCSVParser_ValidateRecordsAvailable(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name        string
		records     [][]string
		startIndex  int
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "no records",
			records:     [][]string{},
			startIndex:  0,
			shouldError: true,
			errorMsg:    "no data records found",
		},
		{
			name: "start index too large",
			records: [][]string{
				{"header1", "header2", "header3"},
			},
			startIndex:  5,
			shouldError: true,
			errorMsg:    "no data records found",
		},
		{
			name: "start index equals length",
			records: [][]string{
				{"header1", "header2", "header3"},
				{"data1", "data2", "data3"},
			},
			startIndex:  2,
			shouldError: true,
			errorMsg:    "no data records found",
		},
		{
			name: "valid records with start index 0",
			records: [][]string{
				{"data1", "data2", "data3"},
			},
			startIndex:  0,
			shouldError: false,
		},
		{
			name: "valid records with start index 1",
			records: [][]string{
				{"header1", "header2", "header3"},
				{"data1", "data2", "data3"},
			},
			startIndex:  1,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.ValidateRecordsAvailable(tt.records, tt.startIndex)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}

				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// Test ValidateRecordsAvailable with mock records
func TestETCCSVParser_ValidateRecordsAvailable_MockScenarios(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Mock scenario 1: Empty CSV data
	emptyRecords := [][]string{}
	err := p.ValidateRecordsAvailable(emptyRecords, 0)
	if err == nil {
		t.Error("Expected error for empty records, got nil")
	}
	if !strings.Contains(err.Error(), "no data records found") {
		t.Errorf("Expected 'no data records found', got: %v", err)
	}

	// Mock scenario 2: Only header, no data
	headerOnlyRecords := [][]string{
		{"利用年月日（自）", "時分（自）", "利用年月日（至）", "時分（至）", "利用ＩＣ（自）", "利用ＩＣ（至）", "通行料金", "車種", "ＥＴＣカード番号"},
	}
	err = p.ValidateRecordsAvailable(headerOnlyRecords, 1)
	if err == nil {
		t.Error("Expected error for header-only records, got nil")
	}
	if !strings.Contains(err.Error(), "no data records found") {
		t.Errorf("Expected 'no data records found', got: %v", err)
	}

	// Mock scenario 3: Valid data with header
	validRecords := [][]string{
		{"利用年月日（自）", "時分（自）", "利用年月日（至）", "時分（至）", "利用ＩＣ（自）", "利用ＩＣ（至）", "通行料金", "車種", "ＥＴＣカード番号"},
		{"25/09/01", "08:00", "25/09/01", "09:00", "東京", "横浜", "1200", "2", "********12345678"},
	}
	err = p.ValidateRecordsAvailable(validRecords, 1)
	if err != nil {
		t.Errorf("Expected no error for valid records, got: %v", err)
	}
}

// Test ValidateRecordsAvailable with different start indices
func TestETCCSVParser_ValidateRecordsAvailable_StartIndices(t *testing.T) {
	p := parser.NewETCCSVParser()

	records := [][]string{
		{"header"},
		{"data1"},
		{"data2"},
		{"data3"},
	}

	// Test various start indices
	testCases := []struct {
		startIndex  int
		shouldError bool
	}{
		{0, false}, // Valid: 4 records, start at 0
		{1, false}, // Valid: 4 records, start at 1
		{2, false}, // Valid: 4 records, start at 2
		{3, false}, // Valid: 4 records, start at 3
		{4, true},  // Invalid: 4 records, start at 4 (equals length)
		{5, true},  // Invalid: 4 records, start at 5 (greater than length)
		{10, true}, // Invalid: 4 records, start at 10 (much greater than length)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("startIndex_%d", tc.startIndex), func(t *testing.T) {
			err := p.ValidateRecordsAvailable(records, tc.startIndex)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for startIndex %d, got nil", tc.startIndex)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for startIndex %d, got: %v", tc.startIndex, err)
				}
			}
		})
	}
}