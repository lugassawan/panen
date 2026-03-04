package checklist

import (
	"fmt"
	"math"
)

// Suggestion holds a computed trade suggestion.
type Suggestion struct {
	Action          ActionType
	Ticker          string
	Lots            int
	PricePerShare   float64
	GrossCost       float64
	Fee             float64
	Tax             float64
	NetCost         float64
	NewAvgBuyPrice  float64
	NewPositionLots int
	NewPositionPct  float64
	CapitalGainPct  float64
}

// ErrHoldNoSuggestion is returned when the action is Hold and no trade is needed.
var ErrHoldNoSuggestion = fmt.Errorf("%w: no trade needed for hold action", ErrChecklistNotReady)

// ComputeSuggestion computes a trade suggestion based on the evaluation input.
// Returns ErrHoldNoSuggestion for Hold action (no trade needed).
func ComputeSuggestion(input EvaluateInput) (*Suggestion, error) {
	switch input.Action {
	case ActionBuy, ActionAverageDown, ActionAverageUp:
		return computeBuySuggestion(input)
	case ActionSellExit, ActionSellStop:
		return computeSellSuggestion(input)
	case ActionHold:
		return nil, ErrHoldNoSuggestion
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidAction, input.Action)
	}
}

func computeBuySuggestion(input EvaluateInput) (*Suggestion, error) {
	data := input.StockData
	holding := input.Holding

	if (input.Action == ActionAverageDown || input.Action == ActionAverageUp) && holding == nil {
		return nil, ErrNoHolding
	}

	maxPositionValue := input.Portfolio.Capital * input.Thresholds.MaxPositionPct / 100

	var currentPositionValue float64
	if holding != nil {
		currentPositionValue = data.Price * float64(holding.Lots) * 100
	}

	available := maxPositionValue - currentPositionValue
	if available <= 0 {
		return nil, fmt.Errorf("%w: no room to buy (position at max)", ErrChecklistNotReady)
	}

	lots := int(math.Floor(available / (data.Price * 100)))
	if lots == 0 {
		return nil, fmt.Errorf("%w: insufficient budget for 1 lot", ErrChecklistNotReady)
	}

	shares := float64(lots) * 100
	grossCost := data.Price * shares
	fee := grossCost * input.BuyFeePct / 100
	netCost := grossCost + fee

	var newAvgBuyPrice float64
	var newPositionLots int

	if holding != nil {
		existingCost := holding.AvgBuyPrice * float64(holding.Lots)
		newCost := data.Price * float64(lots)
		totalLots := holding.Lots + lots
		newAvgBuyPrice = (existingCost + newCost) / float64(totalLots)
		newPositionLots = totalLots
	} else {
		newAvgBuyPrice = data.Price
		newPositionLots = lots
	}

	newPositionValue := data.Price * float64(newPositionLots) * 100
	totalPortfolioValue := computeTotalPortfolioValue(input, newPositionValue)
	var newPositionPct float64
	if totalPortfolioValue > 0 {
		newPositionPct = (newPositionValue / totalPortfolioValue) * 100
	}

	return &Suggestion{
		Action:          input.Action,
		Ticker:          data.Ticker,
		Lots:            lots,
		PricePerShare:   data.Price,
		GrossCost:       grossCost,
		Fee:             fee,
		Tax:             0,
		NetCost:         netCost,
		NewAvgBuyPrice:  newAvgBuyPrice,
		NewPositionLots: newPositionLots,
		NewPositionPct:  newPositionPct,
		CapitalGainPct:  0,
	}, nil
}

func computeSellSuggestion(input EvaluateInput) (*Suggestion, error) {
	holding := input.Holding
	if holding == nil {
		return nil, ErrNoHolding
	}

	data := input.StockData
	shares := float64(holding.Lots) * 100
	grossCost := data.Price * shares
	fee := grossCost * input.SellFeePct / 100
	tax := grossCost * input.SellTaxPct / 100
	netCost := grossCost - fee - tax

	buyCost := holding.AvgBuyPrice * shares * (1 + input.BuyFeePct/100)
	capitalGainPct := ((netCost - buyCost) / buyCost) * 100

	return &Suggestion{
		Action:          input.Action,
		Ticker:          data.Ticker,
		Lots:            holding.Lots,
		PricePerShare:   data.Price,
		GrossCost:       grossCost,
		Fee:             fee,
		Tax:             tax,
		NetCost:         netCost,
		NewAvgBuyPrice:  0,
		NewPositionLots: 0,
		NewPositionPct:  0,
		CapitalGainPct:  capitalGainPct,
	}, nil
}

func computeTotalPortfolioValue(input EvaluateInput, newPositionValue float64) float64 {
	var total float64
	ticker := input.StockData.Ticker
	found := false

	for _, h := range input.AllHoldings {
		if h.Ticker == ticker {
			total += newPositionValue
			found = true
		} else {
			total += h.AvgBuyPrice * float64(h.Lots) * 100
		}
	}

	if !found {
		total += newPositionValue
	}

	return total
}
