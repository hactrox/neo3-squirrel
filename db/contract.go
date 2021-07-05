package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"strings"
)

var contractColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`creator`",
	"`txid`",
	"`contract_id`",
	"`hash`",
	"`state`",
	"`updatecounter`",
	"`magic`",
	"`compiler`",
	"`tokens`",
	"`script`",
	"`checksum`",
	"`name`",
	"`groups`",
	"`features`",
	"`supportedstandards`",
	"`abi`",
	"`permissions`",
	"`trusts`",
	"`extra`",
}

// InsertNativeContract inserts native contract into database.
func InsertNativeContract(contract *models.ContractState) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		return insertContract(sqlTx, contract)
	})
}

// InsertContract inserts contract state into database.
func InsertContract(contract *models.ContractState, notiPK uint, contractHash string, newAsset *models.Asset) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := insertContract(sqlTx, contract); err != nil {
			return err
		}
		// Check if this asset already been added.
		assetExists, err := assetExists(sqlTx, contractHash)
		if err != nil {
			return err
		}

		if newAsset != nil && !assetExists {
			if err := insertNewAsset(sqlTx, newAsset); err != nil {
				return err
			}
		}

		// If transfer events of this contract being handled earlier
		// than the contract task, newAsset will be persisted in
		// transfer task, the contract info may be incorrect.
		// E.g. contract creation block index may be bigger than it should be.
		if assetExists {
			if err := updateAssetDeployInfo(sqlTx,
				contract.Hash,
				contract.BlockIndex,
				contract.BlockTime,
				contract.TxID,
			); err != nil {
				return err
			}
		}

		return updateContractNotiPK(sqlTx, notiPK)
	})
}

// UpdateContract updates contract info.
func UpdateContract(contract *models.ContractState, notiPK uint, contractHash string) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := updateContract(sqlTx, contract, contractHash); err != nil {
			return err
		}

		return updateContractNotiPK(sqlTx, notiPK)
	})
}

// DeleteContract deletes contract from db.
func DeleteContract(contractHash string, notiPK uint) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := deleteContract(sqlTx, contractHash); err != nil {
			return err
		}

		return updateContractNotiPK(sqlTx, notiPK)
	})
}

func insertContract(sqlTx *sql.Tx, contract *models.ContractState) error {
	query := []string{
		"INSERT INTO `contract`",
		fmt.Sprintf("(%s)", strings.Join(contractColumns[1:], ", ")),
		fmt.Sprintf("VALUES (%s)", strings.Repeat(",?", len(contractColumns[1:]))[1:]),
	}

	// Construct sql query args.
	args := []interface{}{
		contract.BlockIndex,
		contract.BlockTime,
		contract.Creator,
		contract.TxID,
		contract.ContractID,
		contract.Hash,
		contract.State,
		contract.UpdateCounter,
		contract.NEF.Magic,
		contract.NEF.Compiler,
		contract.NEF.Tokens,
		contract.NEF.Script,
		contract.NEF.CheckSum,
		contract.Manifest.Name,
		contract.Manifest.Groups,
		contract.Manifest.Features,
		contract.MarshalSupportedStandards(),
		contract.MarshalABI(),
		contract.Manifest.Permissions,
		contract.Manifest.Trusts,
		contract.Manifest.Extra,
	}

	_, err := sqlTx.Exec(mysql.Compose(query), args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func updateContract(sqlTx *sql.Tx, contract *models.ContractState, contractHash string) error {
	query := []string{
		"UPDATE `contract` SET",
		"`block_index` = ?,",
		"`block_time` = ?,",
		"`txid` = ?,",
		"`contract_id` = ?,",
		"`hash` = ?,",
		"`state` = ?,",
		"`updatecounter` = ?,",
		"`magic` = ?,",
		"`compiler` = ?,",
		"`tokens` = ?,",
		"`script` = ?,",
		"`checksum` = ?,",
		"`name` = ?,",
		"`groups` = ?,",
		"`features` = ?,",
		"`supportedstandards` = ?,",
		"`abi` = ?,",
		"`permissions` = ?,",
		"`trusts` = ?,",
		"`extra` = ?",
		fmt.Sprintf("WHERE `hash` = '%s'", contractHash),
		"LIMIT 1",
	}

	args := []interface{}{
		contract.BlockIndex,
		contract.BlockTime,
		contract.TxID,
		contract.ContractID,
		contract.Hash,
		contract.State,
		contract.UpdateCounter,
		contract.NEF.Magic,
		contract.NEF.Compiler,
		contract.NEF.Tokens,
		contract.NEF.Script,
		contract.NEF.CheckSum,
		contract.Manifest.Name,
		contract.Manifest.Groups,
		contract.Manifest.Features,
		contract.MarshalSupportedStandards(),
		contract.MarshalABI(),
		contract.Manifest.Permissions,
		contract.Manifest.Trusts,
		contract.Manifest.Extra,
	}

	_, err := sqlTx.Exec(mysql.Compose(query), args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func deleteContract(sqlTx *sql.Tx, contractHash string) error {
	query := []string{
		"DELETE FROM `contract`",
		fmt.Sprintf("WHERE `hash` = '%s'", contractHash),
		"LIMIT 1",
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

// GetAllNativeContracts returns all Neo3 native contracts.
func GetAllNativeContracts() []*models.ContractState {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(contractColumns, ", ")),
		"FROM `contract`",
		"WHERE `contract_id` <= 0",
		"AND `block_index` = 0",
		"ORDER BY `id` ASC",
	}

	return getContractQuery(query)
}

// GetContract returns the contract of the given hash.
func GetContract(hash string) *models.ContractState {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(contractColumns, ", ")),
		"FROM `contract`",
		fmt.Sprintf("WHERE `hash` = '%s'", hash),
		"LIMIT 1",
	}

	return getContractQueryRow(query)
}

// GetLastContract returns the last contract from db.
func GetLastContract() *models.ContractState {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(contractColumns, ", ")),
		"FROM `contract`",
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	return getContractQueryRow(query)
}

/* ------------------------------
	DB query result parser
------------------------------ */

func getContractQueryRow(query []string) *models.ContractState {
	var contract models.ContractState
	supportedStandards := []byte{}
	abi := []byte{}

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&contract.ID,
		&contract.BlockIndex,
		&contract.BlockTime,
		&contract.Creator,
		&contract.TxID,
		&contract.ContractID,
		&contract.Hash,
		&contract.State,
		&contract.UpdateCounter,
		&contract.NEF.Magic,
		&contract.NEF.Compiler,
		&contract.NEF.Tokens,
		&contract.NEF.Script,
		&contract.NEF.CheckSum,
		&contract.Manifest.Name,
		&contract.Manifest.Groups,
		&contract.Manifest.Features,
		&supportedStandards,
		&abi,
		&contract.Manifest.Permissions,
		&contract.Manifest.Trusts,
		&contract.Manifest.Extra,
	)

	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	contract.UnmarshalSupportedStandards(supportedStandards)
	contract.UnmarshalABI(abi)

	return &contract
}

func getContractQuery(query []string) []*models.ContractState {
	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()

	contracts := []*models.ContractState{}

	for rows.Next() {
		var contract models.ContractState
		supportedStandards := []byte{}
		abi := []byte{}

		err := rows.Scan(
			&contract.ID,
			&contract.BlockIndex,
			&contract.BlockTime,
			&contract.Creator,
			&contract.TxID,
			&contract.ContractID,
			&contract.Hash,
			&contract.State,
			&contract.UpdateCounter,
			&contract.NEF.Magic,
			&contract.NEF.Compiler,
			&contract.NEF.Tokens,
			&contract.NEF.Script,
			&contract.NEF.CheckSum,
			&contract.Manifest.Name,
			&contract.Manifest.Groups,
			&contract.Manifest.Features,
			&supportedStandards,
			&abi,
			&contract.Manifest.Permissions,
			&contract.Manifest.Trusts,
			&contract.Manifest.Extra,
		)
		if err != nil {
			log.Panic(err)
		}

		contract.UnmarshalSupportedStandards(supportedStandards)
		contract.UnmarshalABI(abi)

		contracts = append(contracts, &contract)
	}

	return contracts
}
