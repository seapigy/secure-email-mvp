-- Create securemail user if it doesn't exist
CREATE USER IF NOT EXISTS 'securemail'@'%' IDENTIFIED BY 'securepassword';

-- Grant all privileges on securemail database
GRANT ALL PRIVILEGES ON securemail.* TO 'securemail'@'%';

-- Also grant privileges for localhost connections
CREATE USER IF NOT EXISTS 'securemail'@'localhost' IDENTIFIED BY 'securepassword';
GRANT ALL PRIVILEGES ON securemail.* TO 'securemail'@'localhost';

-- Flush privileges to ensure changes take effect
FLUSH PRIVILEGES;



