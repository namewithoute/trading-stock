package response

// Standard API response messages used across all handlers.
// Using constants avoids hardcoding and ensures consistency across the API.
const (
	MsgInvalidPayload    = "invalid request payload"
	MsgUnauthorizedToken = "could not identify user from token"
	MsgInvalidToken      = "invalid or expired token"
	MsgInvalidRefresh    = "invalid or expired refresh token"
	MsgInternalError     = "internal server error"
	MsgLogoutFailed      = "failed to logout, please try again"

	// Success messages
	MsgRegisterSuccess = "user registered successfully"
	MsgLoginSuccess    = "login successful"
	MsgRefreshSuccess  = "token refreshed successfully"
	MsgLogoutSuccess   = "logout successful"
)
