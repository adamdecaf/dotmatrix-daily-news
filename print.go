package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Constants
var LOCATION = [2]float64{28.5383, -81.3792} // Lat + Lng for Orlando FL

var WMOCODES = map[int]string{
	0:  "Clear",
	1:  "Mostly Clear",
	2:  "Partly Cloudy",
	3:  "Overcast",
	45: "Foggy",
	48: "Foggy",
	51: "Light Drizzle",
	53: "Drizzle",
	55: "Heavy Drizzle",
	56: "Frz Drizzle",
	57: "Frz Drizzle",
	61: "Slight Rain",
	63: "Rain",
	65: "Heavy Rain",
	66: "Light Frz Rain",
	67: "Heavy Frz Rain",
	71: "Light Snow",
	73: "Snow",
	75: "Heavy Snow",
	77: "Snow",
	80: "Light Showers",
	81: "Showers",
	82: "Heavy Showers",
	85: "Snow Showers",
	86: "Heavy Snow Showers",
	95: "Thunderstorms",
	96: "Hail Storms",
	99: "Heavy Hail Storms",
}

var STOCKS = []string{"DIA", "SPY"}

var STOCKSURL = "https://api.twelvedata.com/quote"
var STOCKSKEY = os.Getenv("STOCKS_API_KEY")
var NEWSURL = "https://api.nytimes.com/svc/mostpopular/v2/viewed/1.json"
var NEWSKEY = os.Getenv("NEWS_API_KEY")

const MAXNEWS = 3

var SUBREDDITS = []string{"science", "upliftingnews", "technology", "fauxmoi", "todayilearned"}

// Main function
func main() {
	var buf bytes.Buffer

	// Fetch weather data
	fmt.Println("Fetching weather data...")
	weatherUrl := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&daily=weather_code,temperature_2m_max,temperature_2m_min,sunrise,sunset,daylight_duration,wind_speed_10m_max&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=America/New_York&forecast_days=1", LOCATION[0], LOCATION[1])

	weatherData := fetchData(weatherUrl)
	if weatherData == nil {
		fmt.Println("Unable to retrieve weather data")
		return
	}

	// Fetch stock ticker data
	stockData := map[string]map[string]interface{}{}
	if STOCKSKEY != "" {
		fmt.Println("Fetching stock ticker data...")
		for _, stock := range STOCKS {
			stockData[stock] = fetchData(fmt.Sprintf("%s?symbol=%s&apikey=%s", STOCKSURL, stock, STOCKSKEY))
			if stockData[stock] == nil {
				fmt.Printf("Unable to retrieve stock data for %s\n", stock)
				return
			}
		}
	}

	// Fetch news headlines data
	var newsData map[string]interface{}
	if NEWSKEY != "" {
		fmt.Println("Fetching news headlines data...")
		newsUrl := fmt.Sprintf("%s?api-key=%s", NEWSURL, NEWSKEY)
		newsData = fetchData(newsUrl)
		if newsData == nil || len(newsData["results"].([]interface{})) == 0 {
			fmt.Println("Unable to retrieve news data")
			return
		}
	}

	// Fetch reddit top posts data
	fmt.Println("Fetching reddit top posts data...")
	redditData := map[string]interface{}{}
	for _, subreddit := range SUBREDDITS {
		redditUrl := fmt.Sprintf("https://reddit.com/r/%s.json", subreddit)
		data := fetchData(redditUrl)
		if data == nil {
			fmt.Printf("Unable to fetch reddit data for r/%s\n", subreddit)
			return
		}
		children := data["data"].(map[string]interface{})["children"].([]interface{})
		sort.Slice(children, func(i, j int) bool {
			return children[i].(map[string]interface{})["data"].(map[string]interface{})["ups"].(float64) > children[j].(map[string]interface{})["data"].(map[string]interface{})["ups"].(float64)
		})
		redditData[subreddit] = children[0]
	}

	// Print header (simulating writing to printer)
	fmt.Println("Writing to printer...")
	buf.WriteString(printHeader() + "\n")
	buf.WriteString(printWeather(weatherData) + "\n")
	if len(stockData) > 0 {
		buf.WriteString(printMarkets(stockData) + "\n")
	}
	if len(newsData) > 0 {
		buf.WriteString(printNews(newsData) + "\n")
	}
	buf.WriteString(printReddit(redditData) + "\n")
	buf.WriteString(printFooter())

	cmd := exec.Command("lp", "-d", "Canon_TS3500_series")
	cmd.Stdin = &buf
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		fmt.Printf("problem printing")
		return
	}
}

// Helper functions
func fetchData(url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}
	return result
}

func printHeader() string {
	var buf bytes.Buffer
	date := time.Now().Format("Mon Jan 2 2006")
	buf.WriteString("SCHMELYUN TRIBUNE" + strings.Repeat(" ", 20) + date + "\n")
	buf.WriteString(strings.Repeat("-", 78) + "\n")
	return buf.String()
}

func printFooter() string {
	var buf bytes.Buffer
	buf.WriteString(strings.Repeat("-", 78) + "\n")

	date := time.Now().Format("Mon Jan 2 2006")
	buf.WriteString("SCHMELYUN TRIBUNE" + strings.Repeat(" ", 20) + date + "\n")

	return buf.String()
}

func printWeather(weatherData map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("WEATHER\n")
	daily := weatherData["daily"].(map[string]interface{})
	weatherCode := int(daily["weather_code"].([]interface{})[0].(float64))
	buf.WriteString(fmt.Sprintf("    %s - High: %.1f°F, Low: %.1f°F\n\n",
		WMOCODES[weatherCode],
		daily["temperature_2m_max"].([]interface{})[0].(float64),
		daily["temperature_2m_min"].([]interface{})[0].(float64),
	))
	return buf.String()
}

func printMarkets(stockData map[string]map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("MARKETS\n\n")
	for stock, data := range stockData {
		closePrice := data["close"].(float64)
		percentChange := data["percent_change"].(float64)
		buf.WriteString(fmt.Sprintf("  %s: %.2f (%.2f%%)\n\n", stock, closePrice, percentChange))
	}
	return buf.String()
}

func printNews(newsData map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("HEADLINES\n\n")
	results := newsData["results"].([]interface{})
	for i, article := range results {
		if i >= MAXNEWS {
			break
		}
		buf.WriteString(fmt.Sprintf("  %s\n\n", article.(map[string]interface{})["title"].(string)))
	}
	return buf.String()
}

func printReddit(redditData map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString("REDDIT\n\n")
	for subreddit, item := range redditData {
		title := item.(map[string]interface{})["data"].(map[string]interface{})["title"].(string)
		ups := int(item.(map[string]interface{})["data"].(map[string]interface{})["ups"].(float64))
		buf.WriteString(fmt.Sprintf("  r/%s - %s (Upvotes: %d)\n\n", subreddit, title, ups))
	}
	return buf.String()
}
