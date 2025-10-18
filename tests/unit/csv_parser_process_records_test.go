package unit

import (
	"errors"
	"strings"
	"testing"

	"github.com/yhonda-ohishi/etc_data_processor/src/pkg/parser"
)

// TestCSVParserWithValidationError is a custom parser for testing validation errors
type TestCSVParserWithValidationError struct {
	*parser.CSVParser
}

// ValidateRecord overrides to return validation errors for testing
func (p *TestCSVParserWithValidationError) ValidateRecord(record parser.ETCRecord) error {
	// Trigger validation error for empty EntryIC
	if record.EntryIC == "" {
		return errors.New("entry IC cannot be empty")
	}
	return nil
}

// Test processRecords with validation error to cover the ValidateRecord error path
func TestCSVParser_ProcessRecords_ValidateError(t *testing.T) {
	p := &TestCSVParserWithValidationError{parser.NewCSVParser()}

	// Create test data with invalid record (empty EntryIC)
	records := [][]string{
		{"2023-09-01", "", "横浜IC", "東名高速", "1", "1000", "1234567890"},
	}

	// Test ProcessRecords directly
	_, err := p.ProcessRecords(records, 0)

	// Should return validation error
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	if !strings.Contains(err.Error(), "validation error at line 1") {
		t.Errorf("Expected 'validation error at line 1', got: %v", err)
	}

	if !strings.Contains(err.Error(), "entry IC cannot be empty") {
		t.Errorf("Expected 'entry IC cannot be empty', got: %v", err)
	}
}


// Test ProcessRecords with real CSVParser validation
func TestCSVParser_ProcessRecords_RealValidation(t *testing.T) {
	p := parser.NewCSVParser()

	// Test with empty EntryIC (should trigger validation error)
	records := [][]string{
		{"2023-09-01", "", "横浜IC", "東名高速", "1", "1000", "1234567890"},
	}

	_, err := p.ProcessRecords(records, 0)

	// Should return validation error
	if err == nil {
		t.Error("Expected validation error for empty EntryIC, got nil")
	}

	if !strings.Contains(err.Error(), "validation error at line 1") {
		t.Errorf("Expected 'validation error at line 1', got: %v", err)
	}

	if !strings.Contains(err.Error(), "entry IC cannot be empty") {
		t.Errorf("Expected 'entry IC cannot be empty', got: %v", err)
	}
}