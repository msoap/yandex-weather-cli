package main

import (
	"regexp"
	"strconv"
	"time"

	"github.com/mgutz/ansi"
)

//-----------------------------------------------------------------------------
// suggest date from one day, returns human date and json date
func suggest_date(date string, order_num int) (string, string) {
	day, err := strconv.Atoi(clear_integer_in_string(date))
	if err != nil {
		return date, date
	}

	from := time.Now().AddDate(0, 0, order_num)

	for i := 0; day != from.Day() && i < 3; i++ {
		from = from.AddDate(0, 0, 1)
	}

	weekdays_ru := [...]string{
		"вс",
		"пн",
		"вт",
		"ср",
		"чт",
		"пт",
		"сб",
	}

	return from.Format("02.01") + " (" + weekdays_ru[from.Weekday()] + ")",
		from.Format("2006-01-02")
}

//-----------------------------------------------------------------------------
// safe convert string to int, return 0 on error
func convert_str_to_int(str string) int {
	number, err := strconv.Atoi(clear_integer_in_string(str))
	if err != nil {
		return 0
	}
	return number
}

//-----------------------------------------------------------------------------
// get max length of string in slice of map of string
func get_max_length_in_slice(list []map[string]interface{}, key string) int {
	max_lengh := 0
	for _, row := range list {
		length := len([]rune(row[key].(string)))
		if max_lengh < length {
			max_lengh = length
		}
	}

	return max_lengh
}

//-----------------------------------------------------------------------------
// clear all non numeric symbols in string
func clear_integer_in_string(in string) (out string) {
	// replace dashes to minus
	out = regexp.MustCompile(string([]byte{0xE2, 0x88, 0x92})).ReplaceAllString(in, "-")

	// clear non numeric symbols
	out = regexp.MustCompile(`[^\d-]+`).ReplaceAllString(out, "")

	return out
}

//-----------------------------------------------------------------------------
// clear all non print symbols in string
func clear_nonprint_in_string(in string) (out string) {
	// replace spaces
	out = regexp.MustCompile(string([]byte{0xE2, 0x80, 0x89})).ReplaceAllString(in, " ")

	return out
}

//-----------------------------------------------------------------------------
// convert "<red>123</> str <green>456</green>" to ansi color string
// color -- color or simple remove color tags
func ansi_colour_string(str string, color bool) string {
	one_color := `(black|red|green|yellow|blue|magenta|cyan|white|grey|\d{1,3})(\+[bBuih]+)?`
	re := regexp.MustCompile(`<(` + one_color + `(:` + one_color + `)?|/\w*)>`)
	result := re.ReplaceAllStringFunc(str, func(in string) string {
		if !color {
			return ""
		}

		out := in
		tag := in[1 : len(in)-1]

		if tag[0] == '/' {
			out = ansi.ColorCode("reset")
		} else {
			out = ansi.ColorCode(tag)
		}

		return out
	})

	return result
}