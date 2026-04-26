@echo off
REM Run the POS system with MySQL connection to XAMPP phpMyAdmin
SET DB_DRIVER=mysql
SET DB_DSN=root:@tcp(127.0.0.1:3306)/posdb?parseTime=true

echo Starting POS system using MySQL database posdb...

go run main.go
