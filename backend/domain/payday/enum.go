package payday

import (
	"errors"
	"fmt"
)

// Status defines the lifecycle state of a payday event.
type Status string

const (
	StatusScheduled Status = "SCHEDULED"
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusDeferred  Status = "DEFERRED"
	StatusSkipped   Status = "SKIPPED"
)

// FlowType defines the type of cash flow transaction.
type FlowType string

const (
	FlowTypeInitial  FlowType = "INITIAL"
	FlowTypeMonthly  FlowType = "MONTHLY"
	FlowTypeDividend FlowType = "DIVIDEND"
	FlowTypeSale     FlowType = "SALE"
)

// ErrInvalidStatus indicates an unrecognized payday status string.
var ErrInvalidStatus = errors.New("invalid payday status")

// ErrInvalidFlowType indicates an unrecognized flow type string.
var ErrInvalidFlowType = errors.New("invalid flow type")

// ParseStatus converts a string to a Status enum value.
func ParseStatus(s string) (Status, error) {
	switch s {
	case "SCHEDULED":
		return StatusScheduled, nil
	case "PENDING":
		return StatusPending, nil
	case "CONFIRMED":
		return StatusConfirmed, nil
	case "DEFERRED":
		return StatusDeferred, nil
	case "SKIPPED":
		return StatusSkipped, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidStatus, s)
	}
}

// ParseFlowType converts a string to a FlowType enum value.
func ParseFlowType(s string) (FlowType, error) {
	switch s {
	case "INITIAL":
		return FlowTypeInitial, nil
	case "MONTHLY":
		return FlowTypeMonthly, nil
	case "DIVIDEND":
		return FlowTypeDividend, nil
	case "SALE":
		return FlowTypeSale, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidFlowType, s)
	}
}
