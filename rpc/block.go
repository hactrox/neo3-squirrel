package rpc

// BlockCountResponse returns block height of chain.
type BlockCountResponse struct {
	responseCommon
	Result int `json:"result"`
}

// BlockResponse returns full block data of a specific index.
type BlockResponse struct {
	responseCommon
	Result *Block `json:"result"`
}

// Block is the raw block structure used in rpc response.
type Block struct {
	Hash              string    `json:"hash"`
	Size              int       `json:"size"`
	Version           uint      `json:"version"`
	PreviousBlockHash string    `json:"previousblockhash"`
	MerkleRoot        string    `json:"merkleroot"`
	Time              uint64    `json:"time"`
	Index             uint      `json:"index"`
	Primary           uint      `json:"primary"`
	NextConsensus     string    `json:"nextconsensus"`
	Witnesses         []Witness `json:"witnesses"`
	Tx                []Tx      `json:"tx"`
}

// SyncBlock from rpc server.
func SyncBlock(index uint) *Block {
	params := []interface{}{index, 1}
	args := generateRequestBody("getblock", params)

	respData := BlockResponse{}
	request(index, args, &respData)

	block := respData.Result
	if block != nil && block.Index > 0 {
		bestHeight.SetIfHigher(int(block.Index))
	}

	return block
}
