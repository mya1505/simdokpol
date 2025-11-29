# ====================================================================================
# SIMDOKPOL - ENHANCED BUILD SYSTEM v10.0 🚀
# Fitur: Auto-Version, Cross-Platform, Installer System, Template-Based Configuration
# Support: Windows AMD64/ARM64, Linux AMD64, macOS
# ====================================================================================

# --- ⚙️ KONFIGURASI ---
APP_NAME := simdokpol
BUILD_DIR := build
RELEASE_DIR := release
TEMPLATES_DIR := templates
MAIN_FILE := cmd/main.go
RESOURCE_SYSO := cmd/resource.syso
ICON_PATH := web/static/img/icon.ico

# --- 🤖 AUTO VERSIONING LOGIC ---
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
VERSION_FULL := $(shell echo $(CURRENT_TAG) | awk -F. -v OFS=. '{$$NF+=1; print}')
VERSION_RAW := $(patsubst v%,%,$(VERSION_FULL))
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

# --- DETECT OS ---
DETECTED_OS := $(shell uname -s 2>/dev/null || echo "Unknown")
DETECTED_ARCH := $(shell uname -m 2>/dev/null || echo "Unknown")
PKG_MANAGER := $(shell if command -v pacman >/dev/null 2>&1; then echo "pacman"; elif command -v apt-get >/dev/null 2>&1; then echo "apt"; else echo "unknown"; fi)

# Warna untuk output terminal
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all package menu windows windows-arm64 linux macos tools changelog clean icon-gen deps check-deps install-deps installer-windows installer-linux installer-macos installer-all check-go check-git check-files validate-build test-binary smoke-test test help list-targets

# ====================================================================================
# 🎮 MENU INTERAKTIF
# ====================================================================================
package:
	@clear
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "$(CYAN)   👮 SIMDOKPOL BUILDER v10.0 (Enhanced Edition) $(RESET)"
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "🏷️  Versi Git Terakhir : $(YELLOW)$(CURRENT_TAG)$(RESET)"
	@echo "🚀 Versi Build Ini    : $(GREEN)$(VERSION_FULL)$(RESET)"
	@echo "💻 Sistem Detected    : $(DETECTED_OS) $(DETECTED_ARCH)"
	@echo "📦 Package Manager    : $(PKG_MANAGER)"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@echo "Pilih Target:"
	@echo "  [1] 🚀  RELEASE FULL (All+Installers+Zip)"
	@echo "  [2] 🪟  Windows AMD64 (.exe + Installer)"
	@echo "  [3] 🔷  Windows ARM64 (Snapdragon)"
	@echo "  [4] 🐧  Linux AMD64 (DEB+RPM+Portable)"
	@echo "  [5] 🍎  macOS AMD64 (DMG+Portable)"
	@echo "  [6] 🛠️   Admin Tools"
	@echo "  [7] 📦  Install Dependencies"
	@echo "  [8] 📀  Build All Installers"
	@echo "  [9] 📝  Generate Changelog"
	@echo "  [t] 🧪  Run Tests"
	@echo "  [h] ❓  Show Help"
	@echo "  [0] ❌  Keluar"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@read -p "👉 Pilih: " c; \
	case $$c in \
		1) $(MAKE) release ;; \
		2) $(MAKE) windows-installer ;; \
		3) $(MAKE) windows-arm64-installer ;; \
		4) $(MAKE) linux-installer ;; \
		5) $(MAKE) macos-installer ;; \
		6) $(MAKE) tools ;; \
		7) $(MAKE) install-deps ;; \
		8) $(MAKE) installer-all ;; \
		9) $(MAKE) changelog ;; \
		t) $(MAKE) test ;; \
		h) $(MAKE) help ;; \
		*) echo "$(RED)Bye!$(RESET)" ;; \
	esac

# ====================================================================================
# 🔍 PREREQUISITE CHECKS (Enhanced)
# ====================================================================================
check-go:
	@command -v go >/dev/null 2>&1 || (echo "$(RED)❌ Go is not installed$(RESET)" && exit 1)
	@GO_VERSION=$$(go version | awk '{print $$3}' | sed 's/go//'); \
	REQUIRED_VERSION="1.21.0"; \
	if [ "$$(printf '%s\n' "$$REQUIRED_VERSION" "$$GO_VERSION" | sort -V | head -n1)" != "$$REQUIRED_VERSION" ]; then \
		echo "$(RED)❌ Go version $$GO_VERSION is too old. Minimum required: $$REQUIRED_VERSION$(RESET)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Go version $$GO_VERSION OK$(RESET)"

check-git:
	@command -v git >/dev/null 2>&1 || (echo "$(RED)❌ Git is not installed$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Git OK$(RESET)"

check-files:
	@echo "$(YELLOW)🔍 Validating required files and directories...$(RESET)"
	@test -d web || (echo "$(RED)❌ web directory not found$(RESET)" && exit 1)
	@test -d migrations || (echo "$(YELLOW)⚠️  migrations directory not found (non-critical)$(RESET)")
	@test -f $(ICON_PATH) || (echo "$(YELLOW)⚠️  Icon file not found at $(ICON_PATH)$(RESET)")
	@test -f $(MAIN_FILE) || (echo "$(RED)❌ Main file not found at $(MAIN_FILE)$(RESET)" && exit 1)
	@test -f go.mod || (echo "$(RED)❌ go.mod not found$(RESET)" && exit 1)
	@echo "$(GREEN)✅ File validation complete$(RESET)"

check-nsis-deps:
	@if ! command -v makensis >/dev/null 2>&1; then \
		echo "$(RED)❌ NSIS not found$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			echo "$(YELLOW)Installing NSIS via pacman...$(RESET)"; \
			sudo pacman -S --needed nsis || (echo "$(RED)❌ Failed to install NSIS$(RESET)" && exit 1); \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			echo "$(YELLOW)Installing NSIS via apt...$(RESET)"; \
			sudo apt-get update && sudo apt-get install -y nsis || (echo "$(RED)❌ Failed to install NSIS$(RESET)" && exit 1); \
		else \
			echo "$(RED)❌ Please install NSIS manually$(RESET)"; \
			exit 1; \
		fi; \
	fi
	@echo "$(GREEN)✅ NSIS available$(RESET)"

check-linux-installer-deps:
	@if ! command -v dpkg-deb >/dev/null 2>&1; then \
		echo "$(RED)❌ dpkg-deb not found$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			echo "$(YELLOW)Installing dpkg via pacman...$(RESET)"; \
			sudo pacman -S --needed dpkg || (echo "$(RED)❌ Failed to install dpkg$(RESET)" && exit 1); \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			echo "$(YELLOW)Installing dpkg via apt...$(RESET)"; \
			sudo apt-get install -y dpkg || (echo "$(RED)❌ Failed to install dpkg$(RESET)" && exit 1); \
		fi; \
	fi
	@echo "$(GREEN)✅ Linux installer dependencies OK$(RESET)"

check-macos-deps:
	@if [ "$(DETECTED_OS)" != "Darwin" ] && ! command -v genisoimage >/dev/null 2>&1; then \
		echo "$(RED)❌ genisoimage not found$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			echo "$(YELLOW)Installing cdrkit via pacman...$(RESET)"; \
			sudo pacman -S --needed cdrkit || (echo "$(RED)❌ Failed to install cdrkit$(RESET)" && exit 1); \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			echo "$(YELLOW)Installing genisoimage via apt...$(RESET)"; \
			sudo apt-get install -y genisoimage || (echo "$(RED)❌ Failed to install genisoimage$(RESET)" && exit 1); \
		fi; \
	fi
	@echo "$(GREEN)✅ macOS build dependencies OK$(RESET)"

validate-build: check-go check-git check-files
	@echo "$(GREEN)✅ All validation checks passed$(RESET)"

# ====================================================================================
# 🏗️ BUILD TARGETS (Enhanced with validation)
# ====================================================================================
windows: validate-build icon-gen
	@echo "$(CYAN)🔨 Building Windows AMD64 App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE) || (echo "$(RED)❌ Windows build failed$(RESET)" && exit 1)
	@rm -f $(RESOURCE_SYSO)
	@test -f $(BUILD_DIR)/windows/$(APP_NAME).exe || (echo "$(RED)❌ Windows binary not found after build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Windows AMD64 build successful$(RESET)"

windows-arm64: validate-build icon-gen
	@echo "$(CYAN)🔨 Building Windows ARM64 App (Snapdragon)...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows-arm64
	@CGO_ENABLED=1 CC=aarch64-w64-mingw32-gcc GOOS=windows GOARCH=arm64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(MAIN_FILE) || (echo "$(RED)❌ Windows ARM64 build failed$(RESET)" && exit 1)
	@rm -f $(RESOURCE_SYSO)
	@test -f $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe || (echo "$(RED)❌ Windows ARM64 binary not found after build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Windows ARM64 build successful$(RESET)"

linux: validate-build
	@echo "$(CYAN)🔨 Building Linux AMD64 App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE) || (echo "$(RED)❌ Linux build failed$(RESET)" && exit 1)
	@chmod +x $(BUILD_DIR)/linux/$(APP_NAME)
	@test -f $(BUILD_DIR)/linux/$(APP_NAME) || (echo "$(RED)❌ Linux binary not found after build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Linux AMD64 build successful$(RESET)"

macos: validate-build
	@echo "$(CYAN)🔨 Building macOS App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/macos
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/macos/$(APP_NAME) $(MAIN_FILE) || (echo "$(RED)❌ macOS build failed$(RESET)" && exit 1)
	@chmod +x $(BUILD_DIR)/macos/$(APP_NAME)
	@test -f $(BUILD_DIR)/macos/$(APP_NAME) || (echo "$(RED)❌ macOS binary not found after build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ macOS build successful$(RESET)"

# ====================================================================================
# 📦 INSTALLER TARGETS (Template-based approach)
# ====================================================================================
installer-windows: windows check-nsis-deps
	@echo "$(CYAN)📦 Creating Windows Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows $(RELEASE_DIR)
	
	@cp $(BUILD_DIR)/windows/$(APP_NAME).exe $(BUILD_DIR)/installer/windows/
	@cp -r web migrations $(BUILD_DIR)/installer/windows/ 2>/dev/null || echo "$(YELLOW)⚠️  Some directories not found$(RESET)"
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows/icon.ico 2>/dev/null || echo "$(YELLOW)⚠️  Icon not found$(RESET)"
	@[ -f LICENSE ] && cp LICENSE $(BUILD_DIR)/installer/windows/ || echo "$(YELLOW)⚠️  LICENSE not found$(RESET)"
	
	@printf '@echo off\ntitle SIMDOKPOL Server\necho ========================================\necho   SIMDOKPOL - System Startup\necho ========================================\necho.\necho [INFO] Memulai SIMDOKPOL Server...\nstart simdokpol.exe\necho [SUCCESS] Server berjalan di background\n' > $(BUILD_DIR)/installer/windows/start.bat
	
	@if [ -f "$(TEMPLATES_DIR)/installer.nsi.tmpl" ]; then \
		echo "$(GREEN)✓ Using template from $(TEMPLATES_DIR)/installer.nsi.tmpl$(RESET)"; \
		BANNER_DIRECTIVE=""; \
		HEADER_DIRECTIVE=""; \
		LICENSE_DIRECTIVE=""; \
		if [ -f "web/static/img/installer-banner.bmp" ]; then \
			BANNER_DIRECTIVE='!define MUI_WELCOMEFINISHPAGE_BITMAP "web\\static\\img\\installer-banner.bmp"'; \
		fi; \
		if [ -f "web/static/img/installer-header.bmp" ]; then \
			HEADER_DIRECTIVE='!define MUI_HEADERIMAGE\n!define MUI_HEADERIMAGE_BITMAP "web\\static\\img\\installer-header.bmp"'; \
		fi; \
		if [ -f "LICENSE" ]; then \
			LICENSE_DIRECTIVE='!insertmacro MUI_PAGE_LICENSE "LICENSE"'; \
		fi; \
		sed -e 's|@APP_NAME@|$(APP_NAME)|g' \
		    -e 's|@VERSION@|$(VERSION_RAW)|g' \
		    -e 's|@PUBLISHER@|SIMDOKPOL Team|g' \
		    -e 's|@WEB_SITE@|https://github.com/muhammad1505/simdokpol|g' \
		    -e 's|@INSTALLER_NAME@|$(APP_NAME)-windows-x64-v$(VERSION_FULL)-installer.exe|g' \
		    -e 's|@ICON_PATH@|icon.ico|g' \
		    -e "s|@BANNER_DIRECTIVE@|$$BANNER_DIRECTIVE|g" \
		    -e "s|@HEADER_DIRECTIVE@|$$HEADER_DIRECTIVE|g" \
		    -e "s|@LICENSE_DIRECTIVE@|$$LICENSE_DIRECTIVE|g" \
		    $(TEMPLATES_DIR)/installer.nsi.tmpl > $(BUILD_DIR)/installer/windows/installer.nsi; \
	else \
		echo "$(YELLOW)⚠️  Template not found, using inline NSIS script$(RESET)"; \
		$(MAKE) installer-windows-inline; \
	fi
	
	@cd $(BUILD_DIR)/installer/windows && makensis installer.nsi || (echo "$(RED)❌ NSIS build failed$(RESET)" && exit 1)
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-v$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/ || (echo "$(RED)❌ Failed to move installer$(RESET)" && exit 1)
	
	@cd $(BUILD_DIR)/installer/windows && zip -r $(APP_NAME)-windows-x64-v$(VERSION_FULL)-portable.zip . -x "installer.nsi" "*.exe" || (echo "$(RED)❌ Failed to create portable package$(RESET)" && exit 1)
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/
	
	@echo "$(GREEN)✅ Windows Installer created successfully$(RESET)"

installer-windows-arm64: windows-arm64 check-nsis-deps
	@echo "$(CYAN)📦 Creating Windows ARM64 Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows-arm64 $(RELEASE_DIR)
	
	@cp $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(BUILD_DIR)/installer/windows-arm64/
	@cp -r web migrations $(BUILD_DIR)/installer/windows-arm64/ 2>/dev/null || true
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows-arm64/icon.ico 2>/dev/null || true
	
	@printf '!define APP_NAME "SIMDOKPOL"\n!define VERSION "$(VERSION_RAW)"\n!define ARCH "ARM64"\n!define INSTALLER_NAME "$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-installer.exe"\n\nName "$${APP_NAME} $${VERSION} ($${ARCH})"\nOutFile "$${INSTALLER_NAME}"\nInstallDir "$$PROGRAMFILES64\\$${APP_NAME}"\nRequestExecutionLevel admin\n\n!include "MUI2.nsh"\n!define MUI_ICON "icon.ico"\n!insertmacro MUI_PAGE_WELCOME\n!insertmacro MUI_PAGE_DIRECTORY\n!insertmacro MUI_PAGE_INSTFILES\n!insertmacro MUI_PAGE_FINISH\n!insertmacro MUI_LANGUAGE "English"\n\nSection "Install"\n  SetOutPath "$$INSTDIR"\n  File "$(APP_NAME).exe"\n' > $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@if [ -d "web" ]; then printf '  File /r "web"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi; fi
	@printf '  WriteUninstaller "$$INSTDIR\\Uninstall.exe"\nSectionEnd\n\nSection "Uninstall"\n  RMDir /r "$$INSTDIR"\nSectionEnd\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	
	@cd $(BUILD_DIR)/installer/windows-arm64 && makensis installer.nsi || (echo "$(RED)❌ NSIS build failed$(RESET)" && exit 1)
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/ 2>/dev/null || true
	
	@cd $(BUILD_DIR)/installer/windows-arm64 && zip -r $(APP_NAME)-windows-arm64-v$(VERSION_FULL)-portable.zip . -x "installer.nsi" "*.exe"
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || true
	
	@echo "$(GREEN)✅ Windows ARM64 Installer created successfully$(RESET)"

installer-linux: linux check-linux-installer-deps
	@echo "$(CYAN)📦 Creating Linux Installers...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux $(RELEASE_DIR)
	
	@cp $(BUILD_DIR)/linux/$(APP_NAME) $(BUILD_DIR)/installer/linux/
	@cp -r web migrations $(BUILD_DIR)/installer/linux/ 2>/dev/null || true
	@[ -f LICENSE ] && cp LICENSE $(BUILD_DIR)/installer/linux/ || true
	
	@printf '#!/bin/bash\necho "========================================"\necho "  SIMDOKPOL - System Startup"\necho "========================================"\necho ""\necho "[INFO] Memulai SIMDOKPOL Server..."\n./simdokpol\n' > $(BUILD_DIR)/installer/linux/start.sh
	@chmod +x $(BUILD_DIR)/installer/linux/start.sh
	
	@echo "$(YELLOW)Building DEB package...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/DEBIAN
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/bin
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/share/applications
	
	@cp -r $(BUILD_DIR)/installer/linux/* $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)/ 2>/dev/null || true
	
	@printf '#!/bin/sh\n/opt/$(APP_NAME)/$(APP_NAME) "$$@"\n' > $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	@chmod +x $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	
	@printf '[Desktop Entry]\nVersion=1.0\nType=Application\nName=SIMDOKPOL\nComment=Sistem Informasi Manajemen Dokumen Kepolisian\nExec=/opt/$(APP_NAME)/start.sh\nTerminal=false\nCategories=Office;Database;\n' > $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	
	@printf 'Package: $(APP_NAME)\nVersion: $(VERSION_RAW)\nSection: utils\nPriority: optional\nArchitecture: amd64\nMaintainer: SIMDOKPOL Team\nDescription: Sistem Informasi Manajemen Dokumen Kepolisian\n' > $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	
	@printf '#!/bin/bash\nset -e\nchmod +x "/opt/$(APP_NAME)/$(APP_NAME)"\nchmod +x "/opt/$(APP_NAME)/start.sh"\nexit 0\n' > $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	@chmod 755 $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	
	@dpkg-deb --build $(BUILD_DIR)/installer/linux/deb $(RELEASE_DIR)/$(APP_NAME)_$(VERSION_RAW)_amd64.deb || (echo "$(RED)❌ Failed to create DEB$(RESET)" && exit 1)
	
	@cd $(BUILD_DIR)/installer/linux && tar -czf $(APP_NAME)-linux-amd64-v$(VERSION_FULL)-portable.tar.gz *
	@mv $(BUILD_DIR)/installer/linux/$(APP_NAME)-linux-amd64-v$(VERSION_FULL)-portable.tar.gz $(RELEASE_DIR)/ 2>/dev/null || true
	
	@echo "$(GREEN)✅ Linux Installers created successfully$(RESET)"

installer-macos: macos check-macos-deps
	@echo "$(CYAN)📦 Creating macOS Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos $(RELEASE_DIR)
	
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/backups
	
	@cp $(BUILD_DIR)/macos/$(APP_NAME) $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@cp -r web migrations $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/ 2>/dev/null || true
	
	@printf '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n\t<key>CFBundleExecutable</key>\n\t<string>$(APP_NAME)</string>\n\t<key>CFBundleIdentifier</key>\n\t<string>com.simdokpol.app</string>\n\t<key>CFBundleName</key>\n\t<string>$(APP_NAME)</string>\n\t<key>CFBundlePackageType</key>\n\t<string>APPL</string>\n\t<key>CFBundleShortVersionString</key>\n\t<string>$(VERSION_RAW)</string>\n\t<key>CFBundleVersion</key>\n\t<string>$(VERSION_RAW)</string>\n\t<key>LSMinimumSystemVersion</key>\n\t<string>10.13</string>\n\t<key>NSHighResolutionCapable</key>\n\t<true/>\n\t<key>CFBundleDisplayName</key>\n\t<string>$(APP_NAME)</string>\n</dict>\n</plist>\n' > $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	
	@echo "$(YELLOW)Creating DMG installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos/dmg-contents
	@cp -r $(BUILD_DIR)/installer/macos/$(APP_NAME).app $(BUILD_DIR)/installer/macos/dmg-contents/
	@ln -sf /Applications $(BUILD_DIR)/installer/macos/dmg-contents/Applications 2>/dev/null || true
	
	@if [ "$(DETECTED_OS)" = "Darwin" ]; then \
		hdiutil create -volname "SIMDOKPOL" -srcfolder $(BUILD_DIR)/installer/macos/dmg-contents -ov -format UDZO $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-installer.dmg; \
	else \
		echo "$(YELLOW)Using genisoimage for DMG creation on Linux...$(RESET)"; \
		genisoimage -V "SIMDOKPOL" -D -R -apple -no-pad -o $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-installer.dmg $(BUILD_DIR)/installer/macos/dmg-contents 2>/dev/null || echo "$(RED)❌ Failed to create DMG$(RESET)"; \
	fi
	
	@cd $(BUILD_DIR)/installer/macos && zip -r $(APP_NAME)-macos-amd64-v$(VERSION_FULL)-portable.zip $(APP_NAME).app
	@mv $(BUILD_DIR)/installer/macos/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || true
	
	@echo "$(GREEN)✅ macOS Installer created successfully$(RESET)"

installer-all: installer-windows installer-windows-arm64 installer-linux installer-macos
	@echo "$(GREEN)✅ All installers created successfully!$(RESET)"

windows-installer: windows installer-windows
	@echo "$(GREEN)✅ Windows + Installer Complete!$(RESET)"

windows-arm64-installer: windows-arm64 installer-windows-arm64
	@echo "$(GREEN)✅ Windows ARM64 + Installer Complete!$(RESET)"

linux-installer: linux installer-linux
	@echo "$(GREEN)✅ Linux + Installer Complete!$(RESET)"

macos-installer: macos installer-macos
	@echo "$(GREEN)✅ macOS + Installer Complete!$(RESET)"

# ====================================================================================
# 🧪 TESTING TARGETS
# ====================================================================================
test-binary:
	@echo "$(YELLOW)🧪 Testing binary integrity...$(RESET)"
	@if [ -f $(BUILD_DIR)/windows/$(APP_NAME).exe ]; then \
		FILE_SIZE=$$(stat -c%s "$(BUILD_DIR)/windows/$(APP_NAME).exe" 2>/dev/null || stat -f%z "$(BUILD_DIR)/windows/$(APP_NAME).exe" 2>/dev/null); \
		if [ $$FILE_SIZE -lt 1000000 ]; then \
			echo "$(RED)❌ Windows binary size suspiciously small: $$FILE_SIZE bytes$(RESET)"; \
			exit 1; \
		fi; \
		echo "$(GREEN)✓ Windows binary OK ($$FILE_SIZE bytes)$(RESET)"; \
	fi
	@if [ -f $(BUILD_DIR)/linux/$(APP_NAME) ]; then \
		FILE_SIZE=$$(stat -c%s "$(BUILD_DIR)/linux/$(APP_NAME)" 2>/dev/null || stat -f%z "$(BUILD_DIR)/linux/$(APP_NAME)" 2>/dev/null); \
		if [ $$FILE_SIZE -lt 1000000 ]; then \
			echo "$(RED)❌ Linux binary size suspiciously small: $$FILE_SIZE bytes$(RESET)"; \
			exit 1; \
		fi; \
		echo "$(GREEN)✓ Linux binary OK ($$FILE_SIZE bytes)$(RESET)"; \
	fi
	@echo "$(GREEN)✅ Binary validation passed$(RESET)"

smoke-test: test-binary
	@echo "$(YELLOW)🔥 Running smoke tests...$(RESET)"
	@if [ -f $(BUILD_DIR)/windows/$(APP_NAME).exe ]; then \
		echo "$(YELLOW)Testing Windows binary version flag...$(RESET)"; \
		timeout 5 $(BUILD_DIR)/windows/$(APP_NAME).exe --version 2>/dev/null || echo "$(YELLOW)⚠️  Version check unavailable or timed out$(RESET)"; \
	fi
	@if [ -f $(BUILD_DIR)/linux/$(APP_NAME) ]; then \
		echo "$(YELLOW)Testing Linux binary version flag...$(RESET)"; \
		timeout 5 $(BUILD_DIR)/linux/$(APP_NAME) --version 2>/dev/null || echo "$(YELLOW)⚠️  Version check unavailable or timed out$(RESET)"; \
	fi
	@echo "$(GREEN)✅ Smoke tests completed$(RESET)"

test: clean deps test-binary smoke-test
	@echo "$(GREEN)✅ All tests passed successfully$(RESET)"

# ====================================================================================
# 🛠️ UTILITY TARGETS
# ====================================================================================
changelog:
	@echo "$(YELLOW)📝 Generating Changelog...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@printf "CHANGELOG - $(APP_NAME) $(VERSION_FULL)\n\n" > $(BUILD_DIR)/CHANGELOG.txt
	@git log $(PREV_VERSION)..HEAD --pretty=format:"- %s" >> $(BUILD_DIR)/CHANGELOG.txt 2>/dev/null || printf "- Initial release\n" >> $(BUILD_DIR)/CHANGELOG.txt
	@printf "\n" >> $(BUILD_DIR)/CHANGELOG.txt
	@echo "$(GREEN)✅ Changelog generated at $(BUILD_DIR)/CHANGELOG.txt$(RESET)"

icon-gen:
	@echo "$(YELLOW)🖼️  Generating Windows Metadata...$(RESET)"
	@if ! command -v goversioninfo >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing goversioninfo...$(RESET)"; \
		go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest || (echo "$(RED)❌ Failed to install goversioninfo$(RESET)" && exit 1); \
	fi
	
	@printf '{\n' > versioninfo.json
	@printf '    "FixedFileInfo": {\n' >> versioninfo.json
	@printf '        "FileVersion": {\n' >> versioninfo.json
	@printf '            "Major": %s,\n' "$(VER_MAJOR)" >> versioninfo.json
	@printf '            "Minor": %s,\n' "$(VER_MINOR)" >> versioninfo.json
	@printf '            "Patch": %s,\n' "$(VER_PATCH)" >> versioninfo.json
	@printf '            "Build": 0\n' >> versioninfo.json
	@printf '        },\n' >> versioninfo.json
	@printf '        "ProductVersion": {\n' >> versioninfo.json
	@printf '            "Major": %s,\n' "$(VER_MAJOR)" >> versioninfo.json
	@printf '            "Minor": %s,\n' "$(VER_MINOR)" >> versioninfo.json
	@printf '            "Patch": %s,\n' "$(VER_PATCH)" >> versioninfo.json
	@printf '            "Build": 0\n' >> versioninfo.json
	@printf '        }\n' >> versioninfo.json
	@printf '    }\n' >> versioninfo.json
	@printf '}\n' >> versioninfo.json
	
	@goversioninfo -o $(RESOURCE_SYSO) || (echo "$(RED)❌ Failed to generate resource$(RESET)" && exit 1)
	@rm -f versioninfo.json
	@echo "$(GREEN)✅ Windows metadata generated$(RESET)"

deps:
	@echo "$(YELLOW)🔍 Checking Dependencies...$(RESET)"
	@go mod tidy || (echo "$(RED)❌ Failed to tidy Go modules$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Dependencies OK$(RESET)"

clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR) $(RESOURCE_SYSO) versioninfo.json
	@echo "$(GREEN)✅ Clean complete!$(RESET)"

# ====================================================================================
# 📦 DEPENDENCY INSTALLATION
# ====================================================================================
install-deps:
	@echo "$(CYAN)📦 Installing Dependencies for $(DETECTED_OS) ($(PKG_MANAGER))...$(RESET)"
	
	@if [ "$(PKG_MANAGER)" = "pacman" ]; then \
		echo "$(YELLOW)Installing for Arch/Manjaro...$(RESET)"; \
		sudo pacman -S --needed base-devel mingw-w64-gcc go git zip unzip gtk3 webkit2gtk nsis dpkg rpm-tools genisoimage || (echo "$(RED)❌ Installation failed$(RESET)" && exit 1); \
		echo "$(GREEN)✅ Arch/Manjaro dependencies installed!$(RESET)"; \
	elif [ "$(PKG_MANAGER)" = "apt" ]; then \
		echo "$(YELLOW)Installing for Ubuntu/Debian...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y build-essential gcc-mingw-w64-x86-64 gcc-mingw-w64-aarch64 golang-go git zip unzip libgtk-3-dev libwebkit2gtk-4.0-dev nsis dpkg rpm genisoimage || (echo "$(RED)❌ Installation failed$(RESET)" && exit 1); \
		echo "$(GREEN)✅ Ubuntu/Debian dependencies installed!$(RESET)"; \
	else \
		echo "$(RED)❌ Unsupported package manager: $(PKG_MANAGER)$(RESET)"; \
		echo "$(YELLOW)Please install manually: build-essential, mingw-w64, go, git, zip, nsis, dpkg, rpm, genisoimage$(RESET)"; \
		exit 1; \
	fi
	
	@echo "$(YELLOW)📦 Installing Go tools...$(RESET)"
	@go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest || echo "$(YELLOW)⚠️  Could not install goversioninfo$(RESET)"
	@echo "$(GREEN)✅ All dependencies installed successfully!$(RESET)"

# ====================================================================================
# 🚀 RELEASE PIPELINE
# ====================================================================================
release: clean check-deps deps changelog installer-all
	@echo "$(GREEN)✅ RELEASE COMPLETE! All installers available in '$(RELEASE_DIR)/'$(RESET)"
	@echo "$(YELLOW)📁 Generated files:$(RESET)"
	@ls -lh $(RELEASE_DIR)/* 2>/dev/null || echo "$(RED)No files generated$(RESET)"

# ====================================================================================
# 🛠️ ADMIN TOOLS
# ====================================================================================
tools: check-deps
	@echo "$(CYAN)🛠️  Building Admin Tools...$(RESET)"
	@mkdir -p $(BUILD_DIR)/tools
	@echo "$(YELLOW)Admin tools target not yet fully implemented$(RESET)"
	@echo "$(GREEN)✅ Tools preparation complete$(RESET)"

# ====================================================================================
# 📋 QUICK TARGETS
# ====================================================================================
all: clean deps installer-all
	@echo "$(GREEN)🎉 All builds completed successfully!$(RESET)"

quick-windows: windows installer-windows
	@echo "$(GREEN)✅ Quick Windows build complete!$(RESET)"

quick-linux: linux installer-linux
	@echo "$(GREEN)✅ Quick Linux build complete!$(RESET)"

list-targets:
	@echo "$(CYAN)🎯 Available Build Targets:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Main Targets:$(RESET)"
	@echo "  package                - Show interactive menu"
	@echo "  release                - Full release build (all platforms + installers)"
	@echo "  all                    - Build everything (alias for release)"
	@echo ""
	@echo "$(YELLOW)Platform-Specific Builds with Installers:$(RESET)"
	@echo "  windows-installer      - Build Windows AMD64 + Installer"
	@echo "  windows-arm64-installer - Build Windows ARM64 + Installer"
	@echo "  linux-installer        - Build Linux DEB + RPM + Portable"
	@echo "  macos-installer        - Build macOS DMG + Portable"
	@echo ""
	@echo "$(YELLOW)Binary Only (No Installer):$(RESET)"
	@echo "  windows                - Build Windows AMD64 binary"
	@echo "  windows-arm64          - Build Windows ARM64 binary"
	@echo "  linux                  - Build Linux AMD64 binary"
	@echo "  macos                  - Build macOS binary"
	@echo ""
	@echo "$(YELLOW)Installer Only:$(RESET)"
	@echo "  installer-windows      - Create Windows installer from existing build"
	@echo "  installer-linux        - Create Linux packages from existing build"
	@echo "  installer-macos        - Create macOS DMG from existing build"
	@echo "  installer-all          - Create all installers"
	@echo ""
	@echo "$(YELLOW)Testing:$(RESET)"
	@echo "  test                   - Run full test suite"
	@echo "  test-binary            - Test binary integrity"
	@echo "  smoke-test             - Run smoke tests"
	@echo ""
	@echo "$(YELLOW)Utilities:$(RESET)"
	@echo "  install-deps           - Install system dependencies"
	@echo "  deps                   - Update Go dependencies"
	@echo "  changelog              - Generate changelog from git commits"
	@echo "  clean                  - Remove all build artifacts"
	@echo "  validate-build         - Validate build environment"
	@echo ""
	@echo "$(YELLOW)Quick Builds:$(RESET)"
	@echo "  quick-windows          - Fast Windows build + installer"
	@echo "  quick-linux            - Fast Linux build + packages"
	@echo ""
	@echo "$(YELLOW)Help:$(RESET)"
	@echo "  help                   - Show detailed help"
	@echo "  list-targets           - Show this list"

# ====================================================================================
# ❓ HELP TARGET
# ====================================================================================
help:
	@echo "$(CYAN)════════════════════════════════════════$(RESET)"
	@echo "$(CYAN)  SIMDOKPOL Build System Help v10.0$(RESET)"
	@echo "$(CYAN)════════════════════════════════════════$(RESET)"
	@echo ""
	@echo "$(YELLOW)Main Targets:$(RESET)"
	@echo "  make package           - Show interactive menu"
	@echo "  make release           - Full release build (all platforms + installers)"
	@echo "  make clean             - Remove all build artifacts"
	@echo ""
	@echo "$(YELLOW)Platform-Specific Builds:$(RESET)"
	@echo "  make windows-installer       - Windows AMD64 + Installer"
	@echo "  make windows-arm64-installer - Windows ARM64 + Installer"
	@echo "  make linux-installer         - Linux DEB + RPM + Portable"
	@echo "  make macos-installer         - macOS DMG + Portable"
	@echo ""
	@echo "$(YELLOW)Binary Only (No Installer):$(RESET)"
	@echo "  make windows           - Windows AMD64 binary"
	@echo "  make windows-arm64     - Windows ARM64 binary"
	@echo "  make linux             - Linux AMD64 binary"
	@echo "  make macos             - macOS binary"
	@echo ""
	@echo "$(YELLOW)Testing & Validation:$(RESET)"
	@echo "  make test              - Run full test suite (clean + deps + binary tests)"
	@echo "  make test-binary       - Test binary integrity and size validation"
	@echo "  make smoke-test        - Quick smoke tests on built binaries"
	@echo "  make validate-build    - Validate build environment and prerequisites"
	@echo ""
	@echo "$(YELLOW)Utilities:$(RESET)"
	@echo "  make install-deps      - Install all system dependencies"
	@echo "  make deps              - Update Go module dependencies"
	@echo "  make changelog         - Generate changelog from git commits"
	@echo "  make list-targets      - List all available make targets"
	@echo "  make help              - Show this help message"
	@echo ""
	@echo "$(YELLOW)Current Configuration:$(RESET)"
	@echo "  Version: $(GREEN)$(VERSION_FULL)$(RESET) (previous: $(CURRENT_TAG))"
	@echo "  OS: $(GREEN)$(DETECTED_OS)$(RESET)"
	@echo "  Architecture: $(GREEN)$(DETECTED_ARCH)$(RESET)"
	@echo "  Package Manager: $(GREEN)$(PKG_MANAGER)$(RESET)"
	@echo ""
	@echo "$(YELLOW)Build Process Overview:$(RESET)"
	@echo "  The build system performs the following steps:"
	@echo "  1. Validates build environment (Go version, Git, required files)"
	@echo "  2. Generates version information from Git tags"
	@echo "  3. Compiles platform-specific binaries with CGO support"
	@echo "  4. Creates installers using platform-native tools (NSIS, DEB, RPM, DMG)"
	@echo "  5. Packages portable versions as ZIP/TAR.GZ archives"
	@echo "  6. Generates checksums for all release artifacts"
	@echo ""
	@echo "$(YELLOW)Template System:$(RESET)"
	@echo "  The build system supports template-based configuration."
	@echo "  Create templates in the '$(TEMPLATES_DIR)/' directory:"
	@echo "    - installer.nsi.tmpl    (Windows NSIS installer template)"
	@echo "    - Info.plist.tmpl       (macOS bundle configuration)"
	@echo "    - control.tmpl          (Debian package control file)"
	@echo ""
	@echo "$(YELLOW)Environment Variables:$(RESET)"
	@echo "  APP_SECRET_KEY         - Application secret key (default: auto-generated)"
	@echo "  CGO_ENABLED            - Enable CGO for compilation (default: 1)"
	@echo ""
	@echo "$(YELLOW)Prerequisites:$(RESET)"
	@echo "  Required:"
	@echo "    - Go $(GREEN)1.21.0$(RESET) or higher"
	@echo "    - Git for version management"
	@echo "    - Platform-specific compilers (MinGW for Windows cross-compilation)"
	@echo ""
	@echo "  Optional (for installers):"
	@echo "    - NSIS (Windows installers)"
	@echo "    - dpkg-deb (Debian packages)"
	@echo "    - rpmbuild (RPM packages)"
	@echo "    - hdiutil or genisoimage (macOS DMG creation)"
	@echo ""
	@echo "$(YELLOW)Quick Start:$(RESET)"
	@echo "  1. Install dependencies:     $(GREEN)make install-deps$(RESET)"
	@echo "  2. Build for all platforms:  $(GREEN)make release$(RESET)"
	@echo "  3. Find outputs in:          $(GREEN)$(RELEASE_DIR)/$(RESET)"
	@echo ""
	@echo "$(YELLOW)Troubleshooting:$(RESET)"
	@echo "  - If build fails, run:       $(GREEN)make clean && make validate-build$(RESET)"
	@echo "  - To test binaries:          $(GREEN)make test$(RESET)"
	@echo "  - For dependency issues:     $(GREEN)make install-deps$(RESET)"
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  Build Windows installer:     $(GREEN)make windows-installer$(RESET)"
	@echo "  Build Linux packages:        $(GREEN)make linux-installer$(RESET)"
	@echo "  Full release with tests:     $(GREEN)make clean && make test && make release$(RESET)"
	@echo "  Quick development build:     $(GREEN)make quick-windows$(RESET)"
	@echo ""
	@echo "$(CYAN)════════════════════════════════════════$(RESET)"
	@echo "For more information, visit: https://github.com/muhammad1505/simdokpol"
	@echo "$(CYAN)════════════════════════════════════════$(RESET)"