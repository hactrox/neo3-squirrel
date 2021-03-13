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
	"`hash`",
	"`src`",
	"`contract`",
	"`from`",
	"`to`",
	"`amount`",
}

// InsertNEP17Transfers inserts NEP17 transfers of a transactions into DB.
func InsertNEP17Transfers(transfers []*models.Transfer,
	addrAssets []*models.AddrAsset,
	txAddrInfo map[string]*models.AddressInfo,
	newGASTotalSupply *big.Float) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		// Insert NEP17 transfers.
		if err := insertNEP17Transfer(sqlTx, transfers); err != nil {
			return err
		}

		// Update asset addresses & transfers column.
		if err := updateAssetAddressesTransfers(sqlTx, addrAssets, transfers); err != nil {
			return err
		}

		// Update balances.
		if err := updateNEP17Balances(sqlTx, addrAssets, false); err != nil {
			return err
		}

		// Update address info.
		if len(txAddrInfo) > 0 {
			if err := updateAddressInfo(sqlTx, txAddrInfo); err != nil {
				return err
			}
		}

		// Update GAS total supply if it changed.
		if newGASTotalSupply != nil {
			if err := updateContractTotalSupply(sqlTx, models.GasToken, newGASTotalSupply); err != nil {
				return err
			}
		}

		return nil
	})
}

// PersistNEP17Balances inserts and updates address contract balances.
func PersistNEP17Balances(addrAssets []*models.AddrAsset) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		updateNEP17Balances(sqlTx, addrAssets, true)
		return nil
	})
}

func updateAssetAddressesTransfers(sqlTx *sql.Tx, addrAssets []*models.AddrAsset, transfers []*models.Transfer) error {
	if len(transfers) == 0 {
		return nil
	}

	addressesChangeDelta := make(map[string]int) // map[contract]delta
	transfersChangeDelta := make(map[string]int) // map[contract]delta

	// Calculate contract holding addresses change.
	for _, addrAsset := range addrAssets {
		contract := addrAsset.Contract
		address := addrAsset.Address
		balance := addrAsset.Balance

		originBalance := getAddrAssetBalance(sqlTx, address, contract)
		zero := convert.Zero

		// New address holds this asset. addresses += 1
		if originBalance.Cmp(zero) == 0 && balance.Cmp(zero) > 0 {
			addressesChangeDelta[contract]++
		} else if originBalance.Cmp(zero) > 0 && balance.Cmp(zero) == 0 {
			addressesChangeDelta[contract]--
		}
	}

	// Calculate contract transfers change.
	for _, transfer := range transfers {
		contract := transfer.Contract
		transfersChangeDelta[contract]++
	}

	query := []string{}

	// Persist contract holding addresses change.
	for contract, delta := range addressesChangeDelta {
		if delta == 0 {
			continue
		}

		query = append(query, []string{
			"UPDATE `asset`",
			fmt.Sprintf("SET `addresses` = `addresses` + %d", delta),
			fmt.Sprintf("WHERE `contract` = '%s'", contract),
			"LIMIT 1;",
		}...)
	}

	// Persist contract transfers change.
	for contract, delta := range transfersChangeDelta {
		if delta == 0 {
			continue
		}

		query = append(query, []string{
			"UPDATE `asset`",
			fmt.Sprintf("SET `transfers` = `transfers` + %d", delta),
			fmt.Sprintf("WHERE `contract` = '%s'", contract),
			"LIMIT 1;",
		}...)
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

func insertNEP17Transfer(sqlTx *sql.Tx, transfers []*models.Transfer) error {
	if len(transfers) == 0 {
		return nil
	}

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
			transfer.Hash,
			transfer.Src,
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

func updateNEP17Balances(sqlTx *sql.Tx, addrAssets []*models.AddrAsset, updateAssetAddrDelta bool) error {
	if len(addrAssets) == 0 {
		return nil
	}

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

	// If contract asset balance not updated from transfer,
	// but from other events(e.g. Mint), then the `addresses`
	// column of table `asset` must be updated.
	addrCntDelta := map[string]int{} // map[contractHash]delta

	for _, addrAsset := range addrAssets {
		contract := addrAsset.Contract
		address := addrAsset.Address
		balance := addrAsset.Balance
		newTransfers := addrAsset.Transfers

		// Check if record already exists.
		addrAssetRec, err := getNEP17AddrAssetRecord(sqlTx, address, contract)
		if err != nil {
			return err
		}

		if updateAssetAddrDelta {
			zero := convert.Zero
			if balance.Cmp(zero) > 0 {
				if addrAssetRec == nil || addrAssetRec.Balance.Cmp(zero) == 0 {
					addrCntDelta[contract]++
				}
			} else if balance.Cmp(zero) == 0 &&
				addrAssetRec != nil && addrAssetRec.Balance.Cmp(zero) > 0 {
				addrCntDelta[contract]--
			}
		}

		if addrAssetRec == nil {
			insertsStrBuilder.WriteString(fmt.Sprintf(", ('%s', '%s', %s, %d)",
				address, contract, convert.BigFloatToString(balance), newTransfers))
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
			fmt.Sprintf("WHERE `address`='%s' AND `contract`='%s'", address, contract),
			"LIMIT 1",
		}

		updatesStrBuilder.WriteString(strings.Join(updateSQL, " ") + ";")
	}

	for contractHash, delta := range addrCntDelta {
		updateSQL := []string{
			"UPDATE `asset`",
			fmt.Sprintf("SET `addresses` = `addresses` + %d", delta),
			fmt.Sprintf("WHERE `contract`='%s'", contractHash),
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
	if len(sql) == 0 {
		return nil
	}

	_, err := sqlTx.Exec(sql)
	if err != nil {
		log.Error(sql)
		log.Panic(err)
	}

	return err
}

func getNEP17AddrAssetRecord(sqlTx *sql.Tx, address, contract string) (*models.AddrAsset, error) {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(addrAssetColumns, ", ")),
		"FROM `addr_asset`",
		fmt.Sprintf("WHERE `address` = '%s' AND `contract` = '%s'", address, contract),
		"LIMIT 1",
	}

	var addrAsset models.AddrAsset
	var balanceStr string
	err := mysql.QueryRow(mysql.Compose(query), nil,
		&addrAsset.ID,
		&addrAsset.Address,
		&addrAsset.Contract,
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

func updateContractTotalSupply(sqlTx *sql.Tx, contract string, totalSupply *big.Float) error {
	if totalSupply == nil {
		log.Panic("total supply cannot be nil")
	}

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
