# ====================================================================================
# SIMDOKPOL - INTERACTIVE BUILD SYSTEM 🎮
# Fitur: Auto-Versioning, Interactive Menu, Cross-Platform, Auto-Dependency
# ====================================================================================

APP_NAME := simdokpol
BUILD_DIR := build
MAIN_FILE := cmd/main.go

# --- 🤖 AUTO DETECT & INCREMENT VERSION ---
# 1. Ambil tag terakhir dari git (jika error/kosong, default v1.0.0)
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")

# 2. Logic Increment (Split by '.', ambil angka terakhir, tambah 1)
# Contoh: v1.0.0 -> v1.0.1 | v1.1.5 -> v1.1.6
VERSION := $(shell echo $(CURRENT_TAG) | awk -F. -v OFS=. '{$NF+=1; print}')

# --- 🔐 SECRET KEY ---
APP_SECRET_KEY ?= 7333333bcdd58fa770265c4f8b661e802de3fe697fb375a77d2095762a904506

# --- LDFLAGS ---
LDFLAGS := -w -s \
	-X 'main.version=$(VERSION)' \
	-X 'simdokpol/internal/services.AppSecretKeyString=$(APP_SECRET_KEY)' \
	-X 'main.appSecretKey=$(APP_SECRET_KEY)'

# Warna Terminal (Biar Keren)
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all package menu windows linux tools deps

# ====================================================================================
# 🎮 INTERACTIVE MENU (DEFAULT TARGET)
# ====================================================================================
package:
	@clear
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "$(CYAN)   👮 SIMDOKPOL BUILD MANAGER v2.0   $(RESET)"
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "🔹 Versi Terakhir di Git : $(YELLOW)$(CURRENT_TAG)$(RESET)"
	@echo "🔹 Versi Build Sekarang  : $(GREEN)$(VERSION)$(RESET) (Auto-Increment)"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@echo "Pilih Target Build:"
	@echo "  [1] 🪟  Windows (x64 .exe)"
	@echo "  [2] 🐧  Linux (x64 Binary)"
	@echo "  [3] 🛠️   Admin Tools (Keygen & License Mgr)"
	@echo "  [4] 📦  Build SEMUANYA (Full Package)"
	@echo "  [5] 🔖  Push Tag Baru ($(VERSION)) ke Git"
	@echo "  [0] ❌  Keluar"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@read -p "👉 Masukkan Nomor: " choice; \
	if [ "$$choice" = "1" ]; then $(MAKE) windows; \
	elif [ "$$choice" = "2" ]; then $(MAKE) linux; \
	elif [ "$$choice" = "3" ]; then $(MAKE) tools; \
	elif [ "$$choice" = "4" ]; then $(MAKE) all; \
	elif [ "$$choice" = "5" ]; then $(MAKE) git-tag; \
	else echo "$(RED)Dibatalkan.$(RESET)"; exit 0; fi

# ====================================================================================
# 🏗️ BUILD TARGETS
# ====================================================================================

# Build All
all: windows linux tools
	@echo "$(GREEN)🎉 SEMUA BUILD SELESAI! Cek folder $(BUILD_DIR)/$(RESET)"

# --- WINDOWS ---
windows: check-windows-deps
	@echo "$(YELLOW)🔨 Membangun Windows x64 (v$(VERSION))...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS) -H=windowsgui" -o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE)
	@cp README.md $(BUILD_DIR)/windows/
	@echo "$(GREEN)✅ Windows Build Selesai: $(BUILD_DIR)/windows/$(APP_NAME).exe$(RESET)"

# --- LINUX ---
linux: check-linux-deps
	@echo "$(YELLOW)🔨 Membangun Linux x64 (v$(VERSION))...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE)
	@chmod +x $(BUILD_DIR)/linux/$(APP_NAME)
	@cp README.md $(BUILD_DIR)/linux/
	@echo "$(GREEN)✅ Linux Build Selesai: $(BUILD_DIR)/linux/$(APP_NAME)$(RESET)"

# --- ADMIN TOOLS ---
tools: check-windows-deps
	@echo "$(YELLOW)🔨 Membangun Admin Tools (v$(VERSION))...$(RESET)"
	@mkdir -p $(BUILD_DIR)/tools
	@# GUI
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS) -H=windowsgui" -o $(BUILD_DIR)/tools/LicenseManager.exe cmd/license-manager/main.go
	@# CLI
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/tools/SignerCLI.exe cmd/signer/main.go
	@echo "$(GREEN)✅ Tools Selesai di $(BUILD_DIR)/tools/$(RESET)"

# --- GIT TAGGING ---
git-tag:
	@echo "$(YELLOW)🔖 Membuat Tag Git Baru: $(VERSION)...$(RESET)"
	@git tag $(VERSION)
	@git push origin $(VERSION)
	@echo "$(GREEN)✅ Tag $(VERSION) berhasil dipush! CI/CD GitHub akan berjalan otomatis.$(RESET)"

# ====================================================================================
# 📦 DEPENDENCY CHECKER
# ====================================================================================
check-linux-deps:
	@if ! dpkg -s libgtk-3-dev >/dev/null 2>&1; then \
		echo "$(RED)📦 Library GTK3 belum ada. Menginstall...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y libgtk-3-dev libayatana-appindicator3-dev; \
	fi

check-windows-deps:
	@if ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then \
		echo "$(RED)📦 Compiler MinGW belum ada. Menginstall...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y gcc-mingw-w64; \
	fi

clean:
	@rm -rf $(BUILD_DIR)
	@echo "$(YELLOW)🧹 Folder build dibersihkan.$(RESET)"