package shared

// Wails event keys emitted by the backend and consumed by the frontend.
const (
	EventRefreshStatus    = "refresh:status"
	EventRefreshProgress  = "refresh:progress"
	EventRefreshSummary   = "refresh:summary"
	EventRefreshError     = "refresh:error"
	EventAlertsUpdated    = "alerts:updated"
	EventConfigChanged    = "config:changed"
	EventBrokerFeesSynced = "brokers:fees-synced"
	EventUpdateProgress   = "update:progress"
)
