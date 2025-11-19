package unit

import (
	"testing"

	"github.com/yhonda-ohishi-pub-dev/etc_data_processor/src/pkg/parser"
)

// Test ParseVehicleClass function directly
func TestETCCSVParser_ParseVehicleClass(t *testing.T) {
	p := parser.NewETCCSVParser()

	tests := []struct {
		name        string
		record      []string
		fieldIndex  int
		expected    int
		description string
	}{
		{
			name:        "valid numeric class",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "2", "m"},
			fieldIndex:  11,
			expected:    2,
			description: "Should parse '2' as vehicle class 2",
		},
		{
			name:        "non-numeric class",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "普通車", "m"},
			fieldIndex:  11,
			expected:    0,
			description: "Should return 0 for non-numeric value '普通車'",
		},
		{
			name:        "empty field",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "", "m"},
			fieldIndex:  11,
			expected:    0,
			description: "Should return 0 for empty field",
		},
		{
			name:        "invalid field index",
			record:      []string{"a", "b", "c"},
			fieldIndex:  11,
			expected:    0,
			description: "Should return 0 for field index out of range",
		},
		{
			name:        "zero value",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "0", "m"},
			fieldIndex:  11,
			expected:    0,
			description: "Should parse '0' as vehicle class 0",
		},
		{
			name:        "large number",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "999", "m"},
			fieldIndex:  11,
			expected:    999,
			description: "Should parse large numbers correctly",
		},
		{
			name:        "negative number",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "-1", "m"},
			fieldIndex:  11,
			expected:    -1,
			description: "Should parse negative numbers correctly",
		},
		{
			name:        "alphanumeric string",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "2abc", "m"},
			fieldIndex:  11,
			expected:    0,
			description: "Should return 0 for alphanumeric string '2abc'",
		},
		{
			name:        "spaces around number",
			record:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", " 3 ", "m"},
			fieldIndex:  11,
			expected:    0,
			description: "Should return 0 for spaces around number (strconv.Atoi doesn't trim)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.ParseVehicleClass(tt.record, tt.fieldIndex)

			if result != tt.expected {
				t.Errorf("ParseVehicleClass() = %d, expected %d. %s", result, tt.expected, tt.description)
			}
		})
	}
}

// Test ParseVehicleClass with real ETC CSV data scenarios
func TestETCCSVParser_ParseVehicleClass_RealScenarios(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Scenario 1: Normal ETC record with vehicle class
	etcRecord := []string{
		"25/09/01", "08:00", "25/09/01", "09:00", "東京", "横浜",
		"1500", "-300", "1200", "2", "1234", "********12345678", "テスト",
	}

	result := p.ParseVehicleClass(etcRecord, 9) // Vehicle class at index 9
	if result != 2 {
		t.Errorf("Expected vehicle class 2, got %d", result)
	}

	// Scenario 2: ETC record with non-numeric vehicle class
	etcRecordInvalid := []string{
		"25/09/01", "08:00", "25/09/01", "09:00", "東京", "横浜",
		"1500", "-300", "1200", "普通車", "1234", "********12345678", "テスト",
	}

	result = p.ParseVehicleClass(etcRecordInvalid, 9)
	if result != 0 {
		t.Errorf("Expected vehicle class 0 for non-numeric value, got %d", result)
	}

	// Scenario 3: Short record (missing vehicle class field)
	shortRecord := []string{
		"25/09/01", "08:00", "25/09/01", "09:00", "東京",
	}

	result = p.ParseVehicleClass(shortRecord, 9)
	if result != 0 {
		t.Errorf("Expected vehicle class 0 for short record, got %d", result)
	}
}

// Test ParseVehicleClass error handling specifically for strconv.Atoi errors
func TestETCCSVParser_ParseVehicleClass_StrconvErrors(t *testing.T) {
	p := parser.NewETCCSVParser()

	// Test cases that specifically trigger strconv.Atoi errors
	errorCases := []struct {
		name     string
		value    string
		expected int
	}{
		{"non-numeric string", "abc", 0},
		{"mixed alphanumeric", "123abc", 0},
		{"decimal number", "2.5", 0},
		{"special characters", "2@#$", 0},
		{"unicode characters", "２", 0}, // Full-width number
		{"hex-like string", "0x1A", 0},
		{"scientific notation", "1e5", 0},
		{"very large number", "999999999999999999999", 0}, // Overflow
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			record := make([]string, 15)
			record[11] = tc.value

			result := p.ParseVehicleClass(record, 11)
			if result != tc.expected {
				t.Errorf("ParseVehicleClass('%s') = %d, expected %d", tc.value, result, tc.expected)
			}
		})
	}
}