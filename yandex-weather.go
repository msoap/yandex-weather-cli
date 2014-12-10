/*

Command line interface for Yandex weather service (https://pogoda.yandex.ru/)

usage:
	go build yandex-weather.go

	./yandex-weather
	./yandex-weather -no-color
	./yandex-weather kiev

	# JSON out
	./yandex-weather -json london

https://github.com/msoap/yandex-weather-cli

*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/mgutz/ansi"
)

const (
	// BASE_URL - yandex pogoda service url
	BASE_URL = "https://pogoda.yandex.ru/"
	// USER_AGENT - for http.request
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/600.1.25 (KHTML, like Gecko) Version/8.0 Safari/600.1.25"
	// FORECAST_DAYS - parse days in forecast
	FORECAST_DAYS = 10
)

// SELECTORS - css selectors for forecast today
var SELECTORS = map[string]string{
	"city":       "div.navigation-city h1",
	"term_now":   "div.current-weather div.current-weather__thermometer_type_now",
	"term_night": "div.current-weather div.current-weather__thermometer_type_after",
	"desc_now":   "div.current-weather span.current-weather__comment",
	"wind":       "div.current-weather div.current-weather__info-row:nth-child(2)",
	"humidity":   "div.current-weather div.current-weather__info-row:nth-child(3)",
	"pressure":   "div.current-weather div.current-weather__info-row:nth-child(4)",
}

// SELECTORS_NEXT_DAYS - css selectors for forecast next days
var SELECTORS_NEXT_DAYS = map[string]string{
	"date":       "div.tabs-panes span.forecast-brief__item-day",
	"desc":       "div.tabs-panes div.forecast-brief__item-comment",
	"term":       "div.tabs-panes div.forecast-brief__item-temp-day",
	"term_night": "div.tabs-panes div.forecast-brief__item-temp-night",
}

//-----------------------------------------------------------------------------
// get weather html page as http.Response
func get_weather_page(city string) *http.Response {
	cookie, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookie,
	}

	weather_url := BASE_URL + city
	request, err := http.NewRequest("GET", weather_url, nil)
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("User-Agent", USER_AGENT)

	// create request for set cookies only
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	response, err = client.Get(weather_url)

	if err != nil {
		log.Fatal(err)
	}

	return response
}

//-----------------------------------------------------------------------------
// parse html via goquery, find DOM-nodes with weather forecast data
func get_weather(http_response *http.Response) (map[string]string, []map[string]string) {
	doc, err := goquery.NewDocumentFromResponse(http_response)
	if err != nil {
		log.Fatal(err)
	}

	forecast_now := map[string]string{}

	for name, selector := range SELECTORS {
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			forecast_now[name] = selection.Text()
		})
	}

	forecast_next := make([]map[string]string, 0, FORECAST_DAYS)
	for name, selector := range SELECTORS_NEXT_DAYS {
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if len(forecast_next)-1 < i {
				forecast_next = append(forecast_next, map[string]string{})
			}

			forecast_next[i][name] = selection.Text()
		})
	}

	return forecast_now, forecast_next
}

//-----------------------------------------------------------------------------
// get command line parameters
func get_params() (string, bool, bool) {
	get_json := false
	no_color := false
	flag.BoolVar(&get_json, "json", false, "get JSON")
	flag.BoolVar(&no_color, "no-color", false, "disable colored output")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [city]\noptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("\nexamples:\n  %s kiev\n  %s -json london\n", os.Args[0], os.Args[0])
	}
	flag.Parse()

	city := ""
	if flag.NArg() >= 1 {
		city = flag.Args()[0]
	}

	return city, get_json, no_color
}

//-----------------------------------------------------------------------------
// render data as text or JSON
func render(forecast_now map[string]string, forecast_next []map[string]string, city string, get_json, no_color bool) {
	if _, ok := forecast_now["city"]; ok {
		var json_data map[string]interface{}

		var (
			cl_green, cl_blue, cl_yellow, cl_reset string
		)
		if !no_color {
			cl_green = ansi.ColorCode("green")
			cl_blue = ansi.ColorCode("blue")
			cl_yellow = ansi.ColorCode("yellow")
			cl_reset = ansi.ColorCode("reset")
		}

		if get_json {
			json_data = map[string]interface{}{}
			for key, value := range forecast_now {
				json_data[key] = value
			}
		} else {
			fmt.Printf("%s (%s)\n", forecast_now["city"], cl_yellow+BASE_URL+city+cl_reset)
			fmt.Printf("Сейчас: %s, %s, ночью: %s\n",
				cl_green+forecast_now["term_now"]+cl_reset,
				cl_green+forecast_now["desc_now"]+cl_reset,
				cl_green+forecast_now["term_night"]+cl_reset,
			)
			fmt.Printf("%s\n", forecast_now["pressure"])
			fmt.Printf("%s\n", forecast_now["humidity"])
			fmt.Printf("%s\n", forecast_now["wind"])
		}

		if len(forecast_next) > 0 {
			if get_json {
				json_data["next_days"] = forecast_next
			} else {
				fmt.Printf("───────────────────────────────────────────────────────────────\n")
				fmt.Printf("%s%12s%s %s%5s%s %s%-35s%s %s%8s%s\n",
					cl_blue, "дата", cl_reset,
					cl_blue, "°C", cl_reset,
					cl_blue, "погода", cl_reset,
					cl_blue, "°C ночью", cl_reset,
				)
				fmt.Printf("───────────────────────────────────────────────────────────────\n")
				for _, row := range forecast_next {
					fmt.Printf("%12s %5s %-35s %8s\n", row["date"], row["term"], row["desc"], row["term_night"])
				}
			}
		}

		if get_json {
			json, _ := json.Marshal(json_data)
			fmt.Println(string(json))
		}
	} else {
		fmt.Printf("City \"%s\" dont found\n", city)
	}
}

//-----------------------------------------------------------------------------
func main() {
	city, get_json, no_color := get_params()
	forecast_now, forecast_next := get_weather(get_weather_page(city))
	render(forecast_now, forecast_next, city, get_json, no_color)
}
