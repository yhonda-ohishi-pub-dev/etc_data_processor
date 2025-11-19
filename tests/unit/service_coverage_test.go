package unit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/proto"
	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/handler"
	"google.golang.org/grpc/codes"
)

// mockDBClient is a mock implementation of DBClient interface
type mockDBClient struct {
	saveFunc  func(data interface{}) error
	savedData []interface{}
}

func (m *mockDBClient) SaveETCData(data interface{}) error {
	m.savedData = append(m.savedData, data)
	if m.saveFunc != nil {
		return m.saveFunc(data)
	}
	return nil
}

// Test ProcessCSVFile edge cases for 100% coverage
func TestProcessCSVFile_Coverage(t *testing.T) {
	// Create a temporary test CSV file
	tmpDir := t.TempDir()
	validCSVPath := filepath.Join(tmpDir, "valid.csv")
	invalidCSVPath := filepath.Join(tmpDir, "invalid.csv")

	// Create valid CSV file
	validCSV := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト`

	if err := os.WriteFile(validCSVPath, []byte(validCSV), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid CSV file
	invalidCSV := `invalid data`
	if err := os.WriteFile(invalidCSVPath, []byte(invalidCSV), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		req      *pb.ProcessCSVFileRequest
		dbClient handler.DBClient
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "successful file processing",
			req: &pb.ProcessCSVFileRequest{
				CsvFilePath:    validCSVPath,
				AccountId:      "test-account",
				SkipDuplicates: false,
			},
			dbClient: &mockDBClient{},
			wantErr:  false,
		},
		{
			name: "file parse error",
			req: &pb.ProcessCSVFileRequest{
				CsvFilePath:    invalidCSVPath,
				AccountId:      "test-account",
				SkipDuplicates: false,
			},
			dbClient: &mockDBClient{},
			wantErr:  false, // Returns response with error details
		},
		{
			name: "db save error",
			req: &pb.ProcessCSVFileRequest{
				CsvFilePath:    validCSVPath,
				AccountId:      "test-account",
				SkipDuplicates: false,
			},
			dbClient: &mockDBClient{
				saveFunc: func(data interface{}) error {
					return errors.New("db error")
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := handler.NewDataProcessorService(tt.dbClient)
			resp, err := service.ProcessCSVFile(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Errorf("Response is nil")
				}
			}
		})
	}
}

// Test processRecords with context cancellation
func TestProcessRecords_ContextCancellation(t *testing.T) {
	mockDB := &mockDBClient{}
	service := handler.NewDataProcessorService(mockDB)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &pb.ProcessCSVDataRequest{
		CsvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト
25/09/02,08:00,25/09/02,09:00,横浜,名古屋,3000,-500,2500,2,1234,********12345678,テスト`,
		AccountId:      "test-account",
		SkipDuplicates: false,
	}

	resp, err := service.ProcessCSVData(ctx, req)

	// Should still return without error but with processing interrupted
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	// Check that errors mention cancellation
	hasCancel := false
	for _, e := range resp.Errors {
		if e == "Processing cancelled at record 0" {
			hasCancel = true
			break
		}
	}

	if !hasCancel {
		t.Errorf("Expected cancellation error in response")
	}
}

// Test ValidateCSVData with duplicates
func TestValidateCSVData_Duplicates(t *testing.T) {
	mockDB := &mockDBClient{}
	service := handler.NewDataProcessorService(mockDB)

	// CSV with duplicate records
	req := &pb.ValidateCSVDataRequest{
		CsvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト`,
		AccountId: "test-account",
	}

	resp, err := service.ValidateCSVData(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.DuplicateCount != 2 {
		t.Errorf("Expected 2 duplicates, got %d", resp.DuplicateCount)
	}

	if resp.TotalRecords != 3 {
		t.Errorf("Expected 3 total records, got %d", resp.TotalRecords)
	}
}

// Test ValidateCSVData with validation errors
func TestValidateCSVData_ValidationErrors(t *testing.T) {
	mockDB := &mockDBClient{}
	service := handler.NewDataProcessorService(mockDB)

	// CSV with records that will fail validation (no card number)
	req := &pb.ValidateCSVDataRequest{
		CsvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,,テスト
invalid,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト`,
		AccountId: "test-account",
	}

	resp, err := service.ValidateCSVData(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.IsValid {
		t.Errorf("Expected validation to fail")
	}

	if len(resp.Errors) == 0 {
		t.Errorf("Expected validation errors")
	}
}

// Test error in SaveETCData during processRecords
func TestProcessRecords_SaveError(t *testing.T) {
	saveAttempts := 0
	mockDB := &mockDBClient{
		saveFunc: func(data interface{}) error {
			saveAttempts++
			if saveAttempts == 1 {
				// First save succeeds
				return nil
			}
			// Second save fails
			return errors.New("database save error")
		},
	}

	service := handler.NewDataProcessorService(mockDB)

	req := &pb.ProcessCSVDataRequest{
		CsvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト1
25/09/02,08:00,25/09/02,09:00,横浜,名古屋,3000,-500,2500,2,1234,********87654321,テスト2`,
		AccountId:      "test-account",
		SkipDuplicates: false,
	}

	resp, err := service.ProcessCSVData(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have 1 saved, 1 error
	if resp.Stats.SavedRecords != 1 {
		t.Errorf("Expected 1 saved record, got %d", resp.Stats.SavedRecords)
	}

	if resp.Stats.ErrorRecords != 1 {
		t.Errorf("Expected 1 error record, got %d", resp.Stats.ErrorRecords)
	}

	// Check error message
	if len(resp.Errors) == 0 {
		t.Errorf("Expected error messages")
	}
}

// Test record conversion error during processRecords
func TestProcessRecords_ConversionError(t *testing.T) {
	mockDB := &mockDBClient{}
	service := handler.NewDataProcessorService(mockDB)

	// CSV with invalid dates that will fail conversion
	req := &pb.ProcessCSVDataRequest{
		CsvData: `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
invalid1,08:00,invalid2,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト`,
		AccountId:      "test-account",
		SkipDuplicates: false,
	}

	resp, err := service.ProcessCSVData(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have conversion error
	if resp.Stats.ErrorRecords != 1 {
		t.Errorf("Expected 1 error record, got %d", resp.Stats.ErrorRecords)
	}

	if len(resp.Errors) == 0 {
		t.Errorf("Expected error messages for conversion failure")
	}
}

// Test SkipDuplicates with multiple duplicates
func TestProcessCSVData_SkipMultipleDuplicates(t *testing.T) {
	mockDB := &mockDBClient{}
	service := handler.NewDataProcessorService(mockDB)

	// Same record repeated 5 times
	csvData := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考`
	for i := 0; i < 5; i++ {
		csvData += "\n25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト"
	}

	req := &pb.ProcessCSVDataRequest{
		CsvData:        csvData,
		AccountId:      "test-account",
		SkipDuplicates: true,
	}

	resp, err := service.ProcessCSVData(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Stats.TotalRecords != 5 {
		t.Errorf("Expected 5 total records, got %d", resp.Stats.TotalRecords)
	}

	if resp.Stats.SavedRecords != 1 {
		t.Errorf("Expected 1 saved record, got %d", resp.Stats.SavedRecords)
	}

	if resp.Stats.SkippedRecords != 4 {
		t.Errorf("Expected 4 skipped records, got %d", resp.Stats.SkippedRecords)
	}
}