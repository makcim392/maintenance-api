-- Users table
CREATE TABLE IF NOT EXISTS users (
                                     id         int auto_increment primary key,
                                     username   varchar(255) not null,
                                     password   varchar(255) not null,
                                     role       enum ('manager', 'technician') not null,
                                     created_at timestamp default CURRENT_TIMESTAMP null,
                                     updated_at timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP,
                                     constraint username unique (username)
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
                                     id             varchar(36) primary key,
                                     summary        text null,
                                     performed_date timestamp not null,
                                     technician_id  int not null,
                                     created_at     timestamp default CURRENT_TIMESTAMP null,
                                     updated_at     timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP,
                                     constraint tasks_ibfk_1
                                         foreign key (technician_id) references users (id),
                                     check (char_length(`summary`) <= 2500)
);

-- Indexes
CREATE INDEX idx_performed_date ON tasks (performed_date);
CREATE INDEX idx_technician ON tasks (technician_id);

-- Insert users if table is empty
INSERT INTO users (id, username, password, role, created_at, updated_at)
SELECT * FROM (
                  SELECT 1, 'manager1', '$2a$10$test_hash_replace_this', 'manager', '2024-12-29 19:05:18', '2024-12-29 19:05:18' UNION ALL
                  SELECT 2, 'john_tech', '$2a$10$WK2MAXCLDJxpnXME2y6lDO8sEYZ/eadY6pyjW7ZMJaWHIWW4lZvFe', 'technician', '2024-12-31 19:57:11', '2024-12-31 19:57:11' UNION ALL
                  SELECT 4, 'sarah_manager', '$2a$10$VTvCshCQFLLhqcqYa73/ruo2j4suLgR82M05hkaM/hGwOW1Jlvbxa', 'manager', '2025-01-01 15:07:50', '2025-01-01 15:07:50' UNION ALL
                  SELECT 7, 'makcim', '$2a$10$8PcQ2DFWVej7.SmzHmat7./IcIWZOrmeC1Hr7YP8AQF2COrLZ7BK.', 'technician', '2025-01-01 15:08:19', '2025-01-01 15:08:19'
              ) AS tmp
WHERE NOT EXISTS (SELECT 1 FROM users);

-- Insert tasks if table is empty
INSERT INTO tasks (id, summary, performed_date, technician_id, created_at, updated_at)
SELECT * FROM (
                  SELECT '0e1667ec-e1ad-4210-9055-7c35e935a316', 'Replaced the faulty part on the second device.', '2024-12-29 10:30:00', 1, '2024-12-29 20:28:17', '2024-12-29 20:28:17' UNION ALL
                  SELECT '2182d110-43d6-4d21-8bd4-ad243ca6dec7', 'Replaced the faulty part on the device.', '2024-12-29 10:30:00', 1, '2024-12-31 21:13:26', '2024-12-31 21:13:26' UNION ALL
                  SELECT '2565edbe-8e94-43ff-9b65-b168016f8314', 'Replaced the faulty part on the second device.', '2024-12-29 10:30:00', 1, '2024-12-29 19:07:20', '2024-12-29 19:07:20' UNION ALL
                  SELECT '2e7fbe7c-563f-4764-a66d-3c7ecc2ab7be', 'Replaced the faulty part on the second device.', '2024-12-29 10:30:00', 1, '2024-12-29 20:28:18', '2024-12-29 20:28:18' UNION ALL
                  SELECT 'c810fb8d-ae98-42f8-b28e-b29f4ea24e9b', 'Replaced the faulty part on the device.', '2024-12-29 10:30:00', 2, '2025-01-01 23:44:06', '2025-01-01 23:44:06' UNION ALL
                  SELECT 'cbe97ff7-0ae6-4317-a06b-560991833ed3', 'Replaced the faulty part on the device.', '2024-12-29 10:30:00', 2, '2025-01-01 23:44:03', '2025-01-01 23:44:03'
              ) AS tmp
WHERE NOT EXISTS (SELECT 1 FROM tasks);