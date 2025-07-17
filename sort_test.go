package table

import (
	"reflect"
	"testing"
)

func TestStringComparison(t *testing.T) {
	// Test cases
	cases := []struct {
		a, b   string
		result int
	}{
		{"apple", "banana", -1},
		{"banana", "apple", 1},
		{"apple", "apple", 0},
	}

	// Iterate over test cases
	for _, c := range cases {
		got := StringComparison(c.a, c.b)
		if got != c.result {
			t.Errorf("StringComparison(%q, %q) == %d, want %d", c.a, c.b, got, c.result)
		}
	}
}

func TestBoolComparison(t *testing.T) {
	// Test cases
	cases := []struct {
		a, b   string
		result int
	}{
		{"true", "false", 1},
		{"false", "true", -1},
		{"true", "true", 0},
		{"false", "false", 0},
		{"invalid", "false", -1}, // Test invalid input handling
	}

	// Iterate over test cases
	for _, c := range cases {
		got := BoolComparison(c.a, c.b)
		if got != c.result {
			t.Errorf("BoolComparison(%q, %q) == %d, want %d", c.a, c.b, got, c.result)
		}
	}
}

func TestNumericalComparison(t *testing.T) {
	// Test cases
	cases := []struct {
		a, b   string
		result int
	}{
		{"10", "5", 1},
		{"5", "10", -1},
		{"5", "5", 0},
		{"invalid", "5", 1}, // Test invalid input handling
	}

	// Iterate over test cases
	for _, c := range cases {
		got := NumericalComparison(c.a, c.b)
		if got != c.result {
			t.Errorf("NumericalComparison(%q, %q) == %d, want %d", c.a, c.b, got, c.result)
		}
	}
}

func TestCurrencyComparison(t *testing.T) {
	// Test cases
	cases := []struct {
		a, b   string
		result int
	}{
		{"$100.00", "$50.00", 1},
		{"$50.00", "$100.00", -1},
		{"$50.00", "$50.00", 0},
		{"invalid", "$50.00", 1}, // Test invalid input handling
	}

	// Iterate over test cases
	for _, c := range cases {
		got := CurrencyComparison(c.a, c.b)
		if got != c.result {
			t.Errorf("CurrencyComparison(%q, %q) == %d, want %d", c.a, c.b, got, c.result)
		}
	}
}

func TestPercentComparison(t *testing.T) {
	// Test cases
	cases := []struct {
		a, b   string
		result int
	}{
		{"10%", "5%", 1},
		{"5%", "10%", -1},
		{"5%", "5%", 0},
		{"invalid", "5%", 1}, // Test invalid input handling
	}

	// Iterate over test cases
	for _, c := range cases {
		got := PercentComparison(c.a, c.b)
		if got != c.result {
			t.Errorf("PercentComparison(%q, %q) == %d, want %d", c.a, c.b, got, c.result)
		}
	}
}

// TestSortByMultiple tests the SortByMultiple function of the table struct.
func TestSortByMultiple(t *testing.T) {
	// Example table
	testTable := table{
		header: []string{"Name", "Age", "City"},
		rows: [][]string{
			{"Alice", "30", "New York"},
			{"Bob", "25", "Los Angeles"},
			{"Charlie", "35", "Chicago"},
			{"David", "30", "Chicago"},
			{"Eve", "35", "Los Angeles"},
		},
	}

	// Define the sorting criteria
	criteria := []*SortCriterion{
		&SortCriterion{ColumnIndex: 1, Compare: StringComparison},                    // Sort by Age (ascending)
		&SortCriterion{ColumnIndex: 0, Compare: InverseComparison(StringComparison)}, // Then by Name (descending)
	}

	// Sort the table
	err := testTable.SortByMultiple(criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Define the expected sorted rows
	expectedRows := [][]string{
		{"Bob", "25", "Los Angeles"},
		{"David", "30", "Chicago"},
		{"Alice", "30", "New York"},
		{"Eve", "35", "Los Angeles"},
		{"Charlie", "35", "Chicago"},
	}

	// Verify the result
	if !reflect.DeepEqual(testTable.rows, expectedRows) {
		t.Errorf("sorted rows do not match expected rows. Got %v, expected %v", testTable.rows, expectedRows)
	}
}
