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

var contractStateColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`state`",
	"`contract_id`",
	"`hash`",
	"`name`",
	"`symbol`",
	"`decimals`",
	"`total_supply`",
	"`script`",
	"`manifest`",
}

var contractColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`hash`",
	"`state`",
	"`new_hash`",
	"`contract_id`",
	"`script`",
	"`manifest`",
}

// GetLastContractStateRecord returns the last record of contract state.
func GetLastContractStateRecord() *models.ContractState {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(contractStateColumns, ", ")),
		"FROM `contract_state`",
		"ORDER BY `block_index` DESC, `id` DESC",
		"LIMIT 1",
	}

	var m models.ContractState
	var totalSupplyStr string

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&m.ID,
		&m.BlockIndex,
		&m.BlockTime,
		&m.TxID,
		&m.State,
		&m.ContractID,
		&m.Hash,
		&m.Name,
		&m.Symbol,
		&m.Decimals,
		&totalSupplyStr,
		&m.Script,
		&m.Manifest,
	)

	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	m.TotalSupply = convert.ToDecimal(totalSupplyStr)

	return &m
}

// HandleContractStates inserts new contract into DB.
func HandleContractStates(contracts, added, deleted []*models.ContractState, migrated map[*models.ContractState]*models.ContractState) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := insertContracStates(sqlTx, contracts); err != nil {
			return err
		}
		if err := updateContracts(sqlTx, added, deleted, migrated); err != nil {
			return err
		}

		return nil
	})
}

func insertContracStates(sqlTx *sql.Tx, contracts []*models.ContractState) error {
	strBuilder := strings.Builder{}
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `contract_state` (%s)", strings.Join(contractStateColumns[1:], ", ")))

	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(contractStateColumns[1:]))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(contracts))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, c := range contracts {
		args = append(args,
			c.BlockIndex,
			c.BlockTime,
			c.TxID,
			c.State,
			c.ContractID,
			c.Hash,
			c.Name,
			c.Symbol,
			c.Decimals,
			convert.BigFloatToString(c.TotalSupply),
			c.Script,
			c.Manifest,
		)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func updateContracts(sqlTx *sql.Tx, added, deleted []*models.ContractState, migrated map[*models.ContractState]*models.ContractState) error {
	if err := insertNewContracts(sqlTx, added); err != nil {
		return err
	}

	if len(deleted) > 0 {
		if err := deleteContracts(sqlTx, deleted); err != nil {
			return err
		}

		if err := removeDeletedAddrAssets(sqlTx, deleted); err != nil {
			return err
		}
	}

	if len(migrated) > 0 {
		if err := migrateContracts(sqlTx, migrated); err != nil {
			return err
		}

		if err := migrateAddrAssets(sqlTx, migrated); err != nil {
			return err
		}
	}

	return nil
}

func insertNewContracts(sqlTx *sql.Tx, added []*models.ContractState) error {
	if len(added) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `contract` (%s)", strings.Join(contractColumns[1:], ", ")))

	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(contractColumns[1:]))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(added))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, c := range added {
		args = append(args,
			c.BlockIndex,
			c.BlockTime,
			c.TxID,
			c.Hash,
			c.State,
			c.NewHash,
			c.ContractID,
			c.Script,
			c.Manifest,
		)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func deleteContracts(sqlTx *sql.Tx, deleted []*models.ContractState) error {
	if len(deleted) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}

	for _, c := range deleted {
		sql := []string{
			"UPDATE `contract`",
			fmt.Sprintf("SET `state` = 'Deleted'"),
			fmt.Sprintf("WHERE `hash` = '%s'", c.Hash),
			"LIMIT 1;",
		}

		strBuilder.WriteString(strings.Join(sql, ", "))
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query)
	if err != nil {
		log.Error(err)
	}

	return err
}

func removeDeletedAddrAssets(sqlTx *sql.Tx, deleted []*models.ContractState) error {
	if len(deleted) == 0 {
		return nil
	}

	hashesParam := ""
	for _, cs := range deleted {
		hashesParam += fmt.Sprintf(", '%s'", cs.Hash)
	}
	hashesParam = hashesParam[2:]

	query := []string{
		"DELETE FROM `addr_asset`",
		fmt.Sprintf("WHERE `contract` IN (%s)", hashesParam),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

func migrateContracts(sqlTx *sql.Tx, migrates map[*models.ContractState]*models.ContractState) error {
	if len(migrates) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}

	for new, old := range migrates {
		sql := []string{
			"UPDATE `contract`",
			fmt.Sprintf("SET `state` = 'Migrated'"),
			fmt.Sprintf(", `new_hash` = '%s'", new.Hash),
			fmt.Sprintf("WHERE `hash` = '%s'", old.Hash),
			"LIMIT 1",
		}

		strBuilder.WriteString(strings.Join(sql, " "))
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query)
	if err != nil {
		log.Error(err)
	}

	return err
}

func migrateAddrAssets(sqlTx *sql.Tx, migrates map[*models.ContractState]*models.ContractState) error {
	if len(migrates) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}

	for new, old := range migrates {
		update := []string{
			"UPDATE `addr_asset`",
			fmt.Sprintf("SET `contract` = '%s'", new.Hash),
			fmt.Sprintf("WHERE `contract` = '%s'", old.Hash),
		}

		strBuilder.WriteString(fmt.Sprintf("%s;", strings.Join(update, " ")))
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query)
	if err != nil {
		log.Error(err)
	}

	return err
}
