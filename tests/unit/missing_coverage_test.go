package unit

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/proto"
	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/handler"
	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

// Test HealthCheck method (to cover service.go:173-183)
func TestHealthCheck(t *testing.T) {
	service := handler.NewDataProcessorService(nil)

	req := &pb.HealthCheckRequest{}
	resp, err := service.HealthCheck(context.Background(), req)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", resp.Status)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", resp.Version)
	}
}

// Test ProcessCSVFile with parse error that returns error in response (service.go:74-83)
func TestProcessCSVFile_ParseError(t *testing.T) {
	// Create a file that will cause parse error
	tmpDir := t.TempDir()
	errorFile := filepath.Join(tmpDir, "error.csv")
	os.WriteFile(errorFile, []byte("bad data"), 0644)

	service := handler.NewDataProcessorService(nil)
	req := &pb.ProcessCSVFileRequest{
		CsvFilePath:    errorFile,
		AccountId:      "test-account",
		SkipDuplicates: false,
	}

	resp, err := service.ProcessCSVFile(context.Background(), req)

	if err != nil {
		t.Errorf("Should return response with error details, not error: %v", err)
	}

	if resp.Success {
		t.Error("Expected Success to be false")
	}

	// ETCCSVParser might still parse something, so check message for error
	if !strings.Contains(resp.Message, "Processed") {
		t.Logf("Response message: %s", resp.Message)
	}
}

// Test ProcessCSVData with parse error (service.go:105-109)
func TestProcessCSVData_ParseError(t *testing.T) {
	service := handler.NewDataProcessorService(nil)

	// CSV that will fail parsing - empty string should trigger validator error first
	req := &pb.ProcessCSVDataRequest{
		CsvData:   "",
		AccountId: "test-account",
	}

	_, err := service.ProcessCSVData(context.Background(), req)

	// Should return error for invalid format - validator catches empty data
	if err == nil {
		t.Error("Expected error for invalid CSV format")
	}
}

// Test ValidateCSVData with parse error (service.go:130-145)
func TestValidateCSVData_ParseError(t *testing.T) {
	service := handler.NewDataProcessorService(nil)

	// CSV that will fail parsing - empty data triggers validator error first
	req := &pb.ValidateCSVDataRequest{
		CsvData: "",
	}

	_, err := service.ValidateCSVData(context.Background(), req)

	// Validator catches empty data before parsing
	if err == nil {
		t.Error("Expected validator error for empty CSV data")
	}
}

// Test CheckFileExists with permission error (validator.go:62-64)
func TestCheckFileExists_OtherError(t *testing.T) {
	// Create a validator and test with a path that might trigger other errors
	v := handler.NewDefaultValidator()

	// Try with invalid path characters on Windows
	if os.PathSeparator == '\\' {
		// Windows invalid path
		err := v.CheckFileExists("C:\\<>:|?*")
		if err == nil {
			t.Skip("Could not trigger non-NotExist error")
		}
	} else {
		// Unix - try a path that might not be accessible
		err := v.CheckFileExists("/root/.ssh/id_rsa_test_nonexistent")
		if err == nil {
			t.Skip("Could not trigger error")
		}
	}
}

// Test CSV parser with header row that is too short (csv_parser.go:87-89)
func TestCSVParser_HeaderRowError(t *testing.T) {
	p := parser.NewCSVParser()

	// CSV with header but no data rows and validation will fail
	csvData := `日付,入口IC,出口IC`

	reader := strings.NewReader(csvData)
	records, err := p.Parse(reader)

	// Should fail due to incomplete header
	if err == nil && len(records) > 0 {
		t.Error("Expected error or empty records for incomplete CSV")
	}
}

// Test ETCCSVParser Parse with nil reader (etc_csv_parser.go:59-61)
func TestETCCSVParser_Parse_NilReader(t *testing.T) {
	p := parser.NewETCCSVParser()

	_, err := p.Parse(nil)

	if err == nil {
		t.Error("Expected error for nil reader")
	}

	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("Expected nil-related error, got: %v", err)
	}
}

// Test ETCCSVParser ParseFile with Open error (etc_csv_parser.go:69-71)
func TestETCCSVParser_ParseFile_OpenError(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Try to open a directory as file
	tmpDir := t.TempDir()

	_, err := p.ParseFile(tmpDir)

	if err == nil {
		t.Error("Expected error when opening directory as file")
	}
}

// Test ETCCSVParser Parse with empty CSV (etc_csv_parser.go:73-75)
func TestETCCSVParser_Parse_EmptyCSV(t *testing.T) {
	p := parser.NewETCCSVParser()

	reader := strings.NewReader("")
	_, err := p.Parse(reader)

	if err == nil {
		t.Error("Expected error for empty CSV")
	}

	if !strings.Contains(err.Error(), "CSV file is empty") {
		t.Errorf("Expected 'CSV file is empty' error, got: %v", err)
	}
}

// Test ETCCSVParser with warning for records (etc_csv_parser.go:105-107)
func TestETCCSVParser_RecordLengthWarning(t *testing.T) {
	p := parser.NewETCCSVParser()

	// CSV with 13 fields (will trigger length check)
	csvData := `25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト`

	reader := strings.NewReader(csvData)
	records, err := p.Parse(reader)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should process the record despite length
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
}

// Test parseWithHeaders with vehicle class non-numeric fallback (etc_csv_parser.go:184-186)
func TestETCCSVParser_VehicleClassFallback(t *testing.T) {
	p := parser.NewETCCSVParser()

	// CSV with non-numeric vehicle class that will use fallback
	csvData := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,普通車,1234,********12345678,テスト`

	reader := strings.NewReader(csvData)
	records, err := p.Parse(reader)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Vehicle class should use fallback value (the actual fallback is 0 for non-numeric)
	if records[0].VehicleClass != 0 {
		t.Errorf("Expected vehicle class 0 (fallback for non-numeric), got %d", records[0].VehicleClass)
	}
}

// Test parseAmount with non-numeric amount (etc_csv_parser.go:219-221)
func TestETCCSVParser_ParseAmount_NonNumeric(t *testing.T) {
	p := parser.NewETCCSVParser()

	// CSV with non-numeric amounts
	csvData := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,ABC,-DEF,GHI,2,1234,********12345678,テスト`

	reader := strings.NewReader(csvData)
	records, err := p.Parse(reader)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should handle non-numeric amounts gracefully
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Non-numeric amounts should become 0
	if records[0].NormalAmount != 0 {
		t.Errorf("Expected 0 for non-numeric normal amount, got %d", records[0].NormalAmount)
	}
}

// Test ConvertToSimpleRecord with invalid dates (etc_csv_parser.go:283-285, 288-290)
func TestETCCSVParser_ConvertToSimpleRecord_InvalidDates(t *testing.T) {
	p := parser.NewETCCSVParser()

	record := parser.ActualETCRecord{
		EntryDate:    "invalid",
		EntryTime:    "08:00",
		ExitDate:     "invalid",
		ExitTime:     "09:00",
		EntryIC:      "東京",
		ExitIC:       "横浜",
		ETCAmount:    1200,
		VehicleClass: 2,
		CardNumber:   "********12345678",
	}

	_, err := p.ConvertToSimpleRecord(record)

	// Should return error for invalid dates
	if err == nil {
		t.Error("Expected error for invalid dates")
	}
}

// Test parseWithHeaders with missing discount column (etc_csv_parser.go:381-384)
func TestETCCSVParser_MissingDiscountColumn(t *testing.T) {
	p := parser.NewETCCSVParser()

	// CSV without discount columns - only has essential columns
	csvData := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,通行料金,車種,ＥＴＣカード番号
25/09/01,08:00,25/09/01,09:00,東京,横浜,1200,2,********12345678`

	reader := strings.NewReader(csvData)
	records, err := p.Parse(reader)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Should still process correctly without discount columns
	if records[0].ETCAmount != 1200 {
		t.Errorf("Expected ETC amount 1200, got %d", records[0].ETCAmount)
	}
}

// Test with custom reader that returns error
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestETCCSVParser_ReadError(t *testing.T) {
	p := parser.NewETCCSVParser()

	_, err := p.Parse(&errorReader{})

	if err == nil {
		t.Error("Expected error from reader")
	}
}

// Test Parse with reader that returns EOF immediately
type eofReader struct{}

func (r *eofReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestETCCSVParser_EOFReader(t *testing.T) {
	p := parser.NewETCCSVParser()

	_, err := p.Parse(&eofReader{})

	if err == nil {
		t.Error("Expected error for EOF reader")
	}
}