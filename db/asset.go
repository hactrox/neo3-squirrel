package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
)

var assetColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`contract`",
	"`name`",
	"`symbol`",
	"`decimals`",
	"`total_supply`",
	"`addresses`",
	"`transfers`",
}

var addrAssetColumns = []string{
	"`id`",
	"`contract`",
	"`address`",
	"`balance`",
	"`transfers`",
}

// GetAllAssets returns all assets from DB.
func GetAllAssets() []*models.Asset {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(assetColumns, ", ")),
		"FROM `asset`",
	}

	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()

	assets := []*models.Asset{}
	for rows.Next() {
		asset := models.Asset{}
		var totalSupplyStr string

		err := rows.Scan(
			&asset.ID,
			&asset.BlockIndex,
			&asset.BlockTime,
			&asset.TxID,
			&asset.Contract,
			&asset.Name,
			&asset.Symbol,
			&asset.Decimals,
			&totalSupplyStr,
			&asset.Addresses,
			&asset.Transfers,
		)
		if err != nil {
			log.Panic(err)
		}

		asset.TotalSupply = convert.ToDecimal(totalSupplyStr)

		assets = append(assets, &asset)
	}

	return assets
}

// GetAddrAssetBalance returns asset balance of the given address.
func GetAddrAssetBalance(addr, assetHash string) *big.Float {
	query := []string{
		"SELECT `balance`",
		"FROM `addr_asset`",
		fmt.Sprintf("WHERE `address`='%s'", addr),
		fmt.Sprintf("AND `contract`='%s'", assetHash),
		"LIMIT 1",
	}

	var balanceStr string
	err := mysql.QueryRow(mysql.Compose(query), nil, &balanceStr)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return convert.Zero
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	return convert.ToDecimal(balanceStr)
}

// GetAsset returns the asset info of the given hash.
func GetAsset(assetHash string) *models.Asset {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(assetColumns, ", ")),
		"FROM `asset`",
		fmt.Sprintf("WHERE `contract` = '%s'", assetHash),
		"LIMIT 1",
	}

	var asset models.Asset
	var totalSupplyStr string

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&asset.ID,
		&asset.BlockIndex,
		&asset.BlockTime,
		&asset.TxID,
		&asset.Contract,
		&asset.Name,
		&asset.Symbol,
		&asset.Decimals,
		&totalSupplyStr,
		&asset.Addresses,
		&asset.Transfers,
	)

	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Panic(err)
	}

	asset.TotalSupply = convert.ToDecimal(totalSupplyStr)

	return &asset
}

// DestroyAsset deletes asset and its related data.
func DestroyAsset(assetHash string) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := deleteAssetTransfers(sqlTx, assetHash); err != nil {
			return err
		}

		if err := deleteAsset(sqlTx, assetHash); err != nil {
			return err
		}

		if err := deleteAssetAddrBalances(sqlTx, assetHash); err != nil {
			return err
		}

		return nil
	})
}

func deleteAssetTransfers(sqlTx *sql.Tx, assetHash string) error {
	query := []string{
		"DELETE FROM `transfer`",
		fmt.Sprintf("WHERE `contract` = '%s'", assetHash),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

func deleteAsset(sqlTx *sql.Tx, assetHash string) error {
	query := []string{
		"DELETE FROM `asset`",
		fmt.Sprintf("WHERE `contract` = '%s'", assetHash),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

func deleteAssetAddrBalances(sqlTx *sql.Tx, assetHash string) error {
	query := []string{
		"DELETE FROM `addr_asset`",
		fmt.Sprintf("WHERE `contract` = '%s'", assetHash),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

// InsertNewAsset persists new asset into DB.
func InsertNewAsset(asset *models.Asset) {
	if asset == nil {
		err := fmt.Errorf("cannot insert nil asset into database")
		log.Panic(err)
	}

	mysql.Trans(func(sqlTx *sql.Tx) error {
		// Check if this asset already been added.
		exists, err := contractExists(sqlTx, asset.Contract)
		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		return insertNewAsset(sqlTx, asset)
	})
}

func contractExists(sqlTx *sql.Tx, contract string) (bool, error) {
	query := []string{
		"SELECT EXISTS(",
		"SELECT `id`",
		"FROM `asset`",
		fmt.Sprintf("WHERE `contract` = '%s'", contract),
		"LIMIT 1)",
	}

	var exists bool
	err := sqlTx.QueryRow(mysql.Compose(query)).Scan(&exists)
	if err != nil {
		log.Error(err)
		return false, err
	}

	return exists, err
}

func insertNewAsset(sqlTx *sql.Tx, asset *models.Asset) error {
	query := []string{
		"INSERT INTO `asset`",
		fmt.Sprintf("(%s)", strings.Join(assetColumns[1:], ", ")),
		fmt.Sprintf("VALUES (%s)", strings.Repeat(",?", len(assetColumns[1:]))[1:]),
	}

	args := []interface{}{
		asset.BlockIndex,
		asset.BlockTime,
		asset.TxID,
		asset.Contract,
		asset.Name,
		asset.Symbol,
		asset.Decimals,
		convert.BigFloatToString(asset.TotalSupply),
		asset.Addresses,
		asset.Transfers,
	}

	result, err := sqlTx.Exec(mysql.Compose(query), args...)
	if err != nil {
		if mysql.IsDuplicateEntryError(err) {
			return nil
		}

		log.Panic(err)
	}

	mysql.CheckIfRowsNotAffected(result, query)
	return nil
}
