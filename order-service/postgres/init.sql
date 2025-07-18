--CREATE DATABASE orders_db;
\c orders_db; 

CREATE TABLE orders (
	order_uid VARCHAR(255) PRIMARY KEY,
	track_number VARCHAR(255),
	entry VARCHAR(50),
	locale VARCHAR(10),
	internal_signature VARCHAR(255),
	customer_id VARCHAR(255),
	delivery_service VARCHAR(100),
	shardkey VARCHAR(10),
	sm_id INT,
	date_created TIMESTAMPTZ,
	oof_shard VARCHAR(10)
);

CREATE TABLE deliveries (
	order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	phone VARCHAR(50) NOT NULL,
	zip VARCHAR(50) NOT NULL,
	city VARCHAR(100) NOT NULL,
	address VARCHAR(255) NOT NULL,
	region VARCHAR(100) NOT NULL,
	email VARCHAR(100) NOT NULL
);

CREATE TABLE payments (
	order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
	transaction VARCHAR(255) NOT NULL,
	request_id VARCHAR(255),
	currency VARCHAR(10) NOT NULL,
	provider VARCHAR(100) NOT NULL,
	amount INT NOT NULL,
	payment_dt BIGINT NOT NULL,
	bank VARCHAR(100) NOT NULL,
	delivery_cost INT NOT NULL,
	goods_total INT NOT NULL,
	custom_fee INT
);

CREATE TABLE items (
	id SERIAL PRIMARY KEY,
	order_uid VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
	chrt_id BIGINT NOT NULL,
	track_number VARCHAR(255),
	price INT NOT NULL,
	rid VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL,
	sale INT NOT NULL,
	size VARCHAR(50) NOT NULL,
	total_price INT NOT NULL,
	nm_id BIGINT NOT NULL,
	brand VARCHAR(255) NOT NULL,
	status INT NOT NULL
);

--CREATE USER test_admin WITH PASSWORD 'admin';
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO test_admin;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO test_admin;
