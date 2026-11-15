package views

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"pomodoro/src/utils"
)

// defaultRingRadius is the ring's radius in (aspect-corrected) grid units;
// the ring occupies 2*radius+1 rows. Compact layouts pass a smaller radius.
const defaultRingRadius = 4

type ringCell struct {
	row, col int
	angle    float64 // normalized [0,1), 0 at 12 o'clock, increasing clockwise
}

// renderRing draws a circular progress indicator: elapsed segments as filled
// dots sweeping clockwise from 12 o'clock, remaining segments as hollow dots.
// centerLines (odd counts look best) are placed in the middle of the ring.
//
//	fraction ∈ [0,1] is the elapsed portion of the session.
func renderRing(fraction float64, filledStyle, trackStyle lipgloss.Style, centerLines []string) string {
	return renderRingSized(fraction, filledStyle, trackStyle, centerLines, defaultRingRadius)
}

// renderRingSized is renderRing with an explicit radius (compact layouts).
func renderRingSized(fraction float64, filledStyle, trackStyle lipgloss.Style, centerLines []string, ringRadius int) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	if ringRadius < 2 {
		ringRadius = 2
	}

	halfWidth := ringRadius * 2 // horizontal char half-span (2 chars per grid unit)
	rowCount := 2*ringRadius + 1
	colCount := 2*halfWidth + 1

	// Pass 1: collect ring cells and map them onto the character grid.
	grid := make([][]int, rowCount)
	for i := range grid {
		grid[i] = make([]int, colCount)
		for j := range grid[i] {
			grid[i][j] = -1
		}
	}
	var cells []ringCell
	for row := 0; row < rowCount; row++ {
		y := float64(row - ringRadius)
		for col := 0; col < colCount; col++ {
			x := float64(col-halfWidth) / 2.0
			if d := math.Sqrt(x*x + y*y); math.Abs(d-float64(ringRadius)) > 0.55 {
				continue
			}
			angle := math.Atan2(x, -y) / (2 * math.Pi) // ∈ [-0.5, 0.5], 0 at top
			if angle < 0 {
				angle += 1
			}
			grid[row][col] = len(cells)
			cells = append(cells, ringCell{row: row, col: col, angle: angle})
		}
	}

	// Pass 2: per row, start from dots, then clear + record the text window on
	// the rows that carry a center line.
	textStart := ringRadius - len(centerLines)/2
	if textStart < 0 {
		textStart = 0
	}

	var out strings.Builder
	for row := 0; row < rowCount; row++ {
		textCol := -1
		text := ""

		centerIdx := row - textStart
		if centerIdx >= 0 && centerIdx < len(centerLines) && centerLines[centerIdx] != "" {
			text = centerLines[centerIdx]
			runes := []rune(utils.StripANSI(text))
			textCol = (colCount - len(runes)) / 2
			if textCol < 0 {
				textCol = 0
			}
		}

		// Pass 3: emit the row with styles.
		col := 0
		textDone := false
		for col < colCount {
			if textCol >= 0 && !textDone && col == textCol {
				out.WriteString(text)
				col += len([]rune(utils.StripANSI(text)))
				textDone = true
				continue
			}
			if idx := grid[row][col]; idx >= 0 {
				if float64(cells[idx].angle) < fraction {
					out.WriteString(filledStyle.Render("●"))
				} else {
					out.WriteString(trackStyle.Render("○"))
				}
			} else {
				out.WriteString(" ")
			}
			col++
		}
		if row < rowCount-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

// spacedTime letterspaces "MM:SS" for the large center readout, e.g. "1 2 : 3 4".
func spacedTime(t string) string {
	return strings.Join(strings.Split(t, ""), " ")
}
