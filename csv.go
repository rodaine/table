package table

import (
	"encoding/csv"
	"fmt"
)

// ExportCSV exports the table data in CSV format.
func (t *table) ExportCSV() error {
	// Create a CSV writer
	writer := csv.NewWriter(t.Writer)
	defer writer.Flush()

	// Write the header row
	if err := writer.Write(t.header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Write each row
	for _, row := range t.rows {
		if len(row) < len(t.header) {
			// Ensure the row has the same number of columns as the header
			paddedRow := make([]string, len(t.header))
			copy(paddedRow, row)
			row = paddedRow
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	// Check for any errors that occurred during writing
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error occurred while writing CSV: %v", err)
	}

	return nil
}
