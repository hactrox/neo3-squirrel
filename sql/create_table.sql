CREATE TABLE IF NOT EXISTS `counter`
(
    `id`          INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`          INT  NOT NULL,
    `addr_count`  INT UNSIGNED  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';

INSERT INTO `counter`(`id`, `block_index`, `addr_count`) VALUES(1, -1, 0);


CREATE TABLE IF NOT EXISTS `block`
(
    `id`                  INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `hash`                    CHAR(66)  NOT NULL,
    `size`                INT UNSIGNED  NOT NULL,
    `version`             INT UNSIGNED  NOT NULL,
    `previous_block_hash`     CHAR(66)  NOT NULL,
    `merkleroot`              CHAR(66)  NOT NULL,
    `txs`                 INT UNSIGNED  NOT NULL,
    `time`             BIGINT UNSIGNED  NOT NULL,
    `index`               INT UNSIGNED  NOT NULL,
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
    `hash`                   CHAR(66)  NOT NULL,
    `size`               INT UNSIGNED  NOT NULL,
    `version`            INT UNSIGNED  NOT NULL,
    `nonce`           BIGINT UNSIGNED  NOT NULL,
    `sender`                 CHAR(34)  NOT NULL,
    `sysfee`           DECIMAL(27, 8)  NOT NULL,
    `netfee`           DECIMAL(27, 8)  NOT NULL,
    `valid_until_block`  INT UNSIGNED  NOT NULL,
    `script`               MEDIUMTEXT  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_signer`
(
    `id`            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `transaction_hash`  CHAR(66)  NOT NULL,
    `account`           CHAR(34)  NOT NULL,
    `scopes`         VARCHAR(32)  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_attribute`
(
    `id`            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `transaction_hash`  CHAR(66)  NOT NULL,
    `type`           VARCHAR(32)  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `transaction_witness`
(
    `id`            INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `transaction_hash`  CHAR(66)  NOT NULL,
    `invocation`            TEXT  NOT NULL,
    `verification`          TEXT  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `applicationlog`
(
    `id`             INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `txid`               CHAR(66)  NOT NULL,
    `trigger`         VARCHAR(16)  NOT NULL,
    `vmstate`          VARCHAR(8)  NOT NULL,
    `gasconsumed`  DECIMAL(27, 8)  NOT NULL,
    `stack`                  JSON  NOT NULL,
    `notifications`  INT UNSIGNED  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `applicationlog_notification`
(
    `id`             INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `txid`               CHAR(66)  NOT NULL,
    `block_index`    INT UNSIGNED  NOT NULL,
    `block_time`  BIGINT UNSIGNED  NOT NULL,
    `vmstate`          VARCHAR(8)  NOT NULL,
    `contract`           CHAR(42)  NOT NULL,
    `eventname`       VARCHAR(64)  NOT NULL,
    `state`                  JSON  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


-- Extra Tables


CREATE TABLE IF NOT EXISTS `transfer`
(
    `id`             INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`    INT UNSIGNED  NOT NULL,
    `block_time`  BIGINT UNSIGNED  NOT NULL,
    `txid`               CHAR(66)  NOT NULL,
    `contract`           CHAR(42)  NOT NULL,
    `from`               CHAR(34)  NOT NULL,
    `to`                 CHAR(34)  NOT NULL,
    `amount`       DECIMAL(35, 8)  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `addr_asset`
(
    `id`         INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `contract`       CHAR(42)  NOT NULL,
    `address`        CHAR(34)  NOT NULL,
    `balance`  DECIMAL(35, 8)  NOT NULL,
    `transfers`  INT UNSIGNED  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `asset`
(
    `id`              INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`     INT UNSIGNED  NOT NULL,
    `block_time`   BIGINT UNSIGNED  NOT NULL,
    `contract`            CHAR(42)  NOT NULL,
    `name`             VARCHAR(64)  NOT NULL,
    `symbol`           VARCHAR(32)  NOT NULL,
    `decimals`    TINYINT UNSIGNED  NOT NULL,
    `type`             VARCHAR(16)  NOT NULL,
    `total_supply`  DECIMAL(35, 8)  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `address`
(
    `id`                INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `address`               CHAR(34)  NOT NULL,
    `first_tx_time`  BIGINT UNSIGNED  NOT NULL,
    `last_tx_time`   BIGINT UNSIGNED  NOT NULL,
    `transfers`         INT UNSIGNED  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `contract`
(
    `id`              INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`     INT UNSIGNED  NOT NULL,
    `block_time`   BIGINT UNSIGNED  NOT NULL,
    `txid`                CHAR(66)  NOT NULL,
    `hash`                CHAR(42)  NOT NULL,
    `state`               CHAR(16)  NOT NULL,
    `new_hash`            CHAR(42)  NOT NULL,
    `contract_id`              INT  NOT NULL,
    `script`            MEDIUMTEXT  NOT NULL,
    `manifest`                JSON  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';


CREATE TABLE IF NOT EXISTS `contract_state`
(
    `id`              INT UNSIGNED  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `block_index`     INT UNSIGNED  NOT NULL,
    `block_time`   BIGINT UNSIGNED  NOT NULL,
    `txid`                CHAR(66)  NOT NULL,
    `state`               CHAR(16)  NOT NULL,
    `contract_id`              INT  NOT NULL,
    `hash`                CHAR(42)  NOT NULL,
    `name`             VARCHAR(64)  NOT NULL,
    `symbol`           VARCHAR(32)  NOT NULL,
    `decimals`    TINYINT UNSIGNED  NOT NULL,
    `total_supply`  DECIMAL(35, 8)  NOT NULL,
    `script`            MEDIUMTEXT  NOT NULL,
    `manifest`                JSON  NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = 'utf8mb4';
