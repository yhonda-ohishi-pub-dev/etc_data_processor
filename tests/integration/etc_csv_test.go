package integration

import (
	"path/filepath"
	"testing"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

func TestParseActualETCFiles(t *testing.T) {
	testDir := "../file"

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "Parse actual ETC file 202509282006.csv",
			filename: "202509282006.csv",
			wantErr:  false,
		},
		{
			name:     "Parse actual ETC file 202509282007.csv",
			filename: "202509282007.csv",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewETCCSVParser()
			filepath := filepath.Join(testDir, tt.filename)

			records, err := p.ParseFile(filepath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(records) > 0 {
				// Display first few records to check encoding
				t.Logf("Successfully parsed %d records from %s", len(records), tt.filename)

				// Show first 3 records for verification
				for i := 0; i < 3 && i < len(records); i++ {
					record := records[i]
					t.Logf("Record %d:", i+1)
					t.Logf("  Date: %s %s -> %s %s", record.EntryDate, record.EntryTime, record.ExitDate, record.ExitTime)
					t.Logf("  IC: %s -> %s", record.EntryIC, record.ExitIC)
					t.Logf("  Amount: ETC=%d, Normal=%d", record.ETCAmount, record.NormalAmount)
					t.Logf("  Card: %s", record.CardNumber)

					// Check for mojibake (garbled characters)
					if containsMojibake(record.EntryIC) || containsMojibake(record.ExitIC) {
						t.Errorf("Mojibake detected in IC names")
					}
				}

				// Test conversion to simple format
				for i, record := range records {
					simple, err := p.ConvertToSimpleRecord(record)
					if err != nil {
						t.Logf("Warning: Failed to convert record %d: %v", i, err)
						continue
					}

					// Verify converted record
					if simple.Amount < 0 {
						t.Errorf("Negative amount in converted record: %d", simple.Amount)
					}

					if i == 0 {
						t.Logf("First converted record:")
						t.Logf("  Date: %v", simple.Date)
						t.Logf("  Route: %s -> %s", simple.EntryIC, simple.ExitIC)
						t.Logf("  Amount: %d", simple.Amount)
					}
				}
			}
		})
	}
}

// containsMojibake checks for common mojibake patterns
func containsMojibake(s string) bool {
	// Check for common mojibake patterns
	// Check for question marks at the beginning (common mojibake indicator)
	if len(s) > 0 && s[0] == '?' && s != "?" {
		return true
	}

	// Check for replacement characters
	mojibakePatterns := []string{
		"�",      // Replacement character
		"ï¿½",    // UTF-8 replacement character seen as Latin-1
		"â€",     // Common mojibake pattern
	}

	for _, pattern := range mojibakePatterns {
		if s == pattern {
			return true
		}
	}

	return false
}

func TestEncodingVerification(t *testing.T) {
	testDir := "../file"
	filename := "202509282007.csv"

	p := parser.NewETCCSVParser()
	filepath := filepath.Join(testDir, filename)

	records, err := p.ParseFile(filepath)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	if len(records) > 0 {
		// Check specific known values
		firstRecord := records[0]

		t.Logf("First record details:")
		t.Logf("  Entry IC: %s (len=%d)", firstRecord.EntryIC, len(firstRecord.EntryIC))
		t.Logf("  Exit IC: %s (len=%d)", firstRecord.ExitIC, len(firstRecord.ExitIC))

		// Check if Japanese characters are properly decoded
		for i, r := range firstRecord.EntryIC {
			if r > 0x7F { // Non-ASCII character
				t.Logf("  Non-ASCII character at position %d: %c (U+%04X)", i, r, r)
				break
			}
		}

		// Verify no question marks in IC names (common mojibake indicator)
		if len(firstRecord.EntryIC) > 0 && firstRecord.EntryIC[0] == '?' {
			t.Errorf("Possible encoding issue: Entry IC starts with '?'")
		}
		if len(firstRecord.ExitIC) > 0 && firstRecord.ExitIC[0] == '?' {
			t.Errorf("Possible encoding issue: Exit IC starts with '?'")
		}
	}
}