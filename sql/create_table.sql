CREATE TABLE IF NOT EXISTS `counter`
(
    `id`                INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`                INT  NOT NULL DEFAULT 0,
    `addr_count`        INT UNSIGNED  NOT NULL DEFAULT 0,
    `contract_noti_pk`  INT UNSIGNED  NOT NULL DEFAULT 0,
    `tx_count`          INT UNSIGNED  NOT NULL DEFAULT 0
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';

INSERT INTO `counter`(`id`, `block_index`)
VALUES(1, -1);


CREATE TABLE IF NOT EXISTS `block`
(
    `id`                  INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `hash`                    CHAR(66)  NOT NULL UNIQUE,
    `size`                INT UNSIGNED  NOT NULL,
    `version`             INT UNSIGNED  NOT NULL,
    `previous_block_hash`     CHAR(66)  NOT NULL,
    `merkleroot`              CHAR(66)  NOT NULL,
    `txs`                 INT UNSIGNED  NOT NULL,
    `time`             BIGINT UNSIGNED  NOT NULL,
    `index`               INT UNSIGNED  NOT NULL UNIQUE,
    `nextconsensus`           CHAR(66)  NOT NULL,
    `consensusdata_primary`   SMALLINT  NOT NULL,
    `consensusdata_nonce`  VARCHAR(16)  NOT NULL,
    `nextblockhash`           CHAR(66)  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `block_witness`
(
    `id`      INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_hash`  CHAR(66)  NOT NULL,
    `invocation`      TEXT  NOT NULL,
    `verification`    TEXT  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction`
(
    `id`                 INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`        INT UNSIGNED  NOT NULL,
    `block_time`      BIGINT UNSIGNED  NOT NULL,
    `hash`                   CHAR(66)  NOT NULL UNIQUE,
    `size`               INT UNSIGNED  NOT NULL,
    `version`            INT UNSIGNED  NOT NULL,
    `nonce`           BIGINT UNSIGNED  NOT NULL,
    `sender`                 CHAR(34)  NOT NULL,
    `sysfee`           DECIMAL(24, 8)  NOT NULL,
    `netfee`           DECIMAL(24, 8)  NOT NULL,
    `valid_until_block`  INT UNSIGNED  NOT NULL,
    `script`               MEDIUMTEXT  NOT NULL,

    INDEX `idx_hash` (`hash`),
    INDEX `idx_block_index` (`block_index`)
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_signer`
(
    `id`            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `transaction_hash`  CHAR(66)  NOT NULL,
    `account`           CHAR(42)  NOT NULL,
    `scopes`         VARCHAR(32)  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_attribute`
(
    `id`            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `transaction_hash`  CHAR(66)  NOT NULL,
    `body`                  JSON  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_witness`
(
    `id`            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `transaction_hash`  CHAR(66)  NOT NULL,
    `invocation`            TEXT  NOT NULL,
    `verification`          TEXT  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `notification`
(
    `id`             INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`    INT UNSIGNED  NOT NULL,
    `block_time`  BIGINT UNSIGNED  NOT NULL,
    `hash`               CHAR(66)  NOT NULL,
    `src`                 CHAR(5)  NOT NULL,
    `exec_idx`  SMALLINT UNSIGNED  NOT NULL,
    `trigger`         VARCHAR(16)  NOT NULL,
    `vmstate`          VARCHAR(8)  NOT NULL,
    `exception`      VARCHAR(512)  NOT NULL,
    `gasconsumed`  DECIMAL(24, 8)  NOT NULL,
    `stack`                  JSON  NOT NULL,
    `n`         SMALLINT UNSIGNED  NOT NULL,
    `contract`           CHAR(42)  NOT NULL,
    `eventname`       VARCHAR(64)  NOT NULL,
    `state`                  JSON  NOT NULL,

    INDEX `uix_hash` (`hash`)
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `contract_notification`
(
    `id`             INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`    INT UNSIGNED  NOT NULL,
    `block_time`  BIGINT UNSIGNED  NOT NULL,
    `hash`               CHAR(66)  NOT NULL,
    `src`                 CHAR(5)  NOT NULL,
    `exec_idx`  SMALLINT UNSIGNED  NOT NULL,
    `trigger`         VARCHAR(16)  NOT NULL,
    `vmstate`          VARCHAR(8)  NOT NULL,
    `exception`      VARCHAR(512)  NOT NULL,
    `gasconsumed`  DECIMAL(24, 8)  NOT NULL,
    `stack`                  JSON  NOT NULL,
    `n`         SMALLINT UNSIGNED  NOT NULL,
    `contract`           CHAR(42)  NOT NULL,
    `eventname`       VARCHAR(64)  NOT NULL,
    `state`                  JSON  NOT NULL,

    INDEX `uix_hash` (`hash`)
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


-- Extra Tables


CREATE TABLE IF NOT EXISTS `transfer`
(
    `id`              INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`     INT UNSIGNED  NOT NULL,
    `block_time`   BIGINT UNSIGNED  NOT NULL,
    `hash`                CHAR(66)  NOT NULL,
    `src`                  CHAR(5)  NOT NULL,
    `contract`            CHAR(42)  NOT NULL,
    `from`                CHAR(34)  NOT NULL,
    `to`                  CHAR(34)  NOT NULL,
    `amount`       DECIMAL(65, 30)  NOT NULL,

    INDEX `idx_contract` (`contract`),
    INDEX `idx_hash` (`hash`),
    INDEX `idx_from` (`from`),
    INDEX `idx_to` (`to`)
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `addr_asset`
(
    `id`          INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `contract`        CHAR(42)  NOT NULL,
    `address`         CHAR(34)  NOT NULL,
    `balance`  DECIMAL(65, 30)  NOT NULL,
    `transfers`   INT UNSIGNED  NOT NULL,

    INDEX `idx_contract` (`contract`),
    INDEX `idx_address` (`address`)
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `asset`
(
    `id`               INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`      INT UNSIGNED  NOT NULL,
    `block_time`    BIGINT UNSIGNED  NOT NULL,
    `txid`                 CHAR(66)  NOT NULL,
    `contract`             CHAR(42)  NOT NULL UNIQUE,
    `name`              VARCHAR(64)  NOT NULL,
    `symbol`            VARCHAR(32)  NOT NULL,
    `decimals`     TINYINT UNSIGNED  NOT NULL,
    `total_supply`  DECIMAL(65, 30)  NOT NULL,
    `addresses`        INT UNSIGNED  NOT NULL,
    `transfers`        INT UNSIGNED  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `address`
(
    `id`                INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `address`               CHAR(34)  NOT NULL UNIQUE,
    `first_tx_time`  BIGINT UNSIGNED  NOT NULL,
    `last_tx_time`   BIGINT UNSIGNED  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `contract`
(
    `id`              INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`     INT UNSIGNED  NOT NULL,
    `block_time`   BIGINT UNSIGNED  NOT NULL,
    `creator`             CHAR(34)  NOT NULL,
    `txid`                CHAR(66)  NOT NULL,
    `contract_id`              INT  NOT NULL,
    `hash`                CHAR(42)  NOT NULL UNIQUE,
    `state`               CHAR(16)  NOT NULL,
    `updatecounter`   INT UNSIGNED  NOT NULL,
    `script`            MEDIUMTEXT  NOT NULL,
    `name`            VARCHAR(256)  NOT NULL,
    `groups`                  JSON  NOT NULL,
    `supportedstandards`      JSON  NOT NULL,
    `abi`                     JSON  NOT NULL,
    `permissions`             JSON  NOT NULL,
    `trusts`                  JSON  NOT NULL,
    `extra`                   JSON  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';
