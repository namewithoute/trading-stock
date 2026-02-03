package risk

import "errors"

// LimitStatus represents the status of risk limits
type LimitStatus string

const (
	LimitStatusActive    LimitStatus = "ACTIVE"
	LimitStatusInactive  LimitStatus = "INACTIVE"
	LimitStatusSuspended LimitStatus = "SUSPENDED"
)

// IsValid checks if the limit status is valid
func (ls LimitStatus) IsValid() bool {
	switch ls {
	case LimitStatusActive, LimitStatusInactive, LimitStatusSuspended:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (ls LimitStatus) String() string {
	return string(ls)
}

// AlertType represents the type of risk alert
type AlertType string

const (
	AlertTypePositionLimit      AlertType = "POSITION_LIMIT_EXCEEDED"
	AlertTypeOrderLimit         AlertType = "ORDER_LIMIT_EXCEEDED"
	AlertTypeLossLimit          AlertType = "LOSS_LIMIT_EXCEEDED"
	AlertTypeLeverageLimit      AlertType = "LEVERAGE_LIMIT_EXCEEDED"
	AlertTypeConcentrationLimit AlertType = "CONCENTRATION_LIMIT_EXCEEDED"
	AlertTypeMarginCall         AlertType = "MARGIN_CALL"
	AlertTypeHighRisk           AlertType = "HIGH_RISK_SCORE"
	AlertTypeSuspiciousActivity AlertType = "SUSPICIOUS_ACTIVITY"
)

// IsValid checks if the alert type is valid
func (at AlertType) IsValid() bool {
	switch at {
	case AlertTypePositionLimit, AlertTypeOrderLimit, AlertTypeLossLimit,
		AlertTypeLeverageLimit, AlertTypeConcentrationLimit, AlertTypeMarginCall,
		AlertTypeHighRisk, AlertTypeSuspiciousActivity:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (at AlertType) String() string {
	return string(at)
}

// Severity represents the severity level of an alert
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// IsValid checks if the severity is valid
func (s Severity) IsValid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (s Severity) String() string {
	return string(s)
}

// AlertStatus represents the status of a risk alert
type AlertStatus string

const (
	AlertStatusActive   AlertStatus = "ACTIVE"
	AlertStatusResolved AlertStatus = "RESOLVED"
	AlertStatusIgnored  AlertStatus = "IGNORED"
)

// IsValid checks if the alert status is valid
func (as AlertStatus) IsValid() bool {
	switch as {
	case AlertStatusActive, AlertStatusResolved, AlertStatusIgnored:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (as AlertStatus) String() string {
	return string(as)
}

// Domain errors
var (
	ErrPositionLimitExceeded      = errors.New("position limit exceeded")
	ErrOrderLimitExceeded         = errors.New("order limit exceeded")
	ErrLossLimitExceeded          = errors.New("loss limit exceeded")
	ErrLeverageLimitExceeded      = errors.New("leverage limit exceeded")
	ErrConcentrationLimitExceeded = errors.New("concentration limit exceeded")
	ErrRiskLimitNotFound          = errors.New("risk limit not found")
	ErrHighRiskScore              = errors.New("risk score too high")
	ErrMarginCallRequired         = errors.New("margin call required")
)
