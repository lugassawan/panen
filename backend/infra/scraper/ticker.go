package scraper

import "strings"

// FormatIDX appends ".JK" suffix for Jakarta Stock Exchange tickers.
// Index tickers starting with "^" (e.g. ^JKSE) are returned unchanged.
func FormatIDX(ticker string) string {
	if strings.HasPrefix(ticker, "^") {
		return ticker
	}
	if strings.HasSuffix(ticker, ".JK") {
		return ticker
	}
	return ticker + ".JK"
}
