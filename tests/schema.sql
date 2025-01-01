CREATE TABLE IF NOT EXISTS users (
                                     id INT AUTO_INCREMENT PRIMARY KEY,
                                     username VARCHAR(255) NOT NULL UNIQUE,
                                     password VARCHAR(255) NOT NULL,
                                     role VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
                                     id VARCHAR(255) PRIMARY KEY, 
                                     technician_id INT NOT NULL,
                                     summary TEXT NOT NULL,
                                     performed_at DATETIME NOT NULL,
                                     FOREIGN KEY (technician_id) REFERENCES users(id)
);