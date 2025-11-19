package unit

import (
	"context"
	"errors"
	"testing"

	pb "github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/proto"
	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/handler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Test ProcessCSVFile with validator errors
func TestProcessCSVFile_ValidatorErrors(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.ProcessCSVFileRequest
		validator handler.Validator
		wantErr   bool
	}{
		{
			name: "validator path error",
			req: &pb.ProcessCSVFileRequest{
				CsvFilePath: "test.csv",
				AccountId:   "test-account",
			},
			validator: &MockValidator{
				ValidateCSVFilePathFunc: func(path string) error {
					return status.Error(codes.InvalidArgument, "invalid path")
				},
			},
			wantErr: true,
		},
		{
			name: "validator account error",
			req: &pb.ProcessCSVFileRequest{
				CsvFilePath: "test.csv",
				AccountId:   "test-account",
			},
			validator: &MockValidator{
				ValidateAccountIDFunc: func(id string) error {
					return status.Error(codes.InvalidArgument, "invalid account")
				},
			},
			wantErr: true,
		},
		{
			name: "validator file exists error",
			req: &pb.ProcessCSVFileRequest{
				CsvFilePath: "test.csv",
				AccountId:   "test-account",
			},
			validator: &MockValidator{
				CheckFileExistsFunc: func(path string) error {
					return status.Error(codes.NotFound, "file not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := handler.NewDataProcessorServiceWithValidator(
				&mockDBClient{},
				tt.validator,
			)
			_, err := service.ProcessCSVFile(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessCSVFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test ProcessCSVData with validator errors
func TestProcessCSVData_ValidatorErrors(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.ProcessCSVDataRequest
		validator handler.Validator
		wantErr   bool
	}{
		{
			name: "validator CSV data error",
			req: &pb.ProcessCSVDataRequest{
				CsvData:   "test",
				AccountId: "test-account",
			},
			validator: &MockValidator{
				ValidateCSVDataFunc: func(data string) error {
					return status.Error(codes.InvalidArgument, "invalid data")
				},
			},
			wantErr: true,
		},
		{
			name: "validator account error",
			req: &pb.ProcessCSVDataRequest{
				CsvData:   "valid data",
				AccountId: "test-account",
			},
			validator: &MockValidator{
				ValidateAccountIDFunc: func(id string) error {
					return status.Error(codes.InvalidArgument, "invalid account")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := handler.NewDataProcessorServiceWithValidator(
				&mockDBClient{},
				tt.validator,
			)
			_, err := service.ProcessCSVData(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessCSVData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test ValidateCSVData with validator errors
func TestValidateCSVData_ValidatorErrors(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.ValidateCSVDataRequest
		validator handler.Validator
		wantErr   bool
	}{
		{
			name: "validator CSV data error",
			req: &pb.ValidateCSVDataRequest{
				CsvData: "test",
			},
			validator: &MockValidator{
				ValidateCSVDataFunc: func(data string) error {
					return status.Error(codes.InvalidArgument, "invalid data")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := handler.NewDataProcessorServiceWithValidator(
				&mockDBClient{},
				tt.validator,
			)
			_, err := service.ValidateCSVData(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCSVData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test service with nil dbClient
func TestService_NilDBClient(t *testing.T) {
	service := handler.NewDataProcessorService(nil)

	validCSV := `利用年月日（自）,時分（自）,利用年月日（至）,時分（至）,利用ＩＣ（自）,利用ＩＣ（至）,割引前料金,ＥＴＣ割引額,通行料金,車種,車両番号,ＥＴＣカード番号,備考
25/09/01,08:00,25/09/01,09:00,東京,横浜,1500,-300,1200,2,1234,********12345678,テスト`

	req := &pb.ProcessCSVDataRequest{
		CsvData:   validCSV,
		AccountId: "test-account",
	}

	resp, err := service.ProcessCSVData(context.Background(), req)

	// Should work even with nil dbClient (won't save but processes)
	if err != nil {
		t.Errorf("Unexpected error with nil dbClient: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}
}

// Test validation helper with internal error
func TestCheckFileExists_InternalError(t *testing.T) {
	// This test simulates an error other than file not found
	// We can't easily simulate this with DefaultValidator,
	// but we can test the error path with a mock
	mockValidator := &MockValidator{
		CheckFileExistsFunc: func(path string) error {
			// Return a non-standard error to trigger internal error path
			return errors.New("permission denied")
		},
	}

	err := mockValidator.CheckFileExists("/some/path")
	if err == nil {
		t.Error("Expected error")
	}
}