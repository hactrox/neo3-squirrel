package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"strings"
)

var blockColumns = []string{
	"`id`",
	"`hash`",
	"`size`",
	"`version`",
	"`previous_block_hash`",
	"`merkleroot`",
	"`txs`",
	"`time`",
	"`index`",
	"`nextconsensus`",
	"`consensusdata_primary`",
	"`consensusdata_nonce`",
	"`nextblockhash`",
}

// InsertBlock inserts bulk blocks into database.
func InsertBlock(blocks []*models.Block, txBulk *models.TxBulk) {
	if len(blocks) == 0 {
		return
	}

	insertBlocksCmd := generateInsertCmdForBlocks(blocks)
	insertBlockWitnessesCmd := generateInsertCmdForBlockWitnesses(blocks)
	insertTxsCmd := generateInsertCmdForTxs(txBulk.Txs)
	insertTxSignersCmd := generateInsertCmdForTxSigners(txBulk.TxSigners)
	insertTxAttrsCmd := generateInsertCmdForTxAttrs(txBulk.TxAttrs)
	insertTxWitnessCmd := generateInsertCmdForTxWitnesses(txBulk.TxWitnesses)

	cmds := []string{
		insertBlocksCmd,
		insertBlockWitnessesCmd,
		insertTxsCmd,
		insertTxSignersCmd,
		insertTxAttrsCmd,
		insertTxWitnessCmd,
	}

	mysql.Trans(func(sqlTx *sql.Tx) error {
		for _, cmd := range cmds {
			if cmd == "" {
				continue
			}
			if _, err := sqlTx.Exec(cmd); err != nil {
				log.Error(err)
				return err
			}
		}

		err := updateBlockIndexCounter(sqlTx, blocks[len(blocks)-1].Index)
		if err != nil {
			return err
		}

		return updateTxCounter(sqlTx, len(txBulk.Txs))
	})
}

// GetBlock returns block record from db.
func GetBlock(index uint) *models.Block {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(blockColumns, ", ")),
		"FROM `block`",
		fmt.Sprintf("WHERE `index` = %d", index),
		"LIMIT 1",
	}

	block := models.Block{}

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&block.ID,
		&block.Hash,
		&block.Size,
		&block.Version,
		&block.PreviousBlockHash,
		&block.MerkleRoot,
		&block.Txs,
		&block.Time,
		&block.Index,
		&block.NextConsensus,
		&block.ConsensusDataPrimary,
		&block.ConsensusDataNonce,
		&block.NextBlockHash,
	)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	return &block
}

func generateInsertCmdForBlocks(blocks []*models.Block) string {
	if len(blocks) == 0 {
		return ""
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `block` (%s) VALUES ", strings.Join(blockColumns[1:], ", ")))

	for _, b := range blocks {
		strBuilder.WriteString(fmt.Sprintf("('%s', %d, %d, '%s', '%s', %d, %d, %d, '%s', %d, '%s', '%s'),",
			b.Hash,
			b.Size,
			b.Version,
			b.PreviousBlockHash,
			b.MerkleRoot,
			b.Txs,
			b.Time,
			b.Index,
			b.NextConsensus,
			b.ConsensusDataPrimary,
			b.ConsensusDataNonce,
			b.NextBlockHash,
		))
	}

	return strings.TrimSuffix(strBuilder.String(), ",")
}

func generateInsertCmdForBlockWitnesses(blocks []*models.Block) string {
	if len(blocks) == 0 {
		return ""
	}

	columns := []string{
		"`block_hash`",
		"`invocation`",
		"`verification`",
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `block_witness` (%s) VALUES ", strings.Join(columns, ",")))

	for _, b := range blocks {
		for _, witness := range b.Witnesses {
			strBuilder.WriteString(fmt.Sprintf("('%s', '%s', '%s'),",
				b.Hash,
				witness.Invocation,
				witness.Verification,
			))
		}
	}

	return strings.TrimSuffix(strBuilder.String(), ",")
}

func generateInsertCmdForTxs(txs []*models.Transaction) string {
	if len(txs) == 0 {
		return ""
	}

	columns := []string{
		"`block_index`",
		"`block_time`",
		"`hash`",
		"`size`",
		"`version`",
		"`nonce`",
		"`sender`",
		"`sysfee`",
		"`netfee`",
		"`valid_until_block`",
		"`script`",
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `transaction` (%s) VALUES ", strings.Join(columns, ", ")))

	for _, tx := range txs {
		strBuilder.WriteString(fmt.Sprintf("(%d, %d, '%s', %d, %d, %d, '%s', %.8f, %.8f, %d, '%s'),",
			tx.BlockIndex,
			tx.BlockTime,
			tx.Hash,
			tx.Size,
			tx.Version,
			tx.Nonce,
			tx.Sender,
			tx.SysFee,
			tx.NetFee,
			tx.ValidUntilBlock,
			tx.Script,
		))
	}

	return strings.TrimSuffix(strBuilder.String(), ",")
}

func generateInsertCmdForTxSigners(signers []*models.TransactionSigner) string {
	if len(signers) == 0 {
		return ""
	}

	columns := []string{
		"`transaction_hash`",
		"`account`",
		"`scopes`",
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `transaction_signer` (%s) VALUES", strings.Join(columns, ", ")))

	for _, signer := range signers {
		strBuilder.WriteString(fmt.Sprintf("('%s', '%s', '%s'),",
			signer.TransactionHash,
			signer.Account,
			signer.Scopes,
		))
	}

	return strings.TrimSuffix(strBuilder.String(), ",")
}

func generateInsertCmdForTxAttrs(attrs []*models.TransactionAttribute) string {
	if len(attrs) == 0 {
		return ""
	}

	columns := []string{
		"`transaction_hash`",
		"`body`",
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `transaction_attribute` (%s) VALUES", strings.Join(columns, ", ")))

	for _, attr := range attrs {
		strBuilder.WriteString(fmt.Sprintf("('%s', '%s'),",
			attr.TransactionHash,
			attr.Body,
		))
	}

	return strings.TrimSuffix(strBuilder.String(), ",")
}

func generateInsertCmdForTxWitnesses(witnesses []*models.TransactionWitness) string {
	if len(witnesses) == 0 {
		return ""
	}

	columns := []string{
		"`transaction_hash`",
		"`invocation`",
		"`verification`",
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `transaction_witness` (%s) VALUES", strings.Join(columns, ", ")))

	for _, witness := range witnesses {
		strBuilder.WriteString(fmt.Sprintf("('%s', '%s', '%s'),",
			witness.TransactionHash,
			witness.Invocation,
			witness.Verification,
		))
	}

	return strings.TrimSuffix(strBuilder.String(), ",")
}
