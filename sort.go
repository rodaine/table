package table

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ComparisonFunc is a function type for comparing two strings.
type ComparisonFunc func(string, string) int

// InverseComparison takes a ComparisonFunc and returns a new ComparisonFunc that inverts the result.
func InverseComparison(cmp ComparisonFunc) ComparisonFunc {
	return func(a, b string) int {
		return -cmp(a, b)
	}
}

// StringComparison compares two strings
func StringComparison(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// BoolComparison compares two bools and returns -1, 0, or 1
func BoolComparison(a, b string) int {
	_a, errA := strconv.ParseBool(a)
	_b, errB := strconv.ParseBool(b)
	if errA != nil || errB != nil {
		// Handle the error case, for example, treating invalid bools as false
		if errA != nil && errB != nil {
			return 0
		} else if errA != nil {
			return -1
		}
		return 1
	}
	if _a == _b {
		return 0
	} else if !_a && _b {
		return -1
	}
	return 1
}

// NumericalComparison compares two numbers (parsed from strings) and returns -1, 0, or 1
func NumericalComparison(a, b string) int {
	numA, errA := strconv.ParseFloat(a, 64)
	numB, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		// Handle the error case, for example, treating invalid numbers as greater to place them after valid numbers
		if errA != nil && errB != nil {
			return 0
		} else if errA != nil {
			return 1
		}
		return -1
	}
	if numA < numB {
		return -1
	} else if numA > numB {
		return 1
	}
	return 0
}

// parseCurrencyString parses a string representing a currency value like "$511,753,000.00" into a float64
func parseCurrencyString(s string) (float64, error) {
	// Remove non-numeric characters
	cleanStr := strings.ReplaceAll(s, "$", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")

	// Parse the cleaned string into a float64
	return strconv.ParseFloat(cleanStr, 64)
}

// CurrencyComparison compares two currency values (parsed from strings) and returns -1, 0, or 1
func CurrencyComparison(a, b string) int {
	numA, errA := parseCurrencyString(a)
	numB, errB := parseCurrencyString(b)
	if errA != nil || errB != nil {
		// Handle the error case, for example, treating invalid currency values as greater to place them after valid values
		if errA != nil && errB != nil {
			return 0
		} else if errA != nil {
			return 1
		}
		return -1
	}
	if numA < numB {
		return -1
	} else if numA > numB {
		return 1
	}
	return 0
}

// parsePercentString parses a string representing a percent value like "10%" into a float64
func parsePercentString(s string) (float64, error) {
	// Remove non-numeric characters
	cleanStr := strings.ReplaceAll(s, "%", "")

	// Parse the cleaned string into a float64
	return strconv.ParseFloat(cleanStr, 64)
}

// PercentComparison compares two percent values (parsed from strings) and returns -1, 0, or 1
func PercentComparison(a, b string) int {
	numA, errA := parsePercentString(a)
	numB, errB := parsePercentString(b)
	if errA != nil || errB != nil {
		// Handle the error case, for example, treating invalid percent values as greater to place them after valid values
		if errA != nil && errB != nil {
			return 0
		} else if errA != nil {
			return 1
		}
		return -1
	}
	if numA < numB {
		return -1
	} else if numA > numB {
		return 1
	}
	return 0
}

// SortBy sorts the table rows based on the values in the specified
// column (indexed from 0) using the provided comparison function.  It
// returns an error if the specified column index is out of bounds.
func (t *table) SortBy(n int, cmpFn ComparisonFunc) error {
	// Check if n is a valid col index
	if n < 0 || n >= len(t.header) {
		return fmt.Errorf("invalid col index %d (len %d)", n, len(t.header))
	}

	// Sort rows based on the values in the nth column using the provided comparison function
	sort.SliceStable(t.rows, func(i, j int) bool {
		return cmpFn(t.rows[i][n], t.rows[j][n]) < 0
	})

	return nil
}

// SortCriterion represents a single column index and its comparison function.
type SortCriterion struct {
	ColumnIndex int
	Compare     ComparisonFunc
}

// SortByMultiple sorts the table rows based on multiple sorting criteria.
// Each criterion specifies a column index and its comparison function.
// It returns an error if any specified column index is out of bounds.
func (t *table) SortByMultiple(criteria []*SortCriterion) error {
	for _, criterion := range criteria {
		// Check if each column index is valid
		if criterion.ColumnIndex < 0 || criterion.ColumnIndex >= len(t.header) {
			return fmt.Errorf("invalid col index %d (len %d)", criterion.ColumnIndex, len(t.header))
		}
	}

	// Sort rows based on the values in the specified columns using the provided comparison functions
	sort.SliceStable(t.rows, func(i, j int) bool {
		for _, criterion := range criteria {
			colIdx := criterion.ColumnIndex
			cmpFn := criterion.Compare

			// Compare rows based on the current criterion
			result := cmpFn(t.rows[i][colIdx], t.rows[j][colIdx])
			if result < 0 {
				return true
			} else if result > 0 {
				return false
			}
			// If the values are equal, continue to the next criterion
		}
		return false
	})

	return nil
}
