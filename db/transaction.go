package db

import (
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
)

var txColumns = []string{
	"`id`",
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

// GetLastTransaction returns the last transaction record.
func GetLastTransaction() *models.Transaction {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(txColumns, ", ")),
		"FROM `transaction`",
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	return getTransactionQueryRow(query)
}

// GetTxCount return the number of txs starts
// from the given primary key(>=startPK).
func GetTxCount(startPK uint) uint {
	query := []string{
		"SELECT COUNT(`id`)",
		"FROM `transaction`",
		fmt.Sprintf("WHERE `id` >= %d", startPK),
	}

	var count uint
	err := mysql.QueryRow(mysql.Compose(query), nil, &count)
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	return count
}

// GetBlockTxs gets transactions of the given block index.
func GetBlockTxs(blockIndex uint) []*models.Transaction {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(txColumns, ", ")),
		"FROM `transaction`",
		fmt.Sprintf("WHERE `block_index` = %d", blockIndex),
	}

	return getTransactionsQuery(query)
}

// GetTransaction returns transaction by txID.
func GetTransaction(txID string) *models.Transaction {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(txColumns, ", ")),
		"FROM `transaction`",
		fmt.Sprintf("WHERE `hash` = '%s'", txID),
		"LIMIT 1",
	}

	return getTransactionQueryRow(query)
}

/* ------------------------------
    DB query result parser
------------------------------ */

func getTransactionQueryRow(query []string) *models.Transaction {
	var tx models.Transaction
	var sysFee string
	var netFee string

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&tx.ID,
		&tx.BlockIndex,
		&tx.BlockTime,
		&tx.Hash,
		&tx.Size,
		&tx.Version,
		&tx.Nonce,
		&tx.Sender,
		&sysFee,
		&netFee,
		&tx.ValidUntilBlock,
		&tx.Script,
	)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	tx.SysFee = convert.ToDecimal(sysFee)
	tx.NetFee = convert.ToDecimal(netFee)

	return &tx
}

func getTransactionsQuery(query []string) []*models.Transaction {
	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()

	txs := []*models.Transaction{}

	for rows.Next() {
		var tx models.Transaction
		var sysFee string
		var netFee string

		err := rows.Scan(
			&tx.ID,
			&tx.BlockIndex,
			&tx.BlockTime,
			&tx.Hash,
			&tx.Size,
			&tx.Version,
			&tx.Nonce,
			&tx.Sender,
			&sysFee,
			&netFee,
			&tx.ValidUntilBlock,
			&tx.Script,
		)
		if err != nil {
			log.Panic(err)
		}

		tx.SysFee = convert.ToDecimal(sysFee)
		tx.NetFee = convert.ToDecimal(netFee)

		txs = append(txs, &tx)
	}

	return txs
}
