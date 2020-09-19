package db

import (
	"database/sql"
	"fmt"
	"math/big"
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
	"`contract`",
	"`from`",
	"`to`",
	"`amount`",
}

var addrAssetColumns = []string{
	"`id`",
	"`contract`",
	"`address`",
	"`balance`",
	"`transfers`",
}

// InsertNEP5Transfers inserts NEP5 transfers of a transactions into DB.
func InsertNEP5Transfers(transfers []*models.Transfer,
	addrAssets []*models.AddrAsset,
	addrTransferCntDelta map[string]*models.AddressInfo,
	newGASTotalSupply *big.Float,
	committeeGASBalances map[string]*big.Float) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		// Insert NEP5 transfers.
		if err := insertNEP5Transfer(sqlTx, transfers); err != nil {
			return err
		}

		// Update balances.
		if err := updateNEP5Balances(sqlTx, addrAssets); err != nil {
			return err
		}

		// Update address info.
		if len(addrTransferCntDelta) > 0 {
			if err := updateAddressInfo(sqlTx, addrTransferCntDelta); err != nil {
				return err
			}
		}

		// Update GAS total supply if it changed.
		if newGASTotalSupply != nil {
			if err := updateContractTotalSupply(sqlTx, models.GAS, newGASTotalSupply); err != nil {
				return err
			}
		}

		// Update committee balances.
		if len(committeeGASBalances) > 0 {
			if err := updateCommitteeGASBalances(sqlTx, committeeGASBalances); err != nil {
				return err
			}
		}

		return nil
	})
}

// PersistNEP5Balances inserts and updates address contract balances.
func PersistNEP5Balances(addrAssets []*models.AddrAsset) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		updateNEP5Balances(sqlTx, addrAssets)
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
			transfer.Contract,
			transfer.From,
			transfer.To,
			convert.BigFloatToString(transfer.Amount),
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
		newTransfers := addrAsset.Transfers

		// Check if record already exists.
		addrAssetRec, err := getNEP5AddrAssetRecord(sqlTx, address, contract)
		if err != nil {
			return err
		}

		if addrAssetRec == nil {
			insertsStrBuilder.WriteString(fmt.Sprintf(", ('%s', '%s', %s, %d)",
				contract, address, convert.BigFloatToString(balance), newTransfers))
			continue
		}

		if addrAssetRec.Balance.Cmp(balance) == 0 &&
			newTransfers == 0 {
			continue
		}

		updateSQL := []string{
			"UPDATE `addr_asset`",
			fmt.Sprintf("SET `balance`=%s", convert.BigFloatToString(balance)),
			fmt.Sprintf(", `transfers`=`transfers`+%d", newTransfers),
			fmt.Sprintf("WHERE `contract`='%s' AND `address`='%s'", contract, address),
			"LIMIT 1",
		}

		updatesStrBuilder.WriteString(strings.Join(updateSQL, " ") + ";")
	}

	sql := ""
	if insertsStrBuilder.Len() > 0 {
		sql += fmt.Sprintf("INSERT INTO `addr_asset`(%s) VALUES ", strings.Join(addrAssetColumns[1:], ", "))
		sql += insertsStrBuilder.String()[2:] + ";"
	}
	sql += updatesStrBuilder.String()

	_, err := sqlTx.Exec(sql)
	if err != nil {
		log.Error(sql)
		log.Panic(err)
	}

	return err
}

func getNEP5AddrAssetRecord(sqlTx *sql.Tx, address, contract string) (*models.AddrAsset, error) {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(addrAssetColumns, ", ")),
		"FROM `addr_asset`",
		fmt.Sprintf("WHERE `contract` = '%s' AND `address` = '%s'", contract, address),
		"LIMIT 1",
	}

	var addrAsset models.AddrAsset
	var balanceStr string
	err := mysql.QueryRow(mysql.Compose(query), nil,
		&addrAsset.ID,
		&addrAsset.Contract,
		&addrAsset.Address,
		&balanceStr,
		&addrAsset.Transfers,
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

func updateAddressInfo(sqlTx *sql.Tx, delta map[string]*models.AddressInfo) error {
	if len(delta) == 0 {
		return nil
	}

	// Sort addresses to avoid potential sql dead lock.
	addresses := []string{}
	for addr := range delta {
		addresses = append(addresses, addr)
	}

	sort.Strings(addresses)

	var insertionStrBuilder strings.Builder
	var updatesStrBuilder strings.Builder

	for _, addr := range addresses {
		firstTxTime := delta[addr].FirstTxTime
		lastTxTime := delta[addr].LastTxTime
		transfersDelta := delta[addr].Transfers

		// The new address info should be inserted.
		if firstTxTime == lastTxTime {
			insertionStrBuilder.WriteString(fmt.Sprintf(", ('%s', %d, %d, %d)",
				addr, firstTxTime, lastTxTime, transfersDelta))
			continue
		}

		// The address record should be updated.
		updateSQL := []string{
			"UPDATE `address`",
			fmt.Sprintf("SET `last_tx_time` = %d", lastTxTime),
			fmt.Sprintf(", `transfers` = `transfers` + %d", transfersDelta),
			fmt.Sprintf("WHERE `address` = '%s'", addr),
			"LIMIT 1",
		}

		updatesStrBuilder.WriteString(strings.Join(updateSQL, " ") + ";")
	}

	sql := ""
	if insertionStrBuilder.Len() > 0 {
		sql += fmt.Sprintf("INSERT INTO `address`(%s) VALUES ", strings.Join(addressInfoColumn[1:], ", "))
		sql += insertionStrBuilder.String()[2:] + ";"
	}
	sql += updatesStrBuilder.String()

	_, err := sqlTx.Exec(sql)
	if err != nil {
		log.Error(sql)
		log.Panic(err)
	}

	return err
}

func updateContractTotalSupply(sqlTx *sql.Tx, contract string, totalSupply *big.Float) error {
	query := []string{
		"UPDATE `asset`",
		fmt.Sprintf("SET `total_supply` = %s", convert.BigFloatToString(totalSupply)),
		fmt.Sprintf("WHERE `contract` = '%s'", contract),
		"LIMIT 1",
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func updateCommitteeGASBalances(sqlTx *sql.Tx, committeeBalances map[string]*big.Float) error {
	var sqlBuilder strings.Builder

	for addr, gasBalance := range committeeBalances {
		query := []string{
			"UPDATE `addr_asset`",
			fmt.Sprintf("SET `balance` = %s", convert.BigFloatToString(gasBalance)),
			fmt.Sprintf("WHERE `contract` = '%s' AND `address` = '%s'", models.GAS, addr),
			"LIMIT 1",
		}

		sqlBuilder.WriteString(mysql.Compose(query) + ";")
	}

	_, err := sqlTx.Exec(sqlBuilder.String())
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
