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

func FormatMonthHeader(matrix [][]ContributionDay) string {
	first_row := matrix[0]
	last_month := first_row[0].Month
	last_month_position := 0

	var output = ""
	for i, day := range first_row {
		if day.Month != last_month {
			span := i - last_month_position
			last_month = day.Month
			last_month_position = i

			if span < 2 {
				output += strings.Repeat(" ", 2*span)
				continue
			}
			output += fmt.Sprintf("%s%s", day.GetPreviousMonthAbreviation(), strings.Repeat(" ", 2*span-3))
		}
	}
	if len(first_row)-last_month_position >= 2 {
		span := len(first_row) - last_month_position
		day := first_row[last_month_position]
		output += fmt.Sprintf("%s%s", day.GetMonthAbreviation(), strings.Repeat(" ", 2*span-3))
	}

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

func FormatCalendar(matrix [][]ContributionDay, withWeekHeader bool) string {
	calendar := FormatMonthHeader(matrix) + "\n"
	if withWeekHeader {
		calendar = formatWeekDays(0) + calendar
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
		calendar += rowStr + "\n"
	}
	return calendar
}
