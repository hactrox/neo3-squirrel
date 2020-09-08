package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"strings"
)

var transferColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`from`",
	"`to`",
	"`amount`",
}

// InsertNEP5Transfers inserts NEP5 transfers of a transactions into DB.
func InsertNEP5Transfers(transfers []*models.Transfer) {
	mysql.Trans(func(sqlTx *sql.Tx) error {

		var strBuilder strings.Builder
		strBuilder.WriteString(fmt.Sprintf("INSERT INTO `transfer` (%s)", strings.Join(transferColumns[1:], ", ")))

		strBuilder.WriteString("VALUES")

		// Construct (?, ?, ?) list.
		statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(transferColumns[1:]))[1:])
		strBuilder.WriteString(strings.Repeat(statement, len(transfers))[1:])

		// Construct sql query args.
		args := []interface{}{}
		for _, transfer := range transfers {
			args = append(args,
				transfer.BlockIndex,
				transfer.BlockTime,
				transfer.TxID,
				transfer.From,
				transfer.To,
				fmt.Sprintf("%.8f", transfer.Amount),
			)
		}

		query := strBuilder.String()
		_, err := sqlTx.Exec(query, args...)
		if err != nil {
			log.Error(err)
		}

		return err
	})
}
