CREATE TABLE IF NOT EXISTS account_transaction
(
    user_id VARCHAR(36),
    operation_type VARCHAR(36),
    order_id VARCHAR(36),
    amount NUMERIC(9,2),
    created_at TIMESTAMP without time zone DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS usr
(
    user_id VARCHAR(36) PRIMARY KEY,
    login VARCHAR(36),
    password VARCHAR(36)
);
CREATE TABLE IF NOT EXISTS orders
(
    order_id  VARCHAR(36) PRIMARY KEY,
    user_id   VARCHAR(36),
    status VARCHAR(36),
    amount NUMERIC(5,2),
    created_at TIMESTAMP without time zone DEFAULT NOW(),
    updated_at TIMESTAMP without time zone DEFAULT NOW()
    );