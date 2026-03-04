package checklist

import "fmt"

// ActionType defines the trading action for a checklist evaluation.
type ActionType string

const (
	ActionBuy         ActionType = "BUY"
	ActionAverageDown ActionType = "AVERAGE_DOWN"
	ActionAverageUp   ActionType = "AVERAGE_UP"
	ActionSellExit    ActionType = "SELL_EXIT"
	ActionSellStop    ActionType = "SELL_STOP"
	ActionHold        ActionType = "HOLD"
)

// CheckType defines whether a check is automatic or manual.
type CheckType string

const (
	CheckTypeAuto   CheckType = "AUTO"
	CheckTypeManual CheckType = "MANUAL"
)

// CheckStatus defines the result status of a check.
type CheckStatus string

const (
	CheckStatusPass    CheckStatus = "PASS"
	CheckStatusFail    CheckStatus = "FAIL"
	CheckStatusPending CheckStatus = "PENDING"
)

// ParseActionType converts a string to an ActionType enum value.
func ParseActionType(s string) (ActionType, error) {
	switch s {
	case "BUY":
		return ActionBuy, nil
	case "AVERAGE_DOWN":
		return ActionAverageDown, nil
	case "AVERAGE_UP":
		return ActionAverageUp, nil
	case "SELL_EXIT":
		return ActionSellExit, nil
	case "SELL_STOP":
		return ActionSellStop, nil
	case "HOLD":
		return ActionHold, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidAction, s)
	}
}
