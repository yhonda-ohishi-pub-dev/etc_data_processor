package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_data_processor/src/pkg/parser"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []parser.ETCRecord
		wantErr bool
	}{
		{
			name: "valid CSV data",
			input: `日付,入口IC,出口IC,路線,車種,金額,カード番号
2024-01-01,東京IC,横浜IC,東名高速,1,1500,1234567890
2024-01-02,名古屋IC,大阪IC,名神高速,2,3000,0987654321`,
			want: []parser.ETCRecord{
				{
					Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EntryIC:     "東京IC",
					ExitIC:      "横浜IC",
					Route:       "東名高速",
					VehicleType: 1,
					Amount:      1500,
					CardNumber:  "1234567890",
				},
				{
					Date:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					EntryIC:     "名古屋IC",
					ExitIC:      "大阪IC",
					Route:       "名神高速",
					VehicleType: 2,
					Amount:      3000,
					CardNumber:  "0987654321",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty CSV",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid amount",
			input: `日付,入口IC,出口IC,路線,車種,金額,カード番号
2024-01-01,東京IC,横浜IC,東名高速,1,invalid,1234567890`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid date format",
			input: `日付,入口IC,出口IC,路線,車種,金額,カード番号
invalid-date,東京IC,横浜IC,東名高速,1,1500,1234567890`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing fields",
			input: `日付,入口IC,出口IC,路線,車種,金額,カード番号
2024-01-01,東京IC,横浜IC`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewCSVParser()
			reader := strings.NewReader(tt.input)
			got, err := p.Parse(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !equalRecords(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCSVFile(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		wantErr  bool
	}{
		{
			name:     "non-existent file",
			filepath: "/path/to/nonexistent.csv",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewCSVParser()
			_, err := p.ParseFile(tt.filepath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRecord(t *testing.T) {
	tests := []struct {
		name    string
		record  parser.ETCRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: parser.ETCRecord{
				Date:        time.Now(),
				EntryIC:     "東京IC",
				ExitIC:      "横浜IC",
				Route:       "東名高速",
				VehicleType: 1,
				Amount:      1500,
				CardNumber:  "1234567890",
			},
			wantErr: false,
		},
		{
			name: "negative amount",
			record: parser.ETCRecord{
				Date:        time.Now(),
				EntryIC:     "東京IC",
				ExitIC:      "横浜IC",
				Route:       "東名高速",
				VehicleType: 1,
				Amount:      -100,
				CardNumber:  "1234567890",
			},
			wantErr: true,
		},
		{
			name: "empty entry IC",
			record: parser.ETCRecord{
				Date:        time.Now(),
				EntryIC:     "",
				ExitIC:      "横浜IC",
				Route:       "東名高速",
				VehicleType: 1,
				Amount:      1500,
				CardNumber:  "1234567890",
			},
			wantErr: true,
		},
		{
			name: "empty card number",
			record: parser.ETCRecord{
				Date:        time.Now(),
				EntryIC:     "東京IC",
				ExitIC:      "横浜IC",
				Route:       "東名高速",
				VehicleType: 1,
				Amount:      1500,
				CardNumber:  "",
			},
			wantErr: true,
		},
		{
			name: "future date",
			record: parser.ETCRecord{
				Date:        time.Now().AddDate(0, 0, 1),
				EntryIC:     "東京IC",
				ExitIC:      "横浜IC",
				Route:       "東名高速",
				VehicleType: 1,
				Amount:      1500,
				CardNumber:  "1234567890",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewCSVParser()
			err := p.ValidateRecord(tt.record)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func equalRecords(a, b []parser.ETCRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Date.Equal(b[i].Date) ||
			a[i].EntryIC != b[i].EntryIC ||
			a[i].ExitIC != b[i].ExitIC ||
			a[i].Route != b[i].Route ||
			a[i].VehicleType != b[i].VehicleType ||
			a[i].Amount != b[i].Amount ||
			a[i].CardNumber != b[i].CardNumber {
			return false
		}
	}
	return true
}