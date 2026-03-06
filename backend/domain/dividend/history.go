package dividend

import "time"

// DividendEvent represents a single historical dividend payment.
type DividendEvent struct {
	ID     string
	Ticker string
	ExDate time.Time
	Amount float64
	Source string
}
