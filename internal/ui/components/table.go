package components

import (
	"fmt"
	"strings"

	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// TableCol describes a column in a text table.
type TableCol struct {
	Header string
	Width  int
}

// TableRow is a single row of string values.
type TableRow []string

// TextTable renders a simple aligned text table.
// It accepts column definitions and a slice of rows.
func TextTable(cols []TableCol, rows []TableRow) string {
	if len(cols) == 0 {
		return ""
	}

	// Build header
	var header strings.Builder
	for i, col := range cols {
		header.WriteString(theme.TableHeaderStyle.Width(col.Width).Render(fmt.Sprintf("%-*s", col.Width, col.Header)))
		if i < len(cols)-1 {
			header.WriteString(" ")
		}
	}
	header.WriteString("\n")

	// Build rows
	var body strings.Builder
	for _, row := range rows {
		for i, cell := range row {
			w := cols[i].Width
			if i >= len(cols) {
				break
			}
			body.WriteString(theme.NormalStyle.Width(w).Render(fmt.Sprintf("%-*s", w, cell)))
			if i < len(row)-1 && i < len(cols)-1 {
				body.WriteString(" ")
			}
		}
		body.WriteString("\n")
	}

	return header.String() + body.String()
}
