package table

// Transpose the table such that the first column becomes the header and the rest of the data is transposed accordingly.
func (t *table) Transpose() *table {
	// Create new header from the first column of each row, plus the first header element
	newHeader := make([]string, 0)
	newHeader = append(newHeader, t.header[0])
	for i := 0; i < len(t.rows); i++ {
		newHeader = append(newHeader, t.rows[i][0])
	}

	// Initialize new rows
	newRows := make([][]string, len(t.header)-1)
	for i := range newRows {
		newRows[i] = make([]string, len(t.rows)+1)
	}

	// Fill in the new rows with transposed data
	for i, header := range t.header[1:] {
		newRows[i][0] = header
		for j, row := range t.rows {
			newRows[i][j+1] = row[i+1]
		}
	}

	return &table{
		FirstColumnFormatter: t.FirstColumnFormatter,
		StandoutFormatter:    t.StandoutFormatter,
		HeaderFormatter:      t.HeaderFormatter,
		Padding:              t.Padding,
		Writer:               t.Writer,
		Width:                t.Width,
		HeaderSeparatorRune:  t.HeaderSeparatorRune,
		PrintHeaders:         t.PrintHeaders,
		header:               newHeader,
		rows:                 newRows,
		widths:               t.widths,
	}
}
