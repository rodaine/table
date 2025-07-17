package table

import (
	"encoding/json"
	"fmt"
)

// ExportJSON exports the table data in JSON format using a specified column as the key.
func (t *table) ExportJSON(keyColumn int) error {
	if keyColumn < 0 || keyColumn >= len(t.header) {
		return fmt.Errorf("invalid keyColumn index")
	}

	// Create a map to hold the exportable data
	data := make(map[string]map[string]string)

	// Populate the map with table's data
	for _, row := range t.rows {
		if keyColumn >= len(row) {
			return fmt.Errorf("keyColumn index out of range in row")
		}

		key := row[keyColumn]
		rowMap := make(map[string]string)
		for i, col := range t.header {
			if i < len(row) {
				rowMap[col] = row[i]
			} else {
				rowMap[col] = ""
			}
		}
		data[key] = rowMap
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Write the JSON data to the Writer
	_, err = t.Writer.Write(jsonData)
	return err
}
