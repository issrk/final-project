-- MySQL schema for POS System
-- Create the database and application tables.

CREATE DATABASE IF NOT EXISTS posdb;
USE posdb;

CREATE TABLE IF NOT EXISTS products (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  price DOUBLE NOT NULL,
  stock INT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  product_id INT NOT NULL,
  quantity INT NOT NULL,
  amount DOUBLE NOT NULL,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE IF NOT EXISTS receipts (
  id INT PRIMARY KEY AUTO_INCREMENT,
  items TEXT NOT NULL,
  total DOUBLE NOT NULL,
  discount DOUBLE DEFAULT 0,
  tax DOUBLE DEFAULT 0,
  net_total DOUBLE NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT IGNORE INTO products (name, price, stock) VALUES
  ('Coffee', 100.00, 50),
  ('Tea', 80.00, 50),
  ('Sandwich', 150.00, 50),
  ('Juice', 60.00, 50);
