package db

import (
	"database/sql"
	"fmt"
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
	"`contract`",
	"`name`",
	"`symbol`",
	"`decimals`",
	"`type`",
	"`total_supply`",
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
		var decimalsStr string
		var totalSupplyStr string

		err := rows.Scan(
			&asset.ID,
			&asset.BlockIndex,
			&asset.BlockTime,
			&asset.Contract,
			&asset.Name,
			&asset.Symbol,
			&decimalsStr,
			&asset.Type,
			&totalSupplyStr,
		)
		if err != nil {
			log.Panic(err)
		}

		asset.Decimals = convert.ToDecimal(decimalsStr)
		asset.TotalSupply = convert.ToDecimal(totalSupplyStr)

		assets = append(assets, &asset)
	}

	return assets
}

// InsertNewAsset persists new asset into DB.
func InsertNewAsset(asset *models.Asset) {
	if asset == nil {
		err := fmt.Errorf("Cannot insert nil asset into database")
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
		asset.Contract,
		asset.Name,
		asset.Symbol,
		fmt.Sprintf("%.8f", asset.Decimals),
		asset.Type,
		fmt.Sprintf("%.8f", asset.TotalSupply),
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
