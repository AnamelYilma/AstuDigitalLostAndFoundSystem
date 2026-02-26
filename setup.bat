@echo off
echo ========================================
echo lostfound
echo ========================================
echo.

:: Step 1: Clean everything
echo [1/7] Cleaning old files...
if exist go.mod del go.mod
if exist go.sum del go.sum
if exist vendor rmdir /s /q vendor
echo Done.
echo.

:: Step 2: Create fresh go.mod
echo [2/7] Creating fresh go.mod...
go mod init lostfound
echo Done.
echo.

:: Step 3: Get all dependencies (one command each)
echo [3/7] Installing Gin framework...
go get github.com/gin-gonic/gin@v1.9.1
echo Done.
echo.

echo [4/7] Installing GORM and PostgreSQL driver...
go get gorm.io/gorm@v1.25.5
go get gorm.io/driver/postgres@v1.5.4
echo Done.
echo.

echo [5/7] Installing crypto and sessions...
go get golang.org/x/crypto@v0.48.0
go get github.com/gorilla/sessions@v1.2.2
echo Done.
echo.

echo [6/7] Installing utilities...
go get github.com/joho/godotenv@v1.5.1
echo Done.
echo.

:: Step 4: Tidy up
echo [7/7] Tidying up modules...
go mod tidy
echo Done.
echo.

:: Step 5: Show result
echo ========================================
echo ✅ SETUP COMPLETE!
echo ========================================
echo.
go list -m all
echo.
echo To run the project:
echo go run cmd/main.go
echo.
pause