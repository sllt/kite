package response

type Raw struct {
	Data any

	// StatusCode overrides Kite's default success HTTP status code when set to a valid HTTP status.
	// If not set (0) or invalid, Kite uses its existing status selection logic.
	StatusCode int
}
