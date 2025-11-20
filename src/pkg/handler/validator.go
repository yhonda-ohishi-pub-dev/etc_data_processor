package handler

import (
	"fmt"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validator interface for request validation
type Validator interface {
	ValidateCSVFilePath(path string) error
	ValidateAccountID(accountID string) error
	ValidateCSVData(data string) error
	CheckFileExists(path string) error
}

// DefaultValidator is the default implementation of Validator
type DefaultValidator struct{}

// NewDefaultValidator creates a new default validator
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// ValidateCSVFilePath validates CSV file path
func (v *DefaultValidator) ValidateCSVFilePath(path string) error {
	if path == "" {
		return status.Error(codes.InvalidArgument, "csv_file_path is required")
	}
	return nil
}

// ValidateAccountID validates account ID
func (v *DefaultValidator) ValidateAccountID(accountID string) error {
	// account_id is optional
	if accountID == "" {
		return nil
	}
	// Additional validation rules can be added here
	if len(accountID) < 3 {
		return status.Error(codes.InvalidArgument, "account_id must be at least 3 characters")
	}
	return nil
}

// ValidateCSVData validates CSV data
func (v *DefaultValidator) ValidateCSVData(data string) error {
	if data == "" {
		return status.Error(codes.InvalidArgument, "csv_data is required")
	}
	if len(data) < 10 {
		return status.Error(codes.InvalidArgument, "csv_data is too short")
	}
	return nil
}

// CheckFileExists checks if a file exists
func (v *DefaultValidator) CheckFileExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return status.Errorf(codes.NotFound, "file not found: %s", path)
	} else if err != nil {
		return status.Errorf(codes.Internal, "failed to check file: %v", err)
	}
	return nil
}

// ValidateProcessCSVFileRequest validates ProcessCSVFile request
func ValidateProcessCSVFileRequest(req interface{}, v Validator) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}

	// Type assertion with interface to allow different request types
	type FileRequest interface {
		GetCsvFilePath() string
		GetAccountId() string
	}

	fileReq, ok := req.(FileRequest)
	if !ok {
		return status.Error(codes.InvalidArgument, "invalid request type")
	}

	if err := v.ValidateCSVFilePath(fileReq.GetCsvFilePath()); err != nil {
		return err
	}

	if err := v.ValidateAccountID(fileReq.GetAccountId()); err != nil {
		return err
	}

	if err := v.CheckFileExists(fileReq.GetCsvFilePath()); err != nil {
		return err
	}

	return nil
}

// ValidateProcessCSVDataRequest validates ProcessCSVData request
func ValidateProcessCSVDataRequest(req interface{}, v Validator) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}

	type DataRequest interface {
		GetCsvData() string
		GetAccountId() string
	}

	dataReq, ok := req.(DataRequest)
	if !ok {
		return status.Error(codes.InvalidArgument, "invalid request type")
	}

	if err := v.ValidateCSVData(dataReq.GetCsvData()); err != nil {
		return err
	}

	if err := v.ValidateAccountID(dataReq.GetAccountId()); err != nil {
		return err
	}

	return nil
}

// ValidateValidateCSVDataRequest validates ValidateCSVData request
func ValidateValidateCSVDataRequest(req interface{}, v Validator) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}

	type ValidateRequest interface {
		GetCsvData() string
	}

	validateReq, ok := req.(ValidateRequest)
	if !ok {
		return status.Error(codes.InvalidArgument, "invalid request type")
	}

	if err := v.ValidateCSVData(validateReq.GetCsvData()); err != nil {
		return err
	}

	return nil
}

// CreateDuplicateKey creates a unique key for duplicate detection
func CreateDuplicateKey(entryDate, entryTime, exitDate, exitTime string, amount int, cardNumber string) string {
	return fmt.Sprintf("%s_%s_%s_%s_%d_%s",
		entryDate, entryTime, exitDate, exitTime, amount, cardNumber)
}