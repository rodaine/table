// Package table provides a convenient way to generate tabular output of any
// data, primarily useful for CLI tools.
//
// Columns are left-aligned and padded to accomodate the largest cell in that
// column.
//
// Source: https://github.com/rodaine/table
//
//	table.DefaultHeaderFormatter = func(format string, vals ...interface{}) string {
//	  return strings.ToUpper(fmt.Sprintf(format, vals...))
//	}
//
//	tbl := table.New("ID", "Name", "Cost ($)")
//
//	for _, widget := range Widgets {
//	  tbl.AddRow(widget.ID, widget.Name, widget.Cost)
//	}
//
//	tbl.Print()
//
//	// Output:
//	// ID  NAME      COST ($)
//	// 1   Foobar    1.23
//	// 2   Fizzbuzz  4.56
//	// 3   Gizmo     78.90
package table

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// These are the default properties for all Tables created from this package
// and can be modified.
var (
	// DefaultPadding specifies the number of spaces between columns in a table.
	DefaultPadding = 2

	// DefaultWriter specifies the output io.Writer for the Table.Print method.
	DefaultWriter io.Writer = os.Stdout

	// DefaultHeaderFormatter specifies the default Formatter for the table header.
	DefaultHeaderFormatter Formatter

	// DefaultFirstColumnFormatter specifies the default Formatter for the first column cells.
	DefaultFirstColumnFormatter Formatter

	// DefaultStandoutFormatter specifies the default Formatter for alternating rows.
	DefaultStandoutFormatter Formatter

	// DefaultWidthFunc specifies the default WidthFunc for calculating column widths
	DefaultWidthFunc WidthFunc = utf8.RuneCountInString

	// DefaultPrintHeaders specifies if headers should be printed
	DefaultPrintHeaders = true
)

// Formatter functions expose a fmt.Sprintf signature that can be used to modify
// the display of the text in either the header or first column of a Table.
// The formatter should not change the width of original text as printed since
// column widths are calculated pre-formatting (though this issue can be mitigated
// with increased padding).
//
//	tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
//	  return strings.ToUpper(fmt.Sprintf(format, vals...))
//	})
//
// A good use case for formatters is to use ANSI escape codes to color the cells
// for a nicer interface. The package color (https://github.com/fatih/color) makes
// it easy to generate these automatically: http://godoc.org/github.com/fatih/color#Color.SprintfFunc
type Formatter func(string, ...interface{}) string

// A WidthFunc calculates the width of a string. By default, the number of runes
// is used but this may not be appropriate for certain character sets. The
// package runewidth (https://github.com/mattn/go-runewidth) could be used to
// accomodate multi-cell characters (such as emoji or CJK characters).
type WidthFunc func(string) int

// ComparisonFunc is a function type for comparing two strings.
type ComparisonFunc func(string, string) int

// SortCriterion represents a single column index and its comparison function.
type SortCriterion struct {
	ColumnIndex int
	Compare     ComparisonFunc
}

// Table describes the interface for building up a tabular representation of data.
// It exposes fluent/chainable methods for convenient table building.
//
// WithHeaderFormatter and WithFirstColumnFormatter sets the Formatter for the
// header and first column, respectively. If nil is passed in (the default), no
// formatting will be applied.
//
//	New("foo", "bar").WithFirstColumnFormatter(func(f string, v ...interface{}) string {
//	  return strings.ToUpper(fmt.Sprintf(f, v...))
//	})
//
// WithPadding specifies the minimum padding between cells in a row and defaults
// to DefaultPadding. Padding values less than or equal to zero apply no extra
// padding between the columns.
//
//	New("foo", "bar").WithPadding(3)
//
// WithWriter modifies the writer which Print outputs to, defaulting to DefaultWriter
// when instantiated. If nil is passed, os.Stdout will be used.
//
//	New("foo", "bar").WithWriter(os.Stderr)
//
// WithWidthFunc sets the function used to calculate the width of the string in
// a column. By default, the number of utf8 runes in the string is used.
//
// WithPrintHeaders specifies whether if the headers of the table should be
// printed or not, which might be useful if the output is being piped to other
// processes. By default, they are printed.
//
// AddRow adds another row of data to the table. Any values can be passed in and
// will be output as its string representation as described in the fmt standard
// package. Rows can have less cells than the total number of columns in the table;
// subsequent cells will be rendered empty. Rows with more cells than the total
// number of columns will be truncated. References to the data are not held, so
// the passed in values can be modified without affecting the table's output.
//
//	New("foo", "bar").AddRow("fizz", "buzz").AddRow(time.Now()).AddRow(1, 2, 3).Print()
//	// Output:
//	// foo                              bar
//	// fizz                             buzz
//	// 2006-01-02 15:04:05.0 -0700 MST
//	// 1                                2
//
// Print writes the string representation of the table to the provided writer.
// Print can be called multiple times, even after subsequent mutations of the
// provided data. The output is always preceded and followed by a new line.
type Table interface {
	WithHeaderFormatter(f Formatter) Table
	WithFirstColumnFormatter(f Formatter) Table
	WithStandoutFormatter(f Formatter) Table
	WithPadding(p int) Table
	WithWriter(w io.Writer) Table
	WithWidthFunc(f WidthFunc) Table
	WithHeaderSeparatorRow(r rune) Table
	WithPrintHeaders(b bool) Table

	AddRow(vals ...interface{}) Table
	AddRowAsArray(vals []interface{}) Table
	SetRows(rows [][]string) Table
	AddHeader(col string)
	AddHeaders(cols ...string)
	AddHeadersAsArray(cols []string)
	Print()

	// New export methods
	ExportCSV() error
	ExportJSON(keyColumn int) error

	// New sorting methods  
	SortBy(columnIndex int, cmpFn ComparisonFunc) error
	SortByMultiple(criteria []*SortCriterion) error

	// New transpose method
	Transpose() Table
}

// New creates a Table instance with the specified header(s) provided. The number
// of columns is fixed at this point to len(columnHeaders) and the defined defaults
// are set on the instance.
func New(columnHeaders ...interface{}) Table {
	t := table{header: make([]string, len(columnHeaders))}

	t.WithPadding(DefaultPadding)
	t.WithWriter(DefaultWriter)
	t.WithHeaderFormatter(DefaultHeaderFormatter)
	t.WithFirstColumnFormatter(DefaultFirstColumnFormatter)
	t.WithStandoutFormatter(DefaultStandoutFormatter)
	t.WithWidthFunc(DefaultWidthFunc)
	t.WithPrintHeaders(DefaultPrintHeaders)

	for i, col := range columnHeaders {
		t.header[i] = fmt.Sprint(col)
	}

	return &t
}

// AddHeader adds a single column header to the table.
func (t *table) AddHeader(col string) {
	t.header = append(t.header, fmt.Sprint(col))
}

// AddHeaders adds multiple column headers to the table.
func (t *table) AddHeaders(cols ...string) {
	t.header = append(t.header, cols...)
}

// AddHeadersAsArray adds an array of column headers to the table.
func (t *table) AddHeadersAsArray(cols []string) {
	for _, col := range cols {
		t.header = append(t.header, col)
	}
}

// AddRowAsArray adds a row to the table using the provided values in
// an array and returns the modified table.  Each element in the
// 'vals' slice represents a column in the row.  If the value in a
// column contains multiple lines separated by newline characters,
// each line will be placed in a separate row in the table.
func (t *table) AddRowAsArray(vals []interface{}) Table {
	maxNumNewlines := 0
	for _, val := range vals {
		maxNumNewlines = max(strings.Count(fmt.Sprint(val), "\n"), maxNumNewlines)
	}
	for i := 0; i <= maxNumNewlines; i++ {
		row := make([]string, len(t.header))
		for j, val := range vals {
			if j >= len(t.header) {
				break
			}
			v := strings.Split(fmt.Sprint(val), "\n")
			row[j] = safeOffset(v, i)
		}
		t.rows = append(t.rows, row)
	}

	return t
}

type table struct {
	FirstColumnFormatter Formatter
	HeaderFormatter      Formatter
	StandoutFormatter    Formatter
	Padding              int
	Writer               io.Writer
	Width                WidthFunc
	HeaderSeparatorRune  rune
	PrintHeaders         bool

	header []string
	rows   [][]string
	widths []int
}

func (t *table) WithHeaderFormatter(f Formatter) Table {
	t.HeaderFormatter = f
	return t
}

func (t *table) WithHeaderSeparatorRow(r rune) Table {
	t.HeaderSeparatorRune = r
	return t
}

func (t *table) WithFirstColumnFormatter(f Formatter) Table {
	t.FirstColumnFormatter = f
	return t
}

func (t *table) WithStandoutFormatter(f Formatter) Table {
	t.StandoutFormatter = f
	return t
}

func (t *table) WithPadding(p int) Table {
	if p < 0 {
		p = 0
	}

	t.Padding = p
	return t
}

func (t *table) WithWriter(w io.Writer) Table {
	if w == nil {
		w = os.Stdout
	}

	t.Writer = w
	return t
}

func (t *table) WithWidthFunc(f WidthFunc) Table {
	t.Width = f
	return t
}

func (t *table) WithPrintHeaders(b bool) Table {
	t.PrintHeaders = b
	return t
}

func (t *table) AddRow(vals ...interface{}) Table {
	return t.AddRowAsArray(vals)
}

func (t *table) SetRows(rows [][]string) Table {
	t.rows = [][]string{}
	headerLength := len(t.header)

	for _, row := range rows {
		if len(row) > headerLength {
			t.rows = append(t.rows, row[:headerLength])
		} else {
			t.rows = append(t.rows, row)
		}
	}

	return t
}

func (t *table) Print() {
	format := strings.Repeat("%s", len(t.header)) + "\n"
	t.calculateWidths()

  if t.PrintHeaders {
    t.printHeader(format)

    if t.HeaderSeparatorRune != 0 {
      t.printHeaderSeparator(format)
    }
  }

	for i, row := range t.rows {
		t.printRow(format, row, i%2 == 1)
	}
}

func (t *table) printHeaderSeparator(format string) {
	separators := make([]string, len(t.header))

	// The separator could be any unicode char. Since some chars take up more
	// than one cell in a monospace context, we can get a number higher than 1
	// here. Am example would be this emoji 🤣.
	separatorCellWidth := t.Width(string([]rune{t.HeaderSeparatorRune}))
	for index, headerName := range t.header {
		headerCellWidth := t.Width(headerName)
		// Note that this might not be evenly divisble. In this case we'll get a
		// separator that is at least 1 cell shorter than the header. This was
		// an intentional design decision in order to prevent widening the cell
		// or overstepping the column bounds.
		repeatCharTimes := headerCellWidth / separatorCellWidth
		separator := make([]rune, repeatCharTimes)
		for i := 0; i < repeatCharTimes; i++ {
			separator[i] = t.HeaderSeparatorRune
		}
		separators[index] = string(separator)
	}

	vals := t.applyWidths(separators, t.widths)
	if t.HeaderFormatter != nil {
		txt := t.HeaderFormatter(format, vals...)
		fmt.Fprint(t.Writer, txt)
	} else {
		fmt.Fprintf(t.Writer, format, vals...)
	}
}

func (t *table) printHeader(format string) {
	vals := t.applyWidths(t.header, t.widths)
	if t.HeaderFormatter != nil {
		txt := t.HeaderFormatter(format, vals...)
		fmt.Fprint(t.Writer, txt)
	} else {
		fmt.Fprintf(t.Writer, format, vals...)
	}
}

func (t *table) printRow(format string, row []string, alternate bool) {
	vals := t.applyWidths(row, t.widths)

	if t.FirstColumnFormatter != nil {
		vals[0] = t.FirstColumnFormatter("%s", vals[0])
	}
	
	if alternate && t.StandoutFormatter != nil {
		for i := range vals {
			vals[i] = t.StandoutFormatter("%s", vals[i])
		}
	}

	fmt.Fprintf(t.Writer, format, vals...)
}

func (t *table) calculateWidths() {
	t.widths = make([]int, len(t.header))
	for _, row := range t.rows {
		for i, v := range row {
			if w := t.Width(v) + t.Padding; w > t.widths[i] {
				t.widths[i] = w
			}
		}
	}

	for i, v := range t.header {
		if w := t.Width(v) + t.Padding; w > t.widths[i] {
			t.widths[i] = w
		}
	}
}

func (t *table) applyWidths(row []string, widths []int) []interface{} {
	out := make([]interface{}, len(row))
	for i, s := range row {
		out[i] = s + t.lenOffset(s, widths[i])
	}
	return out
}

func (t *table) lenOffset(s string, w int) string {
	l := w - t.Width(s)
	if l <= 0 {
		return ""
	}
	return strings.Repeat(" ", l)
}

func max(i1, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}

func safeOffset(sarr []string, idx int) string {
	if idx >= len(sarr) {
		return ""
	}
	return sarr[idx]
}

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

// SortBy sorts the table rows based on the values in the specified
// column (indexed from 0) using the provided comparison function.  It
// returns an error if the specified column index is out of bounds.
func (t *table) SortBy(columnIndex int, cmpFn ComparisonFunc) error {
	// Check if columnIndex is a valid column index
	if columnIndex < 0 || columnIndex >= len(t.header) {
		return fmt.Errorf("invalid column index %d (len %d)", columnIndex, len(t.header))
	}

	// Sort rows based on the values in the specified column using the provided comparison function
	sort.SliceStable(t.rows, func(i, j int) bool {
		return cmpFn(t.rows[i][columnIndex], t.rows[j][columnIndex]) < 0
	})

	return nil
}

// SortByMultiple sorts the table rows based on multiple sorting criteria.
// Each criterion specifies a column index and its comparison function.
// It returns an error if any specified column index is out of bounds.
func (t *table) SortByMultiple(criteria []*SortCriterion) error {
	for _, criterion := range criteria {
		// Check if each column index is valid
		if criterion.ColumnIndex < 0 || criterion.ColumnIndex >= len(t.header) {
			return fmt.Errorf("invalid column index %d (len %d)", criterion.ColumnIndex, len(t.header))
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

// Transpose the table such that the first column becomes the header and the rest of the data is transposed accordingly.
func (t *table) Transpose() Table {
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
			if i+1 < len(row) {
				newRows[i][j+1] = row[i+1]
			} else {
				newRows[i][j+1] = ""
			}
		}
	}

	// Create new table with transposed data
	newTable := &table{
		FirstColumnFormatter: t.FirstColumnFormatter,
		HeaderFormatter:      t.HeaderFormatter,
		StandoutFormatter:    t.StandoutFormatter,
		Padding:              t.Padding,
		Writer:               t.Writer,
		Width:                t.Width,
		HeaderSeparatorRune:  t.HeaderSeparatorRune,
		PrintHeaders:         t.PrintHeaders,
		header:               newHeader,
		rows:                 newRows,
	}

	return newTable
}

// Comparison functions

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
	cleanStr := strings.ReplaceAll(s, "$", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	return strconv.ParseFloat(cleanStr, 64)
}

// CurrencyComparison compares two currency values (parsed from strings) and returns -1, 0, or 1
func CurrencyComparison(a, b string) int {
	numA, errA := parseCurrencyString(a)
	numB, errB := parseCurrencyString(b)
	if errA != nil || errB != nil {
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
	cleanStr := strings.ReplaceAll(s, "%", "")
	return strconv.ParseFloat(cleanStr, 64)
}

// PercentComparison compares two percent values (parsed from strings) and returns -1, 0, or 1
func PercentComparison(a, b string) int {
	numA, errA := parsePercentString(a)
	numB, errB := parsePercentString(b)
	if errA != nil || errB != nil {
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
