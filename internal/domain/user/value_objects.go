package user

// Status represents the user account status
type Status string

const (
	StatusActive    Status = "ACTIVE"    // User can login and trade
	StatusInactive  Status = "INACTIVE"  // User account is inactive
	StatusSuspended Status = "SUSPENDED" // User account is suspended (temporary)
	StatusBanned    Status = "BANNED"    // User account is permanently banned
)

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusSuspended, StatusBanned:
		return true
	default:
		return false
	}
}

// String returns the string representation of Status
func (s Status) String() string {
	return string(s)
}

// KYCStatus represents the KYC (Know Your Customer) verification status
type KYCStatus string

const (
	KYCPending  KYCStatus = "PENDING"  // KYC not submitted or under review
	KYCApproved KYCStatus = "APPROVED" // KYC approved
	KYCRejected KYCStatus = "REJECTED" // KYC rejected
)

// IsValid checks if the KYC status is valid
func (k KYCStatus) IsValid() bool {
	switch k {
	case KYCPending, KYCApproved, KYCRejected:
		return true
	default:
		return false
	}
}

// String returns the string representation of KYCStatus
func (k KYCStatus) String() string {
	return string(k)
}
