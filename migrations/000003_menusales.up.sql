-- Table for the menu and sales relationship using postgresql
-- File: migrations/000003_menusales.up.sql
CREATE TABLE IF NOT EXISTS menu_sales (
    id SERIAL PRIMARY KEY,
    menu_id INT NOT NULL,
    sale_id INT NOT NULL,
    quantity INT NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL,
    last_modified_by VARCHAR(100) NOT NULL,
    FOREIGN KEY (menu_id) REFERENCES menu(id) ON DELETE CASCADE,
    FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);