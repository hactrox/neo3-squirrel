package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"sort"
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
func InsertNEP5Transfers(transfers []*models.Transfer, addrAssets []*models.AddrAsset) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		// Insert NEP5 transfers.
		if err := insertNEP5Transfer(sqlTx, transfers); err != nil {
			return err
		}

		// Update balances.
		if err := updateNEP5Balances(sqlTx, addrAssets); err != nil {
			return err
		}

		return nil
	})
}

func insertNEP5Transfer(sqlTx *sql.Tx, transfers []*models.Transfer) error {
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
}

func updateNEP5Balances(sqlTx *sql.Tx, addrAssets []*models.AddrAsset) error {
	// Sort addresses to avoid potential sql dead lock.
	sort.Slice(addrAssets, func(i, j int) bool {
		addrI := addrAssets[i].Address
		contractI := addrAssets[i].Contract
		addrJ := addrAssets[j].Address
		contractJ := addrAssets[j].Contract

		if addrI == addrJ {
			return contractI < contractJ
		}

		return addrI < addrJ
	})

	var insertsStrBuilder strings.Builder
	var updatesStrBuilder strings.Builder

	for _, addrAsset := range addrAssets {
		contract := addrAsset.Contract
		address := addrAsset.Address
		balance := addrAsset.Balance

		// Check if record already exists.
		addrAssetRec, err := getNEP5AddrAssetRecord(sqlTx, address, contract)
		if err != nil {
			return err
		}

		if addrAssetRec == nil {
			insertsStrBuilder.WriteString(fmt.Sprintf(", ('%s', '%s', %.8f)", address, contract, balance))
			continue
		}

		if addrAssetRec.Balance.Cmp(balance) == 0 {
			continue
		}

		updateSQL := []string{
			"UPDATE `addr_asset`",
			fmt.Sprintf("SET `balance`=%.8f", balance),
			fmt.Sprintf("WHERE `contract`='%s' AND `address`='%s'", contract, address),
			"LIMIT 1;",
		}

		updatesStrBuilder.WriteString(mysql.Compose(updateSQL))
	}

	sql := ""
	if insertsStrBuilder.Len() > 0 {
		sql += "INSERT INTO `addr_asset`(`address`, `contract`, `balance`) VALUES "
		sql += insertsStrBuilder.String()[2:] + ";"
	}
	sql += updatesStrBuilder.String()

	_, err := sqlTx.Exec(sql)
	if err != nil {
		log.Error(sql)
		log.Error(err)
	}

	return err
}

func getNEP5AddrAssetRecord(sqlTx *sql.Tx, address, contract string) (*models.AddrAsset, error) {
	columns := []string{
		"`id`",
		"`address`",
		"`contract`",
		"`balance`",
	}

	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(columns, ", ")),
		"FROM `addr_asset`",
		fmt.Sprintf("WHERE `contract` = '%s' AND `address` = '%s'", contract, address),
		"LIMIT 1",
	}

	var addrAsset models.AddrAsset
	var balanceStr string
	err := mysql.QueryRow(mysql.Compose(query), nil,
		&addrAsset.ID,
		&addrAsset.Address,
		&addrAsset.Contract,
		&balanceStr,
	)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil, nil
		}

		log.Error(err)
		return nil, err
	}

	addrAsset.Balance = convert.ToDecimal(balanceStr)
	return &addrAsset, nil
}
