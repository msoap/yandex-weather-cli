// utility functions
package main

import (
	"regexp"
	"strconv"
	"time"

	"github.com/mgutz/ansi"
)

// HistoChars - chars for draw histogram
var HistoChars = [...]string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

var weekdaysRu = [...]string{
	"вс",
	"пн",
	"вт",
	"ср",
	"чт",
	"пт",
	"сб",
}

//-----------------------------------------------------------------------------
// formatDates gets date in json and human format
func formatDates(date time.Time) (formatDate string, jsonDate string) {
	return date.Format("02.01") + " (" + weekdaysRu[date.Weekday()] + ")",
		date.Format("2006-01-02")
}

//-----------------------------------------------------------------------------
// safe convert string to int, return 0 on error
func convertStrToInt(str string) int {
	number, err := strconv.Atoi(clearIntegerInString(str))
	if err != nil {
		return 0
	}
	return number
}

//-----------------------------------------------------------------------------
func getMaxLengthDesc(list []DayForecast) int {
	maxLengh := 0
	for _, row := range list {
		length := len([]rune(row.Desc))
		if maxLengh < length {
			maxLengh = length
		}
	}

	return maxLengh
}

//-----------------------------------------------------------------------------
// clear all non numeric symbols in string
func clearIntegerInString(in string) (out string) {
	// replace dashes to minus
	out = regexp.MustCompile(string([]byte{0xE2, 0x88, 0x92})).ReplaceAllString(in, "-")

	// clear non numeric symbols
	out = regexp.MustCompile(`[^\d-]+`).ReplaceAllString(out, "")

	return out
}

//-----------------------------------------------------------------------------
// clear all non print symbols in string
func clearNonprintInString(in string) (out string) {
	// replace spaces
	out = regexp.MustCompile(string([]byte{0xE2, 0x80, 0x89})).ReplaceAllString(in, " ")

	return out
}

//-----------------------------------------------------------------------------
// convert "<red>123</> str <green>456</green>" to ansi color string
func (cfg Config) ansiColourString(str string) string {
	oneColor := `(black|red|green|yellow|blue|magenta|cyan|white|grey|\d{1,3})(\+[bBuih]+)?`
	re := regexp.MustCompile(`<(` + oneColor + `(:` + oneColor + `)?|/\w*)>`)
	result := re.ReplaceAllStringFunc(str, func(in string) (out string) {
		if cfg.noColor {
			return ""
		}

		if tag := in[1 : len(in)-1]; tag[0] == '/' {
			out = ansi.ColorCode("reset")
		} else {
			out = ansi.ColorCode(tag)
		}

		return out
	})

	return result
}

// ----------------------------------------------------------------------------
// Render histogram for forecast by hours
func renderHisto(forecastByHours []HourTemp) string {
	// linear interpolation (* 4)
	interpolationFact := 4
	temperatures := make([]float64, len(forecastByHours)*interpolationFact)
	for i, row := range forecastByHours {
		currTemp := float64(row.Temp)
		nextI := i + 1
		if i == len(forecastByHours)-1 {
			nextI = i
		}
		nextTemp := float64(forecastByHours[nextI].Temp)

		temperatures[i*interpolationFact] = currTemp

		for j := 1; j < interpolationFact; j++ {
			temperatures[i*interpolationFact+j] = currTemp +
				(float64(j)/float64(interpolationFact))*((nextTemp-currTemp)/1)
		}
	}

	minTemp, maxTemp := temperatures[0], temperatures[0]
	result := ""

	for _, temp := range temperatures {
		if minTemp > temp {
			minTemp = temp
		}
		if maxTemp < temp {
			maxTemp = temp
		}
	}

	maxGradation := float64(len(HistoChars) - 1)
	if maxTemp-minTemp < maxGradation/2 {
		// if difference between max and min is too small
		maxTemp = minTemp + maxGradation/2
	}
	for _, temp := range temperatures {
		reduceValue := int((temp - minTemp) / (maxTemp - minTemp) * maxGradation)
		result += HistoChars[reduceValue]
	}

	return result
}
