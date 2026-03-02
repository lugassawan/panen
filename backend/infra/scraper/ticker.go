package scraper

import "strings"

// FormatIDX appends ".JK" suffix for Jakarta Stock Exchange tickers.
func FormatIDX(ticker string) string {
	if strings.HasSuffix(ticker, ".JK") {
		return ticker
	}
	return ticker + ".JK"
}
