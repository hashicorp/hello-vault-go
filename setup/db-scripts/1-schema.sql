CREATE TABLE IF NOT EXISTS products (
   id serial PRIMARY KEY,
   name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS customers (
   id serial PRIMARY KEY,
   first_name VARCHAR(50) NOT NULL,
   last_name VARCHAR(50) NOT NULL,
   email VARCHAR(255) NOT NULL,
   phone VARCHAR(15) NOT NULL
);