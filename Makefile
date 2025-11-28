# ====================================================================================
# SIMDOKPOL - ULTIMATE BUILD SYSTEM (DYNAMIC VERSION) 🚀
# Fitur: Auto-Version from Git, Auto-Generate JSON Metadata, Cross-Platform
# ====================================================================================

# --- ⚙️ KONFIGURASI ---
APP_NAME := simdokpol
BUILD_DIR := build
RELEASE_DIR := release
MAIN_FILE := cmd/main.go
RESOURCE_SYSO := cmd/resource.syso
ICON_PATH := web/static/img/icon.ico

# --- 🤖 AUTO VERSIONING LOGIC ---
# 1. Ambil tag terakhir (cth: v1.2.0)
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
# 2. Increment Patch (v1.2.0 -> v1.2.1)
VERSION_FULL := $(shell echo $(CURRENT_TAG) | awk -F. -v OFS=. '{$$NF+=1; print}')
# 3. Bersihkan prefix 'v' (1.2.1)
VERSION_RAW := $(patsubst v%,%,$(VERSION_FULL))
# 4. Pecah jadi Major, Minor, Patch
VER_MAJOR := $(word 1,$(subst ., ,$(VERSION_RAW)))
VER_MINOR := $(word 2,$(subst ., ,$(VERSION_RAW)))
VER_PATCH := $(word 3,$(subst ., ,$(VERSION_RAW)))

PREV_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "HEAD")

# --- 🔐 SECRET KEY ---
APP_SECRET_KEY ?= 5f785386230f725107db9bba20c423c0badd0d5002b09eafd6adb092b2a827f5

# --- LDFLAGS ---
LDFLAGS_COMMON := -w -s -X 'main.version=$(VERSION_FULL)'
LDFLAGS_APP    := $(LDFLAGS_COMMON) -X 'simdokpol/internal/services.AppSecretKeyString=$(APP_SECRET_KEY)'
LDFLAGS_TOOL   := $(LDFLAGS_COMMON) -X 'main.appSecretKey=$(APP_SECRET_KEY)'

# Warna
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all package menu windows linux tools changelog clean icon-gen deps check-windows-deps check-linux-deps

# ====================================================================================
# 🎮 MENU
# ====================================================================================
package:
	@clear
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "$(CYAN)   👮 SIMDOKPOL BUILDER v6.0 (Dynamic Edition)   $(RESET)"
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "🏷️  Versi Git Terakhir : $(YELLOW)$(CURRENT_TAG)$(RESET)"
	@echo "🚀 Versi Build Ini    : $(GREEN)$(VERSION_FULL)$(RESET)"
	@echo "🔢 Metadata Windows   : $(VER_MAJOR).$(VER_MINOR).$(VER_PATCH).0"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@echo "Pilih Target:"
	@echo "  [1] 🚀  RELEASE FULL (Win+Lin+Tools+Zip)"
	@echo "  [2] 🪟  Windows Only (.exe + Metadata)"
	@echo "  [3] 🐧  Linux Only"
	@echo "  [4] 🛠️   Admin Tools"
	@echo "  [5] 📝  Generate Changelog"
	@echo "  [0] ❌  Keluar"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@read -p "👉 Pilih: " c; \
	case $$c in \
		1) $(MAKE) release ;; \
		2) $(MAKE) windows ;; \
		3) $(MAKE) linux ;; \
		4) $(MAKE) tools ;; \
		5) $(MAKE) changelog ;; \
		*) echo "$(RED)Bye!$(RESET)" ;; \
	esac

# --- RELEASE PIPELINE ---
release: clean deps changelog windows linux tools
	@echo "$(YELLOW)📦 Membungkus Paket Rilis...$(RESET)"
	@mkdir -p $(RELEASE_DIR)
	
	@cd $(BUILD_DIR)/windows && zip -r ../../$(RELEASE_DIR)/$(APP_NAME)-windows-portable-$(VERSION_FULL).zip $(APP_NAME).exe
	@cd $(BUILD_DIR) && zip -j ../$(RELEASE_DIR)/$(APP_NAME)-windows-portable-$(VERSION_FULL).zip CHANGELOG.txt
	@zip -j $(RELEASE_DIR)/$(APP_NAME)-windows-portable-$(VERSION_FULL).zip README.md LICENSE
	
	@cd $(BUILD_DIR)/linux && tar -czf ../../$(RELEASE_DIR)/$(APP_NAME)-linux-portable-$(VERSION_FULL).tar.gz $(APP_NAME)
	@cd $(BUILD_DIR) && tar -rf ../$(RELEASE_DIR)/$(APP_NAME)-linux-portable-$(VERSION_FULL).tar.gz CHANGELOG.txt
	@tar -rf $(RELEASE_DIR)/$(APP_NAME)-linux-portable-$(VERSION_FULL).tar.gz README.md LICENSE
	@gzip -f $(RELEASE_DIR)/$(APP_NAME)-linux-portable-$(VERSION_FULL).tar

	@cd $(BUILD_DIR)/tools/windows && zip -r ../../../$(RELEASE_DIR)/AdminTools-$(VERSION_FULL).zip *.exe
	
	@echo "$(GREEN)✅ RELEASE SELESAI! Cek folder '$(RELEASE_DIR)/'$(RESET)"

# --- 🖼️ ICON & METADATA GENERATOR (DINAMIS) ---
icon-gen:
	@echo "$(YELLOW)🖼️  Generating Dynamic Windows Metadata...$(RESET)"
	@if ! command -v goversioninfo >/dev/null 2>&1; then \
		echo "   Installing goversioninfo..."; \
		go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest; \
	fi
	
	@echo '{' > versioninfo.json
	@echo '    "FixedFileInfo": {' >> versioninfo.json
	@echo '        "FileVersion": {' >> versioninfo.json
	@echo '            "Major": $(VER_MAJOR),' >> versioninfo.json
	@echo '            "Minor": $(VER_MINOR),' >> versioninfo.json
	@echo '            "Patch": $(VER_PATCH),' >> versioninfo.json
	@echo '            "Build": 0' >> versioninfo.json
	@echo '        },' >> versioninfo.json
	@echo '        "ProductVersion": {' >> versioninfo.json
	@echo '            "Major": $(VER_MAJOR),' >> versioninfo.json
	@echo '            "Minor": $(VER_MINOR),' >> versioninfo.json
	@echo '            "Patch": $(VER_PATCH),' >> versioninfo.json
	@echo '            "Build": 0' >> versioninfo.json
	@echo '        },' >> versioninfo.json
	@echo '        "FileFlagsMask": "3f",' >> versioninfo.json
	@echo '        "FileFlags": "00",' >> versioninfo.json
	@echo '        "FileOS": "040004",' >> versioninfo.json
	@echo '        "FileType": "01",' >> versioninfo.json
	@echo '        "FileSubType": "00"' >> versioninfo.json
	@echo '    },' >> versioninfo.json
	@echo '    "StringFileInfo": {' >> versioninfo.json
	@echo '        "Comments": "Sistem Informasi Manajemen Dokumen Kepolisian",' >> versioninfo.json
	@echo '        "CompanyName": "MYA",' >> versioninfo.json
	@echo '        "FileDescription": "Aplikasi SIMDOKPOL Desktop",' >> versioninfo.json
	@echo '        "FileVersion": "$(VERSION_RAW)",' >> versioninfo.json
	@echo '        "InternalName": "$(APP_NAME)",' >> versioninfo.json
	@echo '        "LegalCopyright": "Copyright (c) 2025 Muhammad Yusuf Abdurrohman",' >> versioninfo.json
	@echo '        "OriginalFilename": "$(APP_NAME).exe",' >> versioninfo.json
	@echo '        "ProductName": "SIMDOKPOL",' >> versioninfo.json
	@echo '        "ProductVersion": "$(VERSION_FULL)"' >> versioninfo.json
	@echo '    },' >> versioninfo.json
	@echo '    "VarFileInfo": {' >> versioninfo.json
	@echo '        "Translation": {' >> versioninfo.json
	@echo '            "LangID": "0409",' >> versioninfo.json
	@echo '            "CharsetID": "04B0"' >> versioninfo.json
	@echo '        }' >> versioninfo.json
	@echo '    },' >> versioninfo.json
	@echo '    "IconPath": "$(ICON_PATH)",' >> versioninfo.json
	@echo '    "ManifestPath": ""' >> versioninfo.json
	@echo '}' >> versioninfo.json
	
	@goversioninfo -o $(RESOURCE_SYSO)
	@rm -f versioninfo.json
	@echo "   Metadata injected: v$(VERSION_RAW)"

# --- 🏗️ BUILD TARGETS ---

windows: check-windows-deps icon-gen
	@echo "$(CYAN)🔨 Building Windows App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE)
	@rm -f $(RESOURCE_SYSO)
	@echo "$(GREEN)✅ Windows OK.$(RESET)"

linux: check-linux-deps
	@echo "$(CYAN)🔨 Building Linux App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE)
	@chmod +x $(BUILD_DIR)/linux/$(APP_NAME)
	@echo "$(GREEN)✅ Linux OK.$(RESET)"

tools: check-windows-deps
	@echo "$(CYAN)🔨 Building Admin Tools...$(RESET)"
	@mkdir -p $(BUILD_DIR)/tools/windows
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_TOOL) -H=windowsgui" \
	-o $(BUILD_DIR)/tools/windows/LicenseManager.exe cmd/license-manager/main.go
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_TOOL)" \
	-o $(BUILD_DIR)/tools/windows/SignerCLI.exe cmd/signer/main.go
	@echo "$(GREEN)✅ Tools OK.$(RESET)"

changelog:
	@echo "$(YELLOW)📝 Generating Changelog...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@echo "CHANGELOG - $(APP_NAME) $(VERSION_FULL)" > $(BUILD_DIR)/CHANGELOG.txt
	@echo "" >> $(BUILD_DIR)/CHANGELOG.txt
	@git log $(PREV_VERSION)..HEAD --pretty=format:"- %s" >> $(BUILD_DIR)/CHANGELOG.txt 2>/dev/null || echo "- Initial release" >> $(BUILD_DIR)/CHANGELOG.txt
	@echo "" >> $(BUILD_DIR)/CHANGELOG.txt

# --- DEPS CHECK ---
deps:
	@echo "$(YELLOW)🔍 Checking Dependencies...$(RESET)"
	@go mod tidy

check-windows-deps:
	@if ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then \
		echo "$(RED)❌ MinGW not found. Installing...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y gcc-mingw-w64; \
	fi

check-linux-deps:
	@if ! dpkg -s libgtk-3-dev >/dev/null 2>&1; then \
		echo "$(RED)❌ GTK3 not found. Installing...$(RESET)"; \
		sudo apt-get install -y libgtk-3-dev libayatana-appindicator3-dev; \
	fi

clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR) $(RESOURCE_SYSO) versioninfo.json
	@echo "$(GREEN)✅ Clean complete!$(RESET)"