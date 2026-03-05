package dividend

import "sort"

// ScoreInput holds the parameters for computing a dividend attractiveness score.
type ScoreInput struct {
	DY             float64 // current dividend yield
	MinDY          float64 // minimum acceptable yield (threshold)
	PayoutRatio    float64 // current payout ratio
	MaxPayoutRatio float64 // max acceptable payout ratio (threshold)
	Price          float64 // current price
	EntryPrice     float64 // entry target price from valuation
	PositionPct    float64 // current portfolio weight
	MaxPositionPct float64 // max portfolio weight (threshold)
}

// Score computes a dividend attractiveness score (higher = more attractive).
//
// Components:
//   - Yield premium: how much DY exceeds the minimum (0–40 pts)
//   - Payout margin: headroom below max payout ratio (0–20 pts)
//   - Valuation bonus: discount to entry price (0–20 pts)
//   - Weight headroom: room to add before hitting max position (0–20 pts)
func Score(input ScoreInput) float64 {
	var score float64

	// Yield premium (0–40 pts): reward higher yields
	if input.MinDY > 0 && input.DY >= input.MinDY {
		premium := (input.DY - input.MinDY) / input.MinDY
		score += clampMax(premium*40, 40)
	}

	// Payout margin (0–20 pts): reward lower payout ratios
	if input.MaxPayoutRatio > 0 && input.PayoutRatio < input.MaxPayoutRatio {
		margin := (input.MaxPayoutRatio - input.PayoutRatio) / input.MaxPayoutRatio
		score += clampMax(margin*20, 20)
	}

	// Valuation bonus (0–20 pts): reward discount to entry price
	if input.EntryPrice > 0 && input.Price > 0 && input.Price <= input.EntryPrice {
		discount := (input.EntryPrice - input.Price) / input.EntryPrice
		score += clampMax(discount*20, 20)
	}

	// Weight headroom (0–20 pts): reward room to add
	if input.MaxPositionPct > 0 && input.PositionPct < input.MaxPositionPct {
		headroom := (input.MaxPositionPct - input.PositionPct) / input.MaxPositionPct
		score += clampMax(headroom*20, 20)
	}

	return score
}

// Rank sorts items by score in descending order.
func Rank(items []RankItem) []RankItem {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	return items
}

func clampMax(v, hi float64) float64 {
	if v < 0 {
		return 0
	}
	if v > hi {
		return hi
	}
	return v
}
