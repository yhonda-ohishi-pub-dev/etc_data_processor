package handler

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	pb "github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/proto"
	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	version = "1.0.0"
)

// DBClient interface for database operations
type DBClient interface {
	SaveETCData(data interface{}) error
}

// Parser interface for CSV parsing operations
type Parser interface {
	ParseFile(filePath string) ([]parser.ActualETCRecord, error)
	Parse(reader io.Reader) ([]parser.ActualETCRecord, error)
	ValidateRecord(record parser.ActualETCRecord) error
	ConvertToSimpleRecord(record parser.ActualETCRecord) (parser.ETCRecord, error)
}

// DataProcessorService implements the gRPC service
type DataProcessorService struct {
	pb.UnimplementedDataProcessorServiceServer
	dbClient  DBClient
	parser    Parser
	validator Validator
}

// NewDataProcessorService creates a new service instance
func NewDataProcessorService(dbClient DBClient) *DataProcessorService {
	return &DataProcessorService{
		dbClient:  dbClient,
		parser:    parser.NewETCCSVParser(),
		validator: NewDefaultValidator(),
	}
}

// NewDataProcessorServiceWithValidator creates a service with custom validator
func NewDataProcessorServiceWithValidator(dbClient DBClient, validator Validator) *DataProcessorService {
	return &DataProcessorService{
		dbClient:  dbClient,
		parser:    parser.NewETCCSVParser(),
		validator: validator,
	}
}

// NewDataProcessorServiceWithDependencies creates a service with custom dependencies
func NewDataProcessorServiceWithDependencies(dbClient DBClient, csvParser Parser, validator Validator) *DataProcessorService {
	return &DataProcessorService{
		dbClient:  dbClient,
		parser:    csvParser,
		validator: validator,
	}
}

// ProcessCSVFile processes a CSV file from filesystem
func (s *DataProcessorService) ProcessCSVFile(ctx context.Context, req *pb.ProcessCSVFileRequest) (*pb.ProcessCSVFileResponse, error) {
	// Validate request using validator
	if err := ValidateProcessCSVFileRequest(req, s.validator); err != nil {
		return nil, err
	}

	// Parse CSV file
	records, err := s.parser.ParseFile(req.CsvFilePath)
	if err != nil {
		return &pb.ProcessCSVFileResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse CSV file: %v", err),
			Stats: &pb.ProcessingStats{
				TotalRecords: 0,
			},
			Errors: []string{err.Error()},
		}, nil
	}

	// Process records
	stats, errors := s.processRecords(ctx, records, req.AccountId, req.SkipDuplicates)

	return &pb.ProcessCSVFileResponse{
		Success: stats.SavedRecords > 0,
		Message: fmt.Sprintf("Processed %d records from file", stats.TotalRecords),
		Stats:   stats,
		Errors:  errors,
	}, nil
}

// ProcessCSVData processes CSV data directly
func (s *DataProcessorService) ProcessCSVData(ctx context.Context, req *pb.ProcessCSVDataRequest) (*pb.ProcessCSVDataResponse, error) {
	// Validate request using validator
	if err := ValidateProcessCSVDataRequest(req, s.validator); err != nil {
		return nil, err
	}

	// Parse CSV data
	reader := strings.NewReader(req.CsvData)
	records, err := s.parser.Parse(reader)
	if err != nil {
		// All parsing errors should be treated as invalid format for API
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSV format: %v", err)
	}

	// Process records
	stats, errors := s.processRecords(ctx, records, req.AccountId, req.SkipDuplicates)

	return &pb.ProcessCSVDataResponse{
		Success: stats.SavedRecords > 0,
		Message: fmt.Sprintf("Processed %d records", stats.TotalRecords),
		Stats:   stats,
		Errors:  errors,
	}, nil
}

// ValidateCSVData validates CSV data without saving
func (s *DataProcessorService) ValidateCSVData(ctx context.Context, req *pb.ValidateCSVDataRequest) (*pb.ValidateCSVDataResponse, error) {
	// Validate request using validator
	if err := ValidateValidateCSVDataRequest(req, s.validator); err != nil {
		return nil, err
	}

	// Parse CSV data
	reader := strings.NewReader(req.CsvData)
	records, err := s.parser.Parse(reader)

	var validationErrors []*pb.ValidationError

	if err != nil {
		// Parse error means invalid CSV
		return &pb.ValidateCSVDataResponse{
			IsValid: false,
			Errors: []*pb.ValidationError{
				{
					LineNumber: 0,
					Field:      "csv",
					Message:    err.Error(),
				},
			},
			TotalRecords: 0,
		}, nil
	}

	// Validate each record
	duplicateMap := make(map[string]int)
	duplicateCount := int32(0)

	for i, record := range records {
		// Create a unique key for duplicate detection
		key := fmt.Sprintf("%s_%s_%s_%s_%d",
			record.EntryDate, record.EntryTime,
			record.ExitDate, record.ExitTime,
			record.ETCAmount)

		if _, exists := duplicateMap[key]; exists {
			duplicateCount++
			duplicateMap[key]++
		} else {
			duplicateMap[key] = 1
		}

		// Validate record
		if err := s.parser.ValidateRecord(record); err != nil {
			validationErrors = append(validationErrors, &pb.ValidationError{
				LineNumber:  int32(i + 2), // +2 for header and 1-based indexing
				Field:       "",
				Message:     err.Error(),
				RecordData:  fmt.Sprintf("%v", record),
			})
		}
	}

	return &pb.ValidateCSVDataResponse{
		IsValid:        len(validationErrors) == 0,
		Errors:         validationErrors,
		DuplicateCount: duplicateCount,
		TotalRecords:   int32(len(records)),
	}, nil
}

// HealthCheck returns the service health status
func (s *DataProcessorService) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:    "healthy",
		Version:   version,
		Timestamp: time.Now().Unix(),
		Details: map[string]string{
			"service": "etc_data_processor",
			"uptime":  "running",
		},
	}, nil
}

// processRecords processes parsed records and saves to database
func (s *DataProcessorService) processRecords(ctx context.Context, records []parser.ActualETCRecord, accountID string, skipDuplicates bool) (*pb.ProcessingStats, []string) {
	stats := &pb.ProcessingStats{
		TotalRecords:   int32(len(records)),
		SavedRecords:   0,
		SkippedRecords: 0,
		ErrorRecords:   0,
	}

	var errors []string
	processedKeys := make(map[string]bool)

	for i, record := range records {
		// Check context cancellation
		if ctx.Err() != nil {
			errors = append(errors, fmt.Sprintf("Processing cancelled at record %d", i))
			stats.ErrorRecords = int32(len(records) - i)
			break
		}

		// Create unique key for duplicate detection
		key := fmt.Sprintf("%s_%s_%s_%s_%d_%s",
			record.EntryDate, record.EntryTime,
			record.ExitDate, record.ExitTime,
			record.ETCAmount, record.CardNumber)

		// Skip duplicates if requested
		if skipDuplicates && processedKeys[key] {
			stats.SkippedRecords++
			continue
		}

		// Convert to simple format for saving
		simpleRecord, err := s.parser.ConvertToSimpleRecord(record)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Record %d: conversion failed: %v", i+1, err))
			stats.ErrorRecords++
			continue
		}

		// Add account ID
		dataToSave := map[string]interface{}{
			"account_id":   accountID,
			"date":        simpleRecord.Date.Format("2006-01-02"),
			"entry_ic":    simpleRecord.EntryIC,
			"exit_ic":     simpleRecord.ExitIC,
			"route":       simpleRecord.Route,
			"vehicle_type": simpleRecord.VehicleType,
			"amount":      simpleRecord.Amount,
			"card_number": simpleRecord.CardNumber,
		}

		// Save to database
		if s.dbClient != nil {
			if err := s.dbClient.SaveETCData(dataToSave); err != nil {
				errors = append(errors, fmt.Sprintf("Record %d: save failed: %v", i+1, err))
				stats.ErrorRecords++
				continue
			}
		}

		processedKeys[key] = true
		stats.SavedRecords++
	}

	return stats, errors
}