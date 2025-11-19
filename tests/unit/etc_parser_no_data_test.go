package unit

import (
	"io"
	"strings"
	"testing"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

// Test ETCCSVParser "no data records found" error path
func TestETCCSVParser_NoDataRecordsFound(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Create CSV with only header, no data records
	csvData := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,通行料金,車種,ＥＴＣカード番号`

	reader := strings.NewReader(csvData)
	_, err := p.Parse(reader)

	if err == nil {
		t.Error("Expected 'no data records found' error, got nil")
	}

	if !strings.Contains(err.Error(), "no data records found") {
		t.Errorf("Expected 'no data records found', got: %v", err)
	}
}

// Test ETCCSVParser with empty CSV
func TestETCCSVParser_EmptyCSV(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Completely empty CSV
	reader := strings.NewReader("")
	_, err := p.Parse(reader)

	if err == nil {
		t.Error("Expected error for empty CSV, got nil")
	}

	// Should get "CSV file is empty" error before "no data records found"
	if !strings.Contains(err.Error(), "CSV file is empty") {
		t.Errorf("Expected 'CSV file is empty', got: %v", err)
	}
}


// Test with mock reader that returns specific CSV content
type mockReaderNoData struct {
	content string
	pos     int
}

func (r *mockReaderNoData) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.content) {
		return 0, io.EOF
	}

	n = copy(p, r.content[r.pos:])
	r.pos += n
	return n, nil
}

// Test ETCCSVParser "no data records found" with mock reader
func TestETCCSVParser_MockNoDataRecords(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Mock CSV with header only
	mockReader := &mockReaderNoData{
		content: "利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,通行料金,車種,ＥＴＣカード番号\n",
		pos:     0,
	}

	_, err := p.Parse(mockReader)

	if err == nil {
		t.Error("Expected 'no data records found' error with mock reader, got nil")
	}

	if !strings.Contains(err.Error(), "no data records found") {
		t.Errorf("Expected 'no data records found' with mock reader, got: %v", err)
	}
}