package rpc

// Witness is the raw signature structure.
type Witness struct {
	Invocation   string `json:"invocation"`
	Verification string `json:"verification"`
}
