package rpc

// GetCommittee returns committee from fullnode.
func GetCommittee(minBlockIndex uint) *InvokeFunctionResult {
	method := "getCommittee"
	contract := "0xde5f57d430d3dece511cf975a8d37848cb9e0525"
	return InvokeFunction(minBlockIndex, contract, method, nil)
}
