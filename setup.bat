@echo off
echo ========================================
echo ASTU Lost & Found - COMPLETE CLEAN INSTALL
echo ========================================
echo.

echo [1/8] Deleting ALL old files...
if exist go.mod del /f go.mod
if exist go.sum del /f go.sum
if exist vendor rmdir /s /q vendor
echo Done.
echo.

echo [2/8] Creating fresh go.mod...
go mod init lostfound
echo Done.
echo.

echo [3/8] Installing ALL dependencies...
echo Installing Gin...
go get github.com/gin-gonic/gin@v1.9.1
echo Installing Sessions...
go get github.com/gorilla/sessions@v1.2.2
echo Installing Godotenv...
go get github.com/joho/godotenv@v1.5.1
echo Installing GORM...
go get gorm.io/gorm@v1.25.5
echo Installing PostgreSQL driver...
go get gorm.io/driver/postgres@v1.5.4
echo Installing Crypto...
go get golang.org/x/crypto@v0.48.0
echo Done.
echo.

echo [4/8] Tidying up...
go mod tidy
echo Done.
echo.

echo [5/8] Creating vendor folder (for offline)...
go mod vendor
echo Done.
echo.

echo [6/8] Testing compilation...
go build -o test.exe
if exist test.exe del test.exe
echo Done.
echo.

echo [7/8] Creating .env file with your settings...
echo DB_HOST=localhost > .env
echo DB_USER=postgres >> .env
echo DB_PASSWORD=0909 >> .env
echo DB_NAME=lostfound >> .env
echo DB_PORT=5432 >> .env
echo SESSION_SECRET=my-super-secret-key-2024 >> .env
echo Done.
echo.

echo [8/8] Creating database (if not exists)...
echo Please make sure PostgreSQL is running (pgAdmin4)
echo.
echo To create database manually in pgAdmin4:
echo 1. Open pgAdmin4
echo 2. Right-click on "Databases"
echo 3. Create ^> Database
echo 4. Name: lostfound
echo 5. Save
echo.
pause

