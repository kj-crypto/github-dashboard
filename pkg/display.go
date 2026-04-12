package contribution

import (
	"fmt"
	"strings"
)

type DayDisplay struct {
	Empty string
	Full  string
}

var squareDayDisplay = DayDisplay{
	Empty: "□",
	Full:  "■",
}

const Width = 2*53 + 5
const Height = 8 * 2

// GitHub contribution colors using RGB values
var colorPalette = []string{
	"\x1b[38;2;235;237;240m", // Empty (light gray)
	"\x1b[38;2;155;233;168m", // Light green
	"\x1b[38;2;64;196;99m",   // Medium green
	"\x1b[38;2;47;182;125m",  // Dark green
	"\x1b[38;2;26;147;111m",  // Darkest green
}

const Reset = "\x1b[0m" // Reset color

func colorDispatcher(count uint64) string {
	if count == 1 {
		return colorPalette[1]
	}
	if count <= 3 {
		return colorPalette[2]
	}
	if count <= 5 {
		return colorPalette[3]
	}
	return colorPalette[4]
}

func FormatMonthHeader(firstRow []ContributionDay) string {
	lastMonth := firstRow[0].Month
	lastMonthPosition := 0

	output := ""
	for i, contDay := range firstRow {
		if contDay.Month != lastMonth {
			span := i - lastMonthPosition
			lastMonth = contDay.Month
			lastMonthPosition = i

			if span < 2 {
				output += strings.Repeat(" ", 2*span)
				continue
			}
			output += fmt.Sprintf("%s%s", contDay.GetPreviousMonthAbreviation(), strings.Repeat(" ", 2*span-3))
		}
	}
	if len(firstRow)-lastMonthPosition >= 2 {
		span := len(firstRow) - lastMonthPosition
		contDay := firstRow[lastMonthPosition]
		output += fmt.Sprintf("%s%s", contDay.GetMonthAbreviation(), strings.Repeat(" ", 2*span-3))
	}
	output = strings.TrimRight(output, " ")
	return output
}

func formatWeekDays(dayNo int) string {
	if dayNo == 1 {
		return "Mon  "
	}
	if dayNo == 3 {
		return "Wed  "
	}
	if dayNo == 5 {
		return "Fri  "
	}
	return strings.Repeat(" ", 5)
}

func FormatCalendar(matrix [][]ContributionDay, leftPadding uint, withWeekHeader bool) string {
	padding := strings.Repeat(" ", int(leftPadding))
	calendar := FormatMonthHeader(matrix[0]) + "\n"
	if withWeekHeader {
		calendar = padding + formatWeekDays(0) + calendar
	}

	for dayNo, row := range matrix {
		rowStr := ""
		for _, day := range row {
			color := colorDispatcher(day.ContributionCount)

			if day.ContributionCount == 0 {
				rowStr += squareDayDisplay.Empty
			} else {
				rowStr += fmt.Sprintf("%s%s%s", color, squareDayDisplay.Full, Reset)
			}
			rowStr += " "
		}
		if withWeekHeader {
			rowStr = formatWeekDays(dayNo) + rowStr
		}
		calendar += padding + strings.TrimRight(rowStr, " ") + "\n"
	}
	return strings.TrimRight(calendar, "\n")
}
