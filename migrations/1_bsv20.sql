CREATE TABLE progress(
    indexer VARCHAR(32) PRIMARY KEY,
    height INTEGER
);

CREATE TABLE txns(
    txid BYTEA PRIMARY KEY,
	block_id BYTEA,
    height INTEGER,
    idx BIGINT,
    fees BIGINT,
    feeacc BIGINT,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_txns_block_id_idx ON txns(block_id, idx);
CREATE INDEX idx_txns_created_unmined ON txns(created)
    WHERE height IS NULL;
    
CREATE TABLE IF NOT EXISTS bsv20 (
    txid BYTEA,
    vout INT,
    height INT,
    idx BIGINT,
    tick TEXT,
    max NUMERIC NOT NULL,
    lim NUMERIC NOT NULL,
    dec INT DEFAULT 18,
    supply NUMERIC DEFAULT 0,
    status INT DEFAULT 0,
    available NUMERIC GENERATED ALWAYS AS (max - supply) STORED,
    pct_minted NUMERIC GENERATED ALWAYS AS (CASE WHEN max = 0 THEN 0 ELSE ROUND(100.0 * supply / max, 1) END) STORED,
    reason TEXT,
    PRIMARY KEY(txid, vout)
);
-- CREATE INDEX IF NOT EXISTS idx_bsv20_tick ON bsv20(tick);
-- CREATE INDEX IF NOT EXISTS idx_bsv20_available ON bsv20(available);
-- CREATE INDEX IF NOT EXISTS idx_bsv20_pct_minted ON bsv20(pct_minted);
-- CREATE INDEX IF NOT EXISTS idx_bsv20_max ON bsv20(max);
-- CREATE INDEX IF NOT EXISTS idx_bsv20_height_idx ON bsv20(height, idx, vout);
-- CREATE INDEX IF NOT EXISTS idx_bsv20_to_validate ON bsv20(height, idx, vout)
--     WHERE status = 0;

CREATE TABLE bsv20_txos(
    txid BYTEA,
    vout INTEGER,
    height INTEGER,
    idx BIGINT,
    script BYTEA,
    amt NUMERIC,
    outacc BIGINT,
    address TEXT,
    spend BYTEA,
    price BIGINT,
    payout BYTEA,
    status INT DEFAULT 0,
    PRIMARY KEY(txid, vout)
);
