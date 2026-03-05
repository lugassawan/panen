package crashplaybook

// ComputeResponseLevels calculates 3 response tiers based on entry price and 52-week low.
// deployPcts defines the deployment percentage for each level [normalDip, crash, extreme].
func ComputeResponseLevels(entryPrice, low52Week float64, deployPcts [3]float64) []ResponseLevel {
	normalDip := entryPrice
	crash := (entryPrice + low52Week) / 2
	extreme := low52Week * 1.05

	return []ResponseLevel{
		{Level: LevelNormalDip, TriggerPrice: normalDip, DeployPct: deployPcts[0]},
		{Level: LevelCrash, TriggerPrice: crash, DeployPct: deployPcts[1]},
		{Level: LevelExtreme, TriggerPrice: extreme, DeployPct: deployPcts[2]},
	}
}

// DetermineActiveLevel returns the deepest crash level that has been triggered,
// or nil if price is above all trigger levels.
func DetermineActiveLevel(price float64, levels []ResponseLevel) *CrashLevel {
	var active *CrashLevel
	for i := range levels {
		if price <= levels[i].TriggerPrice {
			lvl := levels[i].Level
			active = &lvl
		}
	}
	return active
}
