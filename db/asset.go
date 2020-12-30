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
	"`type`",
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

// GetAllAssets returns all asset of the given type from DB.
func GetAllAssets(typ string) []*models.Asset {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(assetColumns, ", ")),
		"FROM `asset`",
	}

	if typ != "" && strings.ToLower(typ) != "all" {
		query = append(query, fmt.Sprintf("WHERE `type` = '%s'", typ))
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
			&asset.Type,
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
func GetAddrAssetBalance(addr, contract string) *big.Float {
	query := []string{
		"SELECT `balance`",
		"FROM `addr_asset`",
		fmt.Sprintf("WHERE `address`='%s' AND `contract`='%s'", addr, contract),
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
func GetAsset(hash string) *models.Asset {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(assetColumns, ", ")),
		"FROM `asset`",
		fmt.Sprintf("WHERE `contract` = '%s'", hash),
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
		&asset.Type,
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
		asset.Type,
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

func deleteAsset(sqlTx *sql.Tx, hashes []string) error {
	if len(hashes) == 0 {
		return nil
	}

	hashesParam := ""
	for _, hash := range hashes {
		hashesParam += fmt.Sprintf(", '%s'", hash)
	}
	hashesParam = hashesParam[2:]

	query := []string{
		"DELETE FROM `asset`",
		fmt.Sprintf("WHERE `contract` IN (%s)", hashesParam),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}
