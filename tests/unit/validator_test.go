package unit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/handler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockValidator is a mock implementation of the Validator interface
type MockValidator struct {
	ValidateCSVFilePathFunc func(path string) error
	ValidateAccountIDFunc   func(accountID string) error
	ValidateCSVDataFunc     func(data string) error
	CheckFileExistsFunc     func(path string) error
}

func (m *MockValidator) ValidateCSVFilePath(path string) error {
	if m.ValidateCSVFilePathFunc != nil {
		return m.ValidateCSVFilePathFunc(path)
	}
	return nil
}

func (m *MockValidator) ValidateAccountID(accountID string) error {
	if m.ValidateAccountIDFunc != nil {
		return m.ValidateAccountIDFunc(accountID)
	}
	return nil
}

func (m *MockValidator) ValidateCSVData(data string) error {
	if m.ValidateCSVDataFunc != nil {
		return m.ValidateCSVDataFunc(data)
	}
	return nil
}

func (m *MockValidator) CheckFileExists(path string) error {
	if m.CheckFileExistsFunc != nil {
		return m.CheckFileExistsFunc(path)
	}
	return nil
}

// Test DefaultValidator
func TestDefaultValidator(t *testing.T) {
	v := handler.NewDefaultValidator()

	t.Run("ValidateCSVFilePath", func(t *testing.T) {
		// Empty path
		if err := v.ValidateCSVFilePath(""); err == nil {
			t.Error("Expected error for empty path")
		}

		// Valid path
		if err := v.ValidateCSVFilePath("/valid/path"); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("ValidateAccountID", func(t *testing.T) {
		// Empty ID (now optional)
		if err := v.ValidateAccountID(""); err != nil {
			t.Errorf("Expected no error for empty account ID (optional): %v", err)
		}

		// Too short ID
		if err := v.ValidateAccountID("ab"); err == nil {
			t.Error("Expected error for short account ID")
		}

		// Valid ID
		if err := v.ValidateAccountID("valid-account"); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("ValidateCSVData", func(t *testing.T) {
		// Empty data
		if err := v.ValidateCSVData(""); err == nil {
			t.Error("Expected error for empty data")
		}

		// Too short data
		if err := v.ValidateCSVData("short"); err == nil {
			t.Error("Expected error for short data")
		}

		// Valid data
		if err := v.ValidateCSVData("valid csv data with sufficient length"); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("CheckFileExists", func(t *testing.T) {
		// Non-existent file
		if err := v.CheckFileExists("/nonexistent/file"); err == nil {
			t.Error("Expected error for non-existent file")
		}

		// Existing file
		tmpDir := t.TempDir()
		existingFile := filepath.Join(tmpDir, "test.csv")
		if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := v.CheckFileExists(existingFile); err != nil {
			t.Errorf("Unexpected error for existing file: %v", err)
		}
	})
}

// mockFileReq implements the FileRequest interface for testing
type mockFileReq struct {
	csvFilePath string
	accountId   string
}

func (m *mockFileReq) GetCsvFilePath() string {
	return m.csvFilePath
}

func (m *mockFileReq) GetAccountId() string {
	return m.accountId
}

// Test ValidateProcessCSVFileRequest
func TestValidateProcessCSVFileRequest(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "valid.csv")
	os.WriteFile(validFile, []byte("test"), 0644)

	tests := []struct {
		name      string
		req       interface{}
		validator handler.Validator
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:      "nil request",
			req:       nil,
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name:      "invalid request type",
			req:       "invalid",
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty file path",
			req: &mockFileReq{
				csvFilePath: "",
				accountId:   "test-account",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty account ID (now optional)",
			req: &mockFileReq{
				csvFilePath: validFile,
				accountId:   "",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   false,
		},
		{
			name: "file not found",
			req: &mockFileReq{
				csvFilePath: "/nonexistent/file",
				accountId:   "test-account",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.NotFound,
		},
		{
			name: "valid request",
			req: &mockFileReq{
				csvFilePath: validFile,
				accountId:   "test-account",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   false,
		},
		{
			name: "mock validator error",
			req: &mockFileReq{
				csvFilePath: validFile,
				accountId:   "test-account",
			},
			validator: &MockValidator{
				ValidateCSVFilePathFunc: func(path string) error {
					return errors.New("mock error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateProcessCSVFileRequest(tt.req, tt.validator)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProcessCSVFileRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantCode != 0 {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error")
				} else if st.Code() != tt.wantCode {
					t.Errorf("Expected code %v, got %v", tt.wantCode, st.Code())
				}
			}
		})
	}
}

// mockDataReq implements the DataRequest interface for testing
type mockDataReq struct {
	csvData   string
	accountId string
}

func (m *mockDataReq) GetCsvData() string {
	return m.csvData
}

func (m *mockDataReq) GetAccountId() string {
	return m.accountId
}

// Test ValidateProcessCSVDataRequest
func TestValidateProcessCSVDataRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       interface{}
		validator handler.Validator
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:      "nil request",
			req:       nil,
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name:      "invalid request type",
			req:       123,
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty csv data",
			req: &mockDataReq{
				csvData:   "",
				accountId: "test-account",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty account ID (now optional)",
			req: &mockDataReq{
				csvData:   "valid csv data with sufficient length",
				accountId: "",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   false,
		},
		{
			name: "valid request",
			req: &mockDataReq{
				csvData:   "valid csv data with sufficient length",
				accountId: "test-account",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   false,
		},
		{
			name: "mock validator CSV data error",
			req: &mockDataReq{
				csvData:   "valid csv data",
				accountId: "test-account",
			},
			validator: &MockValidator{
				ValidateCSVDataFunc: func(data string) error {
					return status.Error(codes.InvalidArgument, "mock csv error")
				},
			},
			wantErr: true,
		},
		{
			name: "mock validator account ID error",
			req: &mockDataReq{
				csvData:   "valid csv data",
				accountId: "test-account",
			},
			validator: &MockValidator{
				ValidateAccountIDFunc: func(id string) error {
					return status.Error(codes.InvalidArgument, "mock account error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateProcessCSVDataRequest(tt.req, tt.validator)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProcessCSVDataRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantCode != 0 {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error")
				} else if st.Code() != tt.wantCode {
					t.Errorf("Expected code %v, got %v", tt.wantCode, st.Code())
				}
			}
		})
	}
}

// mockValidateReq implements the ValidateRequest interface for testing
type mockValidateReq struct {
	csvData   string
	accountId string
}

func (m *mockValidateReq) GetCsvData() string {
	return m.csvData
}

func (m *mockValidateReq) GetAccountId() string {
	return m.accountId
}

// Test ValidateValidateCSVDataRequest
func TestValidateValidateCSVDataRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       interface{}
		validator handler.Validator
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:      "nil request",
			req:       nil,
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name:      "invalid request type",
			req:       false,
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "empty csv data",
			req: &mockValidateReq{
				csvData: "",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
		},
		{
			name: "valid request",
			req: &mockValidateReq{
				csvData: "valid csv data with sufficient length",
			},
			validator: handler.NewDefaultValidator(),
			wantErr:   false,
		},
		{
			name: "mock validator error",
			req: &mockValidateReq{
				csvData: "valid csv data",
			},
			validator: &MockValidator{
				ValidateCSVDataFunc: func(data string) error {
					return status.Error(codes.InvalidArgument, "mock error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateValidateCSVDataRequest(tt.req, tt.validator)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValidateCSVDataRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantCode != 0 {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error")
				} else if st.Code() != tt.wantCode {
					t.Errorf("Expected code %v, got %v", tt.wantCode, st.Code())
				}
			}
		})
	}
}

// Test CreateDuplicateKey
func TestCreateDuplicateKey(t *testing.T) {
	key := handler.CreateDuplicateKey(
		"2024-01-01",
		"08:00",
		"2024-01-01",
		"09:00",
		1500,
		"1234567890",
	)

	expected := "2024-01-01_08:00_2024-01-01_09:00_1500_1234567890"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}