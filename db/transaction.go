package db

import (
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
)

// GetTransactions gets transactions starts from the given
// primary key(>=) and limits to {limit} records from database.
func GetTransactions(startPK, limit uint) []*models.Transaction {
	columns := []string{
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

	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(columns, ", ")),
		"FROM `transaction`",
		"WHERE `id` >= ?",
		"LIMIT ?",
	}

	rows, err := mysql.Query(mysql.Compose(query), startPK, limit)
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

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
