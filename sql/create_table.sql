CREATE TABLE IF NOT EXISTS `counter`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `block_index`    INT          NOT NULL,
    `tx_pk`          INT UNSIGNED NOT NULL,
    `addr_count`     INT UNSIGNED NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';

INSERT INTO `counter`(`id`, `block_index`, `tx_pk`, `addr_count`) VALUES(1, -1, 0, 0);


CREATE TABLE IF NOT EXISTS `block`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `hash` CHAR(66) NOT NULL,
    `size` INT UNSIGNED NOT NULL,
    `version` INT UNSIGNED NOT NULL,
    `previous_block_hash` CHAR(66) NOT NULL,
    `merkleroot` CHAR(66) NOT NULL,
    `txs` INT UNSIGNED NOT NULL,
    `time` BIGINT UNSIGNED NOT NULL,
    `index` INT UNSIGNED NOT NULL,
    `nextconsensus` CHAR(66) NOT NULL,
    `consensusdata_primary` SMALLINT NOT NULL,
    `consensusdata_nonce` VARCHAR(16) NOT NULL,
    `nextblockhash` CHAR(66) NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `block_witness`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `block_hash` CHAR(66) NOT NULL,
    `invocation` TEXT NOT NULL,
    `verification` TEXT NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `block_index` INT UNSIGNED NOT NULL,
    `block_time` BIGINT UNSIGNED NOT NULL,
    `hash` CHAR(66) NOT NULL,
    `size` INT UNSIGNED NOT NULL,
    `version` INT UNSIGNED NOT NULL,
    `nonce` BIGINT UNSIGNED NOT NULL,
    `sender` CHAR(34) NOT NULL,
    `sysfee` DECIMAL(27, 8) NOT NULL,
    `netfee` DECIMAL(27, 8) NOT NULL,
    `valid_until_block` INT NOT NULL,
    `script` MEDIUMTEXT NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_signer`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `transaction_hash` CHAR(66) NOT NULL,
    `account` CHAR(34) NOT NULL,
    `scopes` VARCHAR(32) NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_attribute`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `transaction_hash` CHAR(66) NOT NULL,
    `type` VARCHAR(32) NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_witness`
(
    `id`             INT UNSIGNED auto_increment PRIMARY KEY,
    `transaction_hash` CHAR(66) NOT NULL,
    `invocation` TEXT NOT NULL,
    `verification` TEXT NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


