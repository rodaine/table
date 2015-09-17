package table

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	DefaultPadding = 2
	DefaultWriter  = os.Stdout
)

type Formatter func(string, ...interface{}) string

type Table interface {
	WithHeaderColumnFormatter(Formatter) Table
	WithFirstColumnFormatter(Formatter) Table
	WithPadding(int) Table
	WithWriter(io.Writer) Table

	AddRow(...interface{}) Table
	Print()
}

type table struct {
	FirstColumnFormatter Formatter
	HeaderFormatter      Formatter
	Padding              int
	Writer               io.Writer

	header []string
	rows   [][]string
	widths []int
}

func New(columnHeaders ...interface{}) Table {
	t := table{
		HeaderFormatter:      nil,
		FirstColumnFormatter: nil,
		Padding:              DefaultPadding,
		Writer:               DefaultWriter,

		header: make([]string, len(columnHeaders)),
		widths: make([]int, len(columnHeaders)),
	}

	for i, col := range columnHeaders {
		t.header[i] = fmt.Sprint(col)
	}

	return &t
}

func (t *table) WithHeaderColumnFormatter(f Formatter) Table {
	t.HeaderFormatter = f
	return t
}

func (t *table) WithFirstColumnFormatter(f Formatter) Table {
	t.FirstColumnFormatter = f
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
	t.Writer = w
	return t
}

func (t *table) AddRow(vals ...interface{}) Table {
	row := make([]string, len(t.header))
	for i, val := range vals {
		if i >= len(t.header) {
			break
		}
		row[i] = fmt.Sprint(val)
	}
	t.rows = append(t.rows, row)

	return t
}

func (t *table) Print() {
	format := strings.Repeat("%s", len(t.header)) + "\n"
	t.calculateWidths()
	fmt.Fprintln(t.Writer)
	t.printHeader(format)
	for _, row := range t.rows {
		t.printRow(format, row)
	}
}

func (t *table) printHeader(format string) {
	vals := applyWidths(t.header, t.widths)
	if t.HeaderFormatter != nil {
		txt := t.HeaderFormatter(format, vals...)
		fmt.Fprint(t.Writer, txt)
	} else {
		fmt.Fprintf(t.Writer, format, vals...)
	}
}

func (t *table) printRow(format string, row []string) {
	vals := applyWidths(row, t.widths)

	if t.FirstColumnFormatter != nil {
		vals[0] = t.FirstColumnFormatter("%s", row[0])
	}

	fmt.Fprintf(t.Writer, format, vals...)
}

func (t *table) calculateWidths() {
	t.widths = make([]int, len(t.widths))
	for _, row := range t.rows {
		for i, v := range row {
			if w := len(v) + t.Padding; w > t.widths[i] {
				t.widths[i] = w
			}
		}
	}

	for i, v := range t.header {
		if w := len(v) + t.Padding; w > t.widths[i] {
			t.widths[i] = w
		}
	}
}

func applyWidths(row []string, widths []int) []interface{} {
	out := make([]interface{}, len(row))
	for i, s := range row {
		out[i] = s + lenOffset(s, widths[i])
	}
	return out
}

func lenOffset(s string, w int) string {
	l := w - len(s)
	if l <= 0 {
		return ""
	}
	return strings.Repeat(" ", l)
}
