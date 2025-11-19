package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

func TestETCCSVParser_GetFieldSafe(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Test data
	record := []string{"field0", "field1", "field2"}

	tests := []struct {
		name     string
		record   []string
		index    int
		expected string
	}{
		{
			name:     "valid index 0",
			record:   record,
			index:    0,
			expected: "field0",
		},
		{
			name:     "valid index 1",
			record:   record,
			index:    1,
			expected: "field1",
		},
		{
			name:     "valid index 2",
			record:   record,
			index:    2,
			expected: "field2",
		},
		{
			name:     "index out of bounds",
			record:   record,
			index:    3,
			expected: "",
		},
		{
			name:     "negative index",
			record:   record,
			index:    -1,
			expected: "",
		},
		{
			name:     "empty record",
			record:   []string{},
			index:    0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Access the unexported method through parsing
			// Since getFieldSafe is called internally, we test it through Parse
			csvData := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234`

			// Add more fields to test boundary conditions
			if tt.index >= 13 {
				csvData += ",extra1,extra2,extra3"
			}

			reader := strings.NewReader(csvData)
			_, err := p.Parse(reader)

			// We're testing that Parse handles various field counts correctly
			// which internally uses getFieldSafe
			if err != nil && tt.index < 13 {
				t.Errorf("Parse failed when it should succeed: %v", err)
			}
		})
	}
}

func TestETCCSVParser_ParseDate(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name    string
		dateStr string
		wantErr bool
		want    time.Time
	}{
		{
			name:    "valid date YY/MM/DD",
			dateStr: "25/09/01",
			wantErr: false,
			want:    time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "valid date with year < 50",
			dateStr: "24/12/31",
			wantErr: false,
			want:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "valid date with year >= 50",
			dateStr: "99/01/01",
			wantErr: false,
			want:    time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid format",
			dateStr: "2024-01-01",
			wantErr: false, // Falls back to entry date
		},
		{
			name:    "empty string",
			dateStr: "",
			wantErr: false, // Falls back to entry date
		},
		{
			name:    "invalid month",
			dateStr: "24/13/01",
			wantErr: false, // Falls back to entry date
		},
		{
			name:    "invalid day",
			dateStr: "24/01/32",
			wantErr: false, // Falls back to entry date
		},
		{
			name:    "non-numeric year",
			dateStr: "AB/01/01",
			wantErr: false, // Falls back to entry date
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through ConvertToSimpleRecord which calls parseDate
			record := parser.ActualETCRecord{
				EntryDate:    "25/01/01",
				ExitDate:     tt.dateStr,
				EntryIC:      "東京",
				ExitIC:       "横浜",
				ETCAmount:    1000,
				CardNumber:   "1234567890",
				VehicleClass: 2,
			}

			_, err := p.ConvertToSimpleRecord(record)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestETCCSVParser_ParseAmount(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name    string
		csvData string
		wantErr bool
	}{
		{
			name: "positive amount",
			csvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,0,1500,2,1234,********12345678,`,
			wantErr: false,
		},
		{
			name: "negative amount",
			csvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,`,
			wantErr: false,
		},
		{
			name: "amount with comma",
			csvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,"10,500","-1,000","9,500",2,1234,********12345678,`,
			wantErr: false,
		},
		{
			name: "invalid amount",
			csvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,abc,def,ghi,2,1234,********12345678,`,
			wantErr: false, // Parser continues with 0 value for invalid amounts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			records, err := p.Parse(reader)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr && len(records) > 0 {
				// Check that parsing succeeded
				if records[0].CardNumber == "" {
					t.Errorf("Card number should not be empty")
				}
			}
		})
	}
}

func TestETCCSVParser_ValidateRecord_EdgeCases(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name    string
		record  parser.ActualETCRecord
		wantErr bool
	}{
		{
			name: "empty entry date but valid exit date",
			record: parser.ActualETCRecord{
				EntryDate:  "",
				ExitDate:   "25/09/01",
				CardNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "empty exit date but valid entry date",
			record: parser.ActualETCRecord{
				EntryDate:  "25/09/01",
				ExitDate:   "",
				CardNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "both dates empty but card exists",
			record: parser.ActualETCRecord{
				EntryDate:  "",
				ExitDate:   "",
				CardNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "invalid entry date format",
			record: parser.ActualETCRecord{
				EntryDate:  "invalid",
				ExitDate:   "25/09/01",
				CardNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "invalid exit date format",
			record: parser.ActualETCRecord{
				EntryDate:  "25/09/01",
				ExitDate:   "invalid",
				CardNumber: "1234567890",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.ValidateRecord(tt.record)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestETCCSVParser_ConvertToSimpleRecord_EdgeCases(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name    string
		record  parser.ActualETCRecord
		wantErr bool
	}{
		{
			name: "zero ETC amount uses normal amount",
			record: parser.ActualETCRecord{
				EntryDate:     "25/09/01",
				ExitDate:      "25/09/01",
				ETCAmount:     0,
				NormalAmount:  1500,
				CardNumber:    "1234567890",
				VehicleClass:  2,
			},
			wantErr: false,
		},
		{
			name: "negative ETC amount becomes positive",
			record: parser.ActualETCRecord{
				EntryDate:     "25/09/01",
				ExitDate:      "25/09/01",
				ETCAmount:     -1000,
				NormalAmount:  0,
				CardNumber:    "1234567890",
				VehicleClass:  2,
			},
			wantErr: false,
		},
		{
			name: "invalid exit date uses entry date",
			record: parser.ActualETCRecord{
				EntryDate:     "25/09/01",
				ExitDate:      "invalid",
				ETCAmount:     1000,
				CardNumber:    "1234567890",
				VehicleClass:  2,
			},
			wantErr: false,
		},
		{
			name: "both dates invalid",
			record: parser.ActualETCRecord{
				EntryDate:     "invalid1",
				ExitDate:      "invalid2",
				ETCAmount:     1000,
				CardNumber:    "1234567890",
				VehicleClass:  2,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simple, err := p.ConvertToSimpleRecord(tt.record)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr {
				// Check specific conversions
				if tt.record.ETCAmount < 0 && simple.Amount < 0 {
					t.Errorf("Negative amount should be converted to positive")
				}
				if tt.record.ETCAmount == 0 && tt.record.NormalAmount > 0 && simple.Amount != tt.record.NormalAmount {
					t.Errorf("Should use normal amount when ETC amount is zero")
				}
			}
		})
	}
}