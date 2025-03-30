CREATE TABLE `trx_users_services` (
    `service_id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NOT NULL,  
    `service_name` VARCHAR(255) NOT NULL,
    `service_user_id` VARCHAR(255) NOT NULL,
    `encrypted_access_token` VARCHAR(512) NOT NULL,
    `encrypted_refresh_token` VARCHAR(512) NOT NULL, 
    `expires_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (`user_id`) REFERENCES `trx_users`(`user_id`) ON DELETE CASCADE
);
