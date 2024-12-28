-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role ENUM('manager', 'technician') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    summary TEXT CHECK (CHAR_LENGTH(summary) <= 2500),
    performed_date TIMESTAMP NOT NULL,
    technician_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (technician_id) REFERENCES users(id),
    INDEX idx_technician (technician_id),
    INDEX idx_performed_date (performed_date)
);

-- Insert test data only if the users table is empty
INSERT INTO users (username, password, role)
SELECT * FROM (
    SELECT 'manager1', '$2a$10$test_hash_replace_this', 'manager'
    UNION ALL
    SELECT 'tech1', '$2a$10$test_hash_replace_this', 'technician'
    UNION ALL
    SELECT 'tech2', '$2a$10$test_hash_replace_this', 'technician'
) AS tmp
WHERE NOT EXISTS (
    SELECT 1 FROM users
) LIMIT 1;
