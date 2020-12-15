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

// ConsensusData is the raw consensus data for the block.
type ConsensusData struct {
	Primary uint64 `json:"primary"`
	Nonce   string `json:"nonce"`
}

// Block is the raw block structure used in rpc response.
type Block struct {
	Hash              string        `json:"hash"`
	Size              int           `json:"size"`
	Version           uint          `json:"version"`
	PreviousBlockHash string        `json:"previousblockhash"`
	MerkleRoot        string        `json:"merkleroot"`
	Time              uint64        `json:"time"`
	Index             uint          `json:"index"`
	NextConsensus     string        `json:"nextconsensus"`
	Witnesses         []Witness     `json:"witnesses"`
	ConsensusData     ConsensusData `json:"consensusdata"`
	Tx                []Tx          `json:"tx"`
	NextBlockHash     string        `json:"nextblockhash"`
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
