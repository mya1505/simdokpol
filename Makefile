# ====================================================================================
# SIMDOKPOL - BUILD SYSTEM (WINDOWS & LINUX) 🛠️
# Target: Windows (x64) & Linux (x64)
# Support: Cross-Compile dari Linux ke Windows menggunakan MinGW
# ====================================================================================

APP_NAME := simdokpol
VERSION := v1.1.0
BUILD_DIR := build
MAIN_FILE := cmd/main.go

# --- 🔐 SECRET KEY (GANTI DENGAN HASIL OPENSSL) ---
# Ini kunci rahasia untuk HMAC. Harus sama di App Utama & Admin Tools.
APP_SECRET_KEY ?= RAHASIA_DAPUR_POLSEK_BAHODOPI_JANGAN_DISEBAR_12345

# --- LDFLAGS MAGIC ---
# Inject Version & Secret Key ke dalam variable Go saat compile
LDFLAGS := -w -s \
	-X 'main.version=$(VERSION)' \
	-X 'simdokpol/internal/services.AppSecretKeyString=$(APP_SECRET_KEY)' \
	-X 'main.appSecretKey=$(APP_SECRET_KEY)'

.PHONY: all clean windows linux tools tools-windows tools-linux

# Build Semuanya (App Utama + Tools Admin untuk kedua OS)
all: clean windows linux tools

# --- 🪟 APP UTAMA (WINDOWS) ---
windows:
	@echo "🪟 [Windows] Membangun SIMDOKPOL..."
	@mkdir -p $(BUILD_DIR)/windows
	@# Menggunakan MinGW Compiler untuk CGO (SQLite)
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS) -H=windowsgui" -o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE)
	@cp README.md $(BUILD_DIR)/windows/
	@echo "✅ Windows App Selesai: $(BUILD_DIR)/windows/$(APP_NAME).exe"

# --- 🐧 APP UTAMA (LINUX) ---
linux:
	@echo "🐧 [Linux] Membangun SIMDOKPOL..."
	@mkdir -p $(BUILD_DIR)/linux
	@# Build native Linux
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE)
	@chmod +x $(BUILD_DIR)/linux/$(APP_NAME)
	@cp README.md $(BUILD_DIR)/linux/
	@echo "✅ Linux App Selesai: $(BUILD_DIR)/linux/$(APP_NAME)"

# --- 🛠️ ADMIN TOOLS (ALL OS) ---
tools: tools-windows tools-linux

# Admin Tools untuk Windows (.exe)
tools-windows:
	@echo "🪟 [Windows] Membangun Admin Tools..."
	@mkdir -p $(BUILD_DIR)/tools/windows
	
	@# 1. License Manager GUI (Windows)
	@echo "   > License Manager GUI..."
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS) -H=windowsgui" -o $(BUILD_DIR)/tools/windows/LicenseManager.exe cmd/license-manager/main.go
	
	@# 2. Signer CLI (Windows)
	@echo "   > Signer CLI..."
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/tools/windows/SignerCLI.exe cmd/signer/main.go
	@echo "✅ Windows Tools Selesai."

# Admin Tools untuk Linux (Binary)
tools-linux:
	@echo "🐧 [Linux] Membangun Admin Tools..."
	@mkdir -p $(BUILD_DIR)/tools/linux
	
	@# 1. License Manager GUI (Linux) - Butuh libgtk-3-dev
	@echo "   > License Manager GUI..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/tools/linux/LicenseManager cmd/license-manager/main.go
	
	@# 2. Signer CLI (Linux)
	@echo "   > Signer CLI..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/tools/linux/SignerCLI cmd/signer/main.go
	
	@chmod +x $(BUILD_DIR)/tools/linux/*
	@echo "✅ Linux Tools Selesai."

# --- UTILS ---
clean:
	@echo "🧹 Membersihkan folder build..."
	@rm -rf $(BUILD_DIR)

package: all
	@echo "📦 Membungkus Release..."
	@zip -j $(BUILD_DIR)/simdokpol-windows.zip $(BUILD_DIR)/windows/$(APP_NAME).exe README.md
	@zip -j $(BUILD_DIR)/simdokpol-admin-tools-windows.zip $(BUILD_DIR)/tools/windows/*.exe
	@tar -czf $(BUILD_DIR)/simdokpol-linux.tar.gz -C $(BUILD_DIR)/linux $(APP_NAME)
	@tar -czf $(BUILD_DIR)/simdokpol-admin-tools-linux.tar.gz -C $(BUILD_DIR)/tools/linux LicenseManager SignerCLI
	@echo "🎉 Paket Siap Didistribusikan!"