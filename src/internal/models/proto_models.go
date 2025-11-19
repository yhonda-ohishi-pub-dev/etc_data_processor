package models

import "time"

// ProtoService defines the gRPC service
type ProtoService struct {
	Name    string `proto:"DataProcessorService"`
	Package string `proto:"etcdataprocessor.v1"`
	GoPackage string `proto:"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/api/pb;pb"`
}

// ProcessCSVFileRequest represents request for CSV file processing
type ProcessCSVFileRequest struct {
	CSVFilePath    string `json:"csv_file_path" proto:"1"`
	AccountID      string `json:"account_id" proto:"2"`
	SkipDuplicates bool   `json:"skip_duplicates" proto:"3"`
}

// ProcessCSVFileResponse represents response for CSV file processing
type ProcessCSVFileResponse struct {
	Success bool             `json:"success" proto:"1"`
	Message string           `json:"message" proto:"2"`
	Stats   *ProcessingStats `json:"stats" proto:"3"`
	Errors  []string         `json:"errors" proto:"4,repeated"`
}

// ProcessCSVDataRequest represents request for CSV data processing
type ProcessCSVDataRequest struct {
	CSVData        string `json:"csv_data" proto:"1"`
	AccountID      string `json:"account_id" proto:"2"`
	SkipDuplicates bool   `json:"skip_duplicates" proto:"3"`
}

// ProcessCSVDataResponse represents response for CSV data processing
type ProcessCSVDataResponse struct {
	Success bool             `json:"success" proto:"1"`
	Message string           `json:"message" proto:"2"`
	Stats   *ProcessingStats `json:"stats" proto:"3"`
	Errors  []string         `json:"errors" proto:"4,repeated"`
}

// ValidateCSVDataRequest represents request for CSV validation
type ValidateCSVDataRequest struct {
	CSVData   string `json:"csv_data" proto:"1"`
	AccountID string `json:"account_id" proto:"2"`
}

// ValidateCSVDataResponse represents response for CSV validation
type ValidateCSVDataResponse struct {
	IsValid        bool               `json:"is_valid" proto:"1"`
	Errors         []ValidationError  `json:"errors" proto:"2,repeated"`
	DuplicateCount int32              `json:"duplicate_count" proto:"3"`
	TotalRecords   int32              `json:"total_records" proto:"4"`
}

// HealthCheckRequest represents health check request
type HealthCheckRequest struct{}

// HealthCheckResponse represents health check response
type HealthCheckResponse struct {
	Status    string            `json:"status" proto:"1"`
	Version   string            `json:"version" proto:"2"`
	Timestamp int64             `json:"timestamp" proto:"3"`
	Details   map[string]string `json:"details" proto:"4,map"`
}

// ProcessingStats represents processing statistics
type ProcessingStats struct {
	TotalRecords   int32 `json:"total_records" proto:"1"`
	SavedRecords   int32 `json:"saved_records" proto:"2"`
	SkippedRecords int32 `json:"skipped_records" proto:"3"`
	ErrorRecords   int32 `json:"error_records" proto:"4"`
}

// ValidationError represents validation error details
type ValidationError struct {
	LineNumber int32  `json:"line_number" proto:"1"`
	Field      string `json:"field" proto:"2"`
	Message    string `json:"message" proto:"3"`
	RecordData string `json:"record_data" proto:"4"`
}

// ServiceMethod represents a gRPC service method
type ServiceMethod struct {
	Name       string      `json:"name"`
	Request    interface{} `json:"request"`
	Response   interface{} `json:"response"`
	HTTPMethod string      `json:"http_method"`
	HTTPPath   string      `json:"http_path"`
}

// ServiceDefinition for generating proto file
type ServiceDefinition struct {
	Service ProtoService
	Methods []ServiceMethod
}

// GetServiceDefinition returns the service definition for proto generation
func GetServiceDefinition() ServiceDefinition {
	return ServiceDefinition{
		Service: ProtoService{
			Name:      "DataProcessorService",
			Package:   "etcdataprocessor.v1",
			GoPackage: "github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/api/pb;pb",
		},
		Methods: []ServiceMethod{
			{
				Name:       "ProcessCSVFile",
				Request:    ProcessCSVFileRequest{},
				Response:   ProcessCSVFileResponse{},
				HTTPMethod: "POST",
				HTTPPath:   "/v1/process/file",
			},
			{
				Name:       "ProcessCSVData",
				Request:    ProcessCSVDataRequest{},
				Response:   ProcessCSVDataResponse{},
				HTTPMethod: "POST",
				HTTPPath:   "/v1/process/data",
			},
			{
				Name:       "ValidateCSVData",
				Request:    ValidateCSVDataRequest{},
				Response:   ValidateCSVDataResponse{},
				HTTPMethod: "POST",
				HTTPPath:   "/v1/validate",
			},
			{
				Name:       "HealthCheck",
				Request:    HealthCheckRequest{},
				Response:   HealthCheckResponse{},
				HTTPMethod: "GET",
				HTTPPath:   "/v1/health",
			},
		},
	}
}

// For actual service implementation
type ETCRecord struct {
	Date        time.Time
	EntryIC     string
	ExitIC      string
	Route       string
	VehicleType string
	Amount      int
	CardNumber  string
}