# ====================================================================================
# SIMDOKPOL - SISTEM BUILD YANG DISEMPURNAKAN v10.0 🚀
# Fitur: Auto-Version, Cross-Platform, Installer System, Template-Based Configuration
# Dukungan: Windows AMD64/ARM64, Linux AMD64, macOS
# ====================================================================================

# --- ⚙️ KONFIGURASI ---
APP_NAME := simdokpol
BUILD_DIR := build
RELEASE_DIR := release
TEMPLATES_DIR := templates
MAIN_FILE := cmd/main.go
RESOURCE_SYSO := cmd/resource.syso
ICON_PATH := web/static/img/icon.ico

# --- 🤖 LOGIKA AUTO VERSIONING ---
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
VERSION_NEXT := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | awk -F. -v OFS=. '{$NF+=1; print}')
VERSION_RAW := $(VERSION_NEXT)
VERSION_FULL := v$(VERSION_RAW)
VER_MAJOR := $(word 1,$(subst ., ,$(VERSION_RAW)))
VER_MINOR := $(word 2,$(subst ., ,$(VERSION_RAW)))
VER_PATCH := $(word 3,$(subst ., ,$(VERSION_RAW)))
VI_VERSION_FULL := $(VER_MAJOR).$(VER_MINOR).$(VER_PATCH).0

PREV_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "HEAD")

# --- 🔐 KUNCI RAHASIA ---
APP_SECRET_KEY ?= SIMDOKPOL_SECRET_KEY_2025

# --- LDFLAGS ---
LDFLAGS_COMMON := -w -s -X 'main.version=$(VERSION_FULL)'
LDFLAGS_APP    := $(LDFLAGS_COMMON) -X 'simdokpol/internal/services.AppSecretKeyString=$(APP_SECRET_KEY)'
LDFLAGS_TOOL   := $(LDFLAGS_COMMON) -X 'main.appSecretKey=$(APP_SECRET_KEY)'

# --- DETEKSI SISTEM OPERASI ---
DETECTED_OS := $(shell uname -s 2>/dev/null || echo "Unknown")
DETECTED_ARCH := $(shell uname -m 2>/dev/null || echo "Unknown")
PKG_MANAGER := $(shell if command -v pacman >/dev/null 2>&1; then echo "pacman"; elif command -v apt-get >/dev/null 2>&1; then echo "apt"; else echo "unknown"; fi)

# Warna untuk output terminal
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all package menu windows windows-arm64 linux macos tools changelog clean icon-gen deps check-deps install-deps installer-windows installer-windows-arm64 installer-linux installer-macos installer-all check-go check-git check-files validate-build test-binary smoke-test test help list-targets

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
	@echo "💻 Sistem Terdeteksi  : $(DETECTED_OS) $(DETECTED_ARCH)"
	@echo "📦 Package Manager    : $(PKG_MANAGER)"
	@echo "🔑 Kunci Rahasia Init : $(GREEN)$(APP_SECRET_KEY)$(RESET)"
	@echo "$(CYAN)--------------------------------------------------$(RESET)"
	@echo "Pilih Target Build:"
	@echo "  [1] 🚀  RELEASE LENGKAP (Semua Platform + Installer)"
	@echo "  [2] 🪟  Windows AMD64 (.exe + Installer)"
	@echo "  [3] 🔷  Windows ARM64 (Snapdragon) [Opsional]"
	@echo "  [4] 🐧  Linux AMD64 (DEB+RPM+Portable)"
	@echo "  [5] 🍎  macOS AMD64 (DMG+Portable)"
	@echo "  [6] 🛠️   Admin Tools (License Manager + Signer)"
	@echo "  [7] 📦  Install Dependencies"
	@echo "  [8] 📀  Build Semua Installer"
	@echo "  [9] 📝  Generate Changelog"
	@echo "  [t] 🧪  Jalankan Testing"
	@echo "  [h] ❓  Tampilkan Bantuan"
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
		*) echo "$(RED)Selesai!$(RESET)" ;; \
	esac

# ====================================================================================
# 🔍 PEMERIKSAAN PRASYARAT
# ====================================================================================
check-go:
	@command -v go >/dev/null 2>&1 || (echo "$(RED)❌ Go belum terinstall$(RESET)" && exit 1)
	@GO_VERSION=$$(go version | awk '{print $$3}' | sed 's/go//'); \
	REQUIRED_VERSION="1.23.0"; \
	if [ "$$(printf '%s\n' "$$REQUIRED_VERSION" "$$GO_VERSION" | sort -V | head -n1)" != "$$REQUIRED_VERSION" ]; then \
		echo "$(RED)❌ Versi Go $$GO_VERSION terlalu lama. Minimal diperlukan: $$REQUIRED_VERSION$(RESET)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Versi Go $$GO_VERSION OK$(RESET)"

check-git:
	@command -v git >/dev/null 2>&1 || (echo "$(RED)❌ Git belum terinstall$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Git OK$(RESET)"

check-files:
	@echo "$(YELLOW)🔍 Memvalidasi file dan direktori yang diperlukan...$(RESET)"
	@test -d web || (echo "$(RED)❌ Direktori web tidak ditemukan$(RESET)" && exit 1)
	@test -d migrations || (echo "$(YELLOW)⚠️  Direktori migrations tidak ditemukan (tidak kritis)$(RESET)")
	@test -f $(ICON_PATH) || (echo "$(YELLOW)⚠️  File ikon tidak ditemukan di $(ICON_PATH)$(RESET)")
	@test -f $(MAIN_FILE) || (echo "$(RED)❌ File utama tidak ditemukan di $(MAIN_FILE)$(RESET)" && exit 1)
	@test -f go.mod || (echo "$(RED)❌ go.mod tidak ditemukan$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Validasi file selesai$(RESET)"

check-nsis-deps:
	@if ! command -v makensis >/dev/null 2>&1; then \
		echo "$(RED)❌ NSIS tidak ditemukan$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			echo "$(YELLOW)Menginstall NSIS via pamac/yay...$(RESET)"; \
			if command -v pamac >/dev/null 2>&1; then \
				pamac install --no-confirm nsis || echo "$(YELLOW)⚠️  Gagal via pamac. Silakan install 'nsis' dari AUR manual (yay -S nsis).$(RESET)"; \
			elif command -v yay >/dev/null 2>&1; then \
				yay -S --noconfirm nsis || echo "$(YELLOW)⚠️  Gagal via yay. Silakan install 'nsis' dari AUR manual.$(RESET)"; \
			else \
				echo "$(YELLOW)⚠️  Silakan install 'nsis' dari AUR manual (yay -S nsis).$(RESET)"; \
			fi; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			echo "$(YELLOW)Menginstall NSIS via apt...$(RESET)"; \
			sudo apt-get update && sudo apt-get install -y nsis || (echo "$(RED)❌ Gagal menginstall NSIS$(RESET)" && exit 1); \
		else \
			echo "$(RED)❌ Silakan install NSIS secara manual$(RESET)"; \
			exit 1; \
		fi; \
	fi
	@echo "$(GREEN)✅ NSIS tersedia$(RESET)"

check-linux-installer-deps:
	@if ! command -v dpkg-deb >/dev/null 2>&1; then \
		echo "$(RED)❌ dpkg-deb tidak ditemukan$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			echo "$(YELLOW)Menginstall dpkg via pacman...$(RESET)"; \
			sudo pacman -S --needed dpkg || (echo "$(RED)❌ Gagal menginstall dpkg$(RESET)" && exit 1); \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			echo "$(YELLOW)Menginstall dpkg via apt...$(RESET)"; \
			sudo apt-get install -y dpkg || (echo "$(RED)❌ Gagal menginstall dpkg$(RESET)" && exit 1); \
		fi; \
	fi
	@echo "$(GREEN)✅ Dependensi installer Linux OK$(RESET)"

check-macos-deps:
	@if [ "$(DETECTED_OS)" != "Darwin" ] && ! command -v genisoimage >/dev/null 2>&1; then \
		echo "$(RED)❌ genisoimage tidak ditemukan$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			echo "$(YELLOW)Menginstall cdrkit via pacman...$(RESET)"; \
			sudo pacman -S --needed cdrkit || (echo "$(RED)❌ Gagal menginstall cdrkit$(RESET)" && exit 1); \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			echo "$(YELLOW)Menginstall genisoimage via apt...$(RESET)"; \
			sudo apt-get install -y genisoimage || (echo "$(RED)❌ Gagal menginstall genisoimage$(RESET)" && exit 1); \
		fi; \
	fi
	@echo "$(GREEN)✅ Dependensi build macOS OK$(RESET)"

validate-build: check-go check-git check-files
	@echo "$(GREEN)✅ Semua pemeriksaan validasi berhasil$(RESET)"

# ====================================================================================
# 🏗️ TARGET BUILD
# ====================================================================================
windows: validate-build icon-gen
	@echo "$(CYAN)🔨 Membangun Aplikasi Windows AMD64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE) || (echo "$(RED)❌ Build Windows gagal$(RESET)" && exit 1)
	@rm -f $(RESOURCE_SYSO)
	@test -f $(BUILD_DIR)/windows/$(APP_NAME).exe || (echo "$(RED)❌ Binary Windows tidak ditemukan setelah build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Build Windows AMD64 berhasil$(RESET)"

windows-arm64: validate-build
	@echo "$(CYAN)🔨 Membangun Aplikasi Windows ARM64 (Snapdragon)...$(RESET)"
	@if ! command -v aarch64-w64-mingw32-gcc >/dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Compiler ARM64 tidak ditemukan, melewati build Windows ARM64$(RESET)"; \
		echo "$(YELLOW)💡 Untuk menginstall: yay -S mingw-w64-gcc (pastikan mendukung ARM64)$(RESET)"; \
		mkdir -p $(BUILD_DIR)/windows-arm64; \
		touch $(BUILD_DIR)/windows-arm64/.skipped; \
	else \
		$(MAKE) windows-arm64-build; \
	fi

windows-arm64-build: icon-gen
	@mkdir -p $(BUILD_DIR)/windows-arm64
	@CGO_ENABLED=1 CC=aarch64-w64-mingw32-gcc GOOS=windows GOARCH=arm64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(MAIN_FILE) || (echo "$(RED)❌ Build Windows ARM64 gagal$(RESET)" && exit 1)
	@rm -f $(RESOURCE_SYSO)
	@test -f $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe || (echo "$(RED)❌ Binary Windows ARM64 tidak ditemukan setelah build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Build Windows ARM64 berhasil$(RESET)"

linux: validate-build
	@echo "$(CYAN)🔨 Membangun Aplikasi Linux AMD64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE) || (echo "$(RED)❌ Build Linux gagal$(RESET)" && exit 1)
	@chmod +x $(BUILD_DIR)/linux/$(APP_NAME)
	@test -f $(BUILD_DIR)/linux/$(APP_NAME) || (echo "$(RED)❌ Binary Linux tidak ditemukan setelah build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Build Linux AMD64 berhasil$(RESET)"

# Di bagian TARGET BUILD
macos: validate-build
	@echo "$(CYAN)🔨 Membangun Aplikasi macOS...$(RESET)"
	@if [ "$(DETECTED_OS)" != "Darwin" ]; then \
		if ! command -v clang >/dev/null 2>&1; then \
			echo "$(YELLOW)⚠️  Compiler clang tidak ditemukan, melewati build macOS$(RESET)"; \
			echo "$(YELLOW)💡 Build macOS dari Linux memerlukan OSXCross toolchain$(RESET)"; \
			echo "$(YELLOW)💡 Untuk informasi: https://github.com/tpoechtrager/osxcross$(RESET)"; \
			mkdir -p $(BUILD_DIR)/macos; \
			touch $(BUILD_DIR)/macos/.skipped; \
			exit 0; \
		else \
			$(MAKE) macos-build; \
		fi; \
	else \
		$(MAKE) macos-build; \
	fi

macos-build:
	@mkdir -p $(BUILD_DIR)/macos
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/macos/$(APP_NAME) $(MAIN_FILE) || (echo "$(RED)❌ Build macOS gagal$(RESET)" && exit 1)
	@chmod +x $(BUILD_DIR)/macos/$(APP_NAME)
	@test -f $(BUILD_DIR)/macos/$(APP_NAME) || (echo "$(RED)❌ Binary macOS tidak ditemukan setelah build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Build macOS berhasil$(RESET)"

# Di bagian TARGET INSTALLER
installer-macos: check-macos-deps
	@if [ -f $(BUILD_DIR)/macos/.skipped ]; then \
		echo "$(YELLOW)⚠️  Build macOS dilewati, melewati pembuatan installer$(RESET)"; \
		exit 0; \
	elif [ ! -f $(BUILD_DIR)/macos/$(APP_NAME) ]; then \
		echo "$(RED)❌ Binary macOS tidak tersedia. Jalankan 'make macos' terlebih dahulu$(RESET)"; \
		exit 1; \
	else \
		$(MAKE) installer-macos-build; \
	fi

# Di bagian akhir (setelah installer-macos-build)
macos-installer:
	@$(MAKE) macos 2>/dev/null || true
	@if [ -f $(BUILD_DIR)/macos/.skipped ]; then \
		echo "$(YELLOW)⚠️  Build macOS dilewati karena compiler tidak tersedia$(RESET)"; \
		echo "$(YELLOW)💡 Untuk build macOS dari Linux, install OSXCross toolchain$(RESET)"; \
		echo "$(YELLOW)💡 Info: https://github.com/tpoechtrager/osxcross$(RESET)"; \
		echo "$(YELLOW)📝 Tidak ada installer macOS yang dibuat$(RESET)"; \
	else \
		$(MAKE) installer-macos; \
		if [ -f $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-$(VERSION_FULL)-installer.dmg ] || [ -f $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-$(VERSION_FULL)-portable.zip ]; then \
			echo "$(GREEN)✅ macOS + Installer Selesai!$(RESET)"; \
		else \
			echo "$(YELLOW)⚠️  Installer macOS tidak berhasil dibuat$(RESET)"; \
		fi \
	fi

macos-build:
	@mkdir -p $(BUILD_DIR)/macos
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/macos/$(APP_NAME) $(MAIN_FILE) || (echo "$(RED)❌ Build macOS gagal$(RESET)" && exit 1)
	@chmod +x $(BUILD_DIR)/macos/$(APP_NAME)
	@test -f $(BUILD_DIR)/macos/$(APP_NAME) || (echo "$(RED)❌ Binary macOS tidak ditemukan setelah build$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Build macOS berhasil$(RESET)"

# ====================================================================================
# 📦 TARGET INSTALLER
# ====================================================================================
installer-windows: windows check-nsis-deps
	@echo "$(CYAN)📦 Membuat Installer Windows...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows $(RELEASE_DIR)
	
	@cp $(BUILD_DIR)/windows/$(APP_NAME).exe $(BUILD_DIR)/installer/windows/
	@cp -r web migrations $(BUILD_DIR)/installer/windows/ 2>/dev/null || echo "$(YELLOW)⚠️  Beberapa direktori tidak ditemukan$(RESET)"
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows/icon.ico 2>/dev/null || echo "$(YELLOW)⚠️  Ikon tidak ditemukan$(RESET)"
	@[ -f LICENSE ] && cp LICENSE $(BUILD_DIR)/installer/windows/ || echo "$(YELLOW)⚠️  LICENSE tidak ditemukan$(RESET)"
	
	@printf '@echo off\ntitle SIMDOKPOL Server\necho ========================================\necho   SIMDOKPOL - System Startup\necho ========================================\necho.\necho [INFO] Memulai SIMDOKPOL Server...\nstart simdokpol.exe\necho [SUCCESS] Server berjalan di background\n' > $(BUILD_DIR)/installer/windows/start.bat
	
	@if [ -f "$(TEMPLATES_DIR)/installer.nsi.tmpl" ]; then \
		echo "$(GREEN)✓ Menggunakan template dari $(TEMPLATES_DIR)/installer.nsi.tmpl$(RESET)"; \
		BANNER_DIRECTIVE=""; \
		HEADER_DIRECTIVE=""; \
		LICENSE_DIRECTIVE=""; \
		if [ -f "web/static/img/installer-banner.bmp" ]; then \
			cp web/static/img/installer-banner.bmp $(BUILD_DIR)/installer/windows/ 2>/dev/null || true; \
			BANNER_DIRECTIVE='!define MUI_WELCOMEFINISHPAGE_BITMAP "installer-banner.bmp"'; \
		fi; \
		if [ -f "web/static/img/installer-header.bmp" ]; then \
			cp web/static/img/installer-header.bmp $(BUILD_DIR)/installer/windows/ 2>/dev/null || true; \
			HEADER_DIRECTIVE='!define MUI_HEADERIMAGE\n!define MUI_HEADERIMAGE_BITMAP "installer-header.bmp"'; \
		fi; \
		if [ -f "LICENSE" ]; then \
			LICENSE_DIRECTIVE='!insertmacro MUI_PAGE_LICENSE "LICENSE"'; \
		fi; \
		sed -e 's|@APP_NAME@|$(APP_NAME)|g' \
		    -e 's|@VERSION@|$(VERSION_RAW)|g' \
		    -e 's|@PUBLISHER@|SIMDOKPOL Team|g' \
		    -e 's|@WEB_SITE@|https://github.com/muhammad1505/simdokpol|g' \
		    -e 's|@INSTALLER_NAME@|$(APP_NAME)-windows-x64-$(VERSION_FULL)-installer.exe|g' \
		    -e 's|@ICON_PATH@|icon.ico|g' \
		    -e 's|@VI_VERSION@|$(VI_VERSION_FULL)|g' \
		    -e "s|@BANNER_DIRECTIVE@|$$BANNER_DIRECTIVE|g" \
		    -e "s|@HEADER_DIRECTIVE@|$$HEADER_DIRECTIVE|g" \
		    -e "s|@LICENSE_DIRECTIVE@|$$LICENSE_DIRECTIVE|g" \
		    $(TEMPLATES_DIR)/installer.nsi.tmpl > $(BUILD_DIR)/installer/windows/installer.nsi; \
	else \
		echo "$(YELLOW)⚠️  Template tidak ditemukan, menggunakan skrip NSIS inline$(RESET)"; \
		printf '!define APP_NAME "SIMDOKPOL"\n!define VERSION "$(VERSION_RAW)"\n!define INSTALLER_NAME "$(APP_NAME)-windows-x64-$(VERSION_FULL)-installer.exe"\n\nName "$${APP_NAME} $${VERSION}"\nOutFile "$${INSTALLER_NAME}"\nInstallDir "$$PROGRAMFILES64\\$${APP_NAME}"\nRequestExecutionLevel admin\n\n!include "MUI2.nsh"\n!define MUI_ICON "icon.ico"\n!insertmacro MUI_PAGE_WELCOME\n!insertmacro MUI_PAGE_DIRECTORY\n!insertmacro MUI_PAGE_INSTFILES\n!insertmacro MUI_PAGE_FINISH\n!insertmacro MUI_LANGUAGE "English"\nVIProductVersion "$(VI_VERSION_FULL)"\n\nSection "Install"\n  SetOutPath "$$INSTDIR"\n  File "$(APP_NAME).exe"\n  File "icon.ico"\n  File "start.bat"\n' > $(BUILD_DIR)/installer/windows/installer.nsi; \
		if [ -d "$(BUILD_DIR)/installer/windows/web" ]; then printf '  File /r "web"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; fi; \
		if [ -d "$(BUILD_DIR)/installer/windows/migrations" ]; then printf '  File /r "migrations"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; fi; \
		if [ -f "$(BUILD_DIR)/installer/windows/LICENSE" ]; then printf '  File "LICENSE"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; fi; \
		printf '  WriteUninstaller "$$INSTDIR\\Uninstall.exe"\n  CreateShortcut "$$DESKTOP\\SIMDOKPOL.lnk" "$$INSTDIR\\start.bat" "" "$$INSTDIR\\icon.ico"\n  CreateDirectory "$$SMPROGRAMS\\SIMDOKPOL"\n  CreateShortcut "$$SMPROGRAMS\\SIMDOKPOL\\SIMDOKPOL.lnk" "$$INSTDIR\\start.bat" "" "$$INSTDIR\\icon.ico"\n  CreateShortcut "$$SMPROGRAMS\\SIMDOKPOL\\Uninstall.lnk" "$$INSTDIR\\Uninstall.exe"\nSectionEnd\n\nSection "Uninstall"\n  Delete "$$DESKTOP\\SIMDOKPOL.lnk"\n  RMDir /r "$$SMPROGRAMS\\SIMDOKPOL"\n  RMDir /r "$$INSTDIR"\nSectionEnd\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; \
	fi
	
	@cd $(BUILD_DIR)/installer/windows && makensis installer.nsi || (echo "$(RED)❌ Build NSIS gagal$(RESET)" && exit 1)
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/ || (echo "$(RED)❌ Gagal memindahkan installer$(RESET)" && exit 1)
	
	@cd $(BUILD_DIR)/installer/windows && zip -r $(APP_NAME)-windows-x64-$(VERSION_FULL)-portable.zip . -x "installer.nsi" "*.exe" || (echo "$(RED)❌ Gagal membuat paket portable$(RESET)" && exit 1)
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/
	
	@echo "$(GREEN)✅ Installer Windows berhasil dibuat$(RESET)"

installer-windows-arm64: windows-arm64 check-nsis-deps
	@if [ -f $(BUILD_DIR)/windows-arm64/.skipped ]; then \
		echo "$(YELLOW)⚠️  Build Windows ARM64 dilewati, melewati pembuatan installer$(RESET)"; \
	elif [ ! -f $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe ]; then \
		echo "$(YELLOW)⚠️  Binary Windows ARM64 tidak tersedia, melewati pembuatan installer$(RESET)"; \
	else \
		$(MAKE) installer-windows-arm64-build; \
	fi

installer-windows-arm64-build:
	@echo "$(CYAN)📦 Membuat Installer Windows ARM64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows-arm64 $(RELEASE_DIR)
	@cp $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(BUILD_DIR)/installer/windows-arm64/
	@cp -r web migrations $(BUILD_DIR)/installer/windows-arm64/ 2>/dev/null || true
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows-arm64/icon.ico 2>/dev/null || true
	@[ -f LICENSE ] && cp LICENSE $(BUILD_DIR)/installer/windows-arm64/ || true
	@printf '@echo off\ntitle SIMDOKPOL Server\necho ========================================\necho   SIMDOKPOL - System Startup\necho ========================================\necho.\necho [INFO] Memulai SIMDOKPOL Server...\nstart simdokpol.exe\necho [SUCCESS] Server berjalan di background\n' > $(BUILD_DIR)/installer/windows-arm64/start.bat
	@printf '!define APP_NAME "SIMDOKPOL"\n!define VERSION "$(VERSION_RAW)"\n!define ARCH "ARM64"\n!define INSTALLER_NAME "$(APP_NAME)-windows-arm64-$(VERSION_FULL)-installer.exe"\n\nName "${APP_NAME} ${VERSION} (${ARCH})"\nOutFile "${INSTALLER_NAME}"\nInstallDir "$PROGRAMFILES64\\${APP_NAME}"\nRequestExecutionLevel admin\n\n!include "MUI2.nsh"\n!define MUI_ICON "icon.ico"\n!insertmacro MUI_PAGE_WELCOME\n!insertmacro MUI_PAGE_DIRECTORY\n!insertmacro MUI_PAGE_INSTFILES\n!insertmacro MUI_PAGE_FINISH\n!insertmacro MUI_LANGUAGE "English"\nVIProductVersion "$(VI_VERSION_FULL)"\nVIAddVersionKey "ProductName" "${APP_NAME}"\nVIAddVersionKey "FileVersion" "${VERSION}"\n\nSection "Install"\n  SetOutPath "$INSTDIR"\n  File "$(APP_NAME).exe"\n  File "icon.ico"\n  File "start.bat"\n' > $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@if [ -d "$(BUILD_DIR)/installer/windows-arm64/web" ]; then printf '  File /r "web"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi; fi
	@if [ -d "$(BUILD_DIR)/installer/windows-arm64/migrations" ]; then printf '  File /r "migrations"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi; fi
	@if [ -f "$(BUILD_DIR)/installer/windows-arm64/LICENSE" ]; then printf '  File "LICENSE"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi; fi
	@printf '  WriteUninstaller "$INSTDIR\\Uninstall.exe"\n  CreateShortcut "$DESKTOP\\SIMDOKPOL.lnk" "$INSTDIR\\start.bat" "" "$INSTDIR\\icon.ico"\n  CreateDirectory "$SMPROGRAMS\\SIMDOKPOL"\n  CreateShortcut "$SMPROGRAMS\\SIMDOKPOL\\SIMDOKPOL.lnk" "$INSTDIR\\start.bat" "" "$INSTDIR\\icon.ico"\n  CreateShortcut "$SMPROGRAMS\\SIMDOKPOL\\Uninstall.lnk" "$INSTDIR\\Uninstall.exe"\nSectionEnd\n\nSection "Uninstall"\n  Delete "$DESKTOP\\SIMDOKPOL.lnk"\n  RMDir /r "$SMPROGRAMS\\SIMDOKPOL"\n  RMDir /r "$INSTDIR"\nSectionEnd\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@cd $(BUILD_DIR)/installer/windows-arm64 && makensis installer.nsi || (echo "$(RED)❌ Build NSIS gagal$(RESET)" && exit 1)
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/ 2>/dev/null || true
	@cd $(BUILD_DIR)/installer/windows-arm64 && zip -r $(APP_NAME)-windows-arm64-$(VERSION_FULL)-portable.zip . -x "installer.nsi" "*.exe"
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || true
	@echo "$(GREEN)✅ Installer Windows ARM64 berhasil dibuat$(RESET)"

installer-linux: linux check-linux-installer-deps
	@echo "$(CYAN)📦 Membuat Installer Linux...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux $(RELEASE_DIR)
	@cp $(BUILD_DIR)/linux/$(APP_NAME) $(BUILD_DIR)/installer/linux/
	@cp -r web migrations $(BUILD_DIR)/installer/linux/ 2>/dev/null || true
	@[ -f LICENSE ] && cp LICENSE $(BUILD_DIR)/installer/linux/ || true
	@printf '#!/bin/bash\necho "========================================"\necho "  SIMDOKPOL - System Startup"\necho "========================================"\necho ""\necho "[INFO] Memulai SIMDOKPOL Server..."\n./simdokpol\n' > $(BUILD_DIR)/installer/linux/start.sh
	@chmod +x $(BUILD_DIR)/installer/linux/start.sh
	@echo "$(YELLOW)Membangun paket DEB...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/DEBIAN
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/bin
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/share/applications
	@cp -r $(BUILD_DIR)/installer/linux/* $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)/ 2>/dev/null || true
	@rm -rf $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)/deb 2>/dev/null || true
	@printf '#!/bin/sh\n/opt/$(APP_NAME)/$(APP_NAME) "$$@"\n' > $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	@chmod +x $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	@printf '[Desktop Entry]\nVersion=1.0\nType=Application\nName=SIMDOKPOL\nComment=Sistem Informasi Manajemen Dokumen Kepolisian\nExec=/opt/$(APP_NAME)/start.sh\nTerminal=false\nCategories=Office;Database;\n' > $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Package: $(APP_NAME)\nVersion: $(VERSION_RAW)\nSection: utils\nPriority: optional\nArchitecture: amd64\nMaintainer: SIMDOKPOL Team\nDescription: Sistem Informasi Manajemen Dokumen Kepolisian\n Aplikasi manajemen dokumen untuk kepolisian\n' > $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf '#!/bin/bash\nset -e\nchmod +x "/opt/$(APP_NAME)/$(APP_NAME)"\nchmod +x "/opt/$(APP_NAME)/start.sh"\nexit 0\n' > $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	@chmod 755 $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	@dpkg-deb --root-owner-group --build $(BUILD_DIR)/installer/linux/deb $(RELEASE_DIR)/$(APP_NAME)_$(VERSION_RAW)_amd64.deb || (echo "$(RED)❌ Gagal membuat DEB$(RESET)" && exit 1)
	@cd $(BUILD_DIR)/installer/linux && tar --exclude='deb' -czf $(APP_NAME)-linux-amd64-$(VERSION_FULL)-portable.tar.gz *
	@mv $(BUILD_DIR)/installer/linux/$(APP_NAME)-linux-amd64-$(VERSION_FULL)-portable.tar.gz $(RELEASE_DIR)/ 2>/dev/null || true
	@echo "$(GREEN)✅ Installer Linux berhasil dibuat$(RESET)"

installer-macos: macos check-macos-deps
	@if [ -f $(BUILD_DIR)/macos/.skipped ]; then \
		echo "$(YELLOW)⚠️  Build macOS dilewati, melewati pembuatan installer$(RESET)"; \
	elif [ ! -f $(BUILD_DIR)/macos/$(APP_NAME) ]; then \
		echo "$(YELLOW)⚠️  Binary macOS tidak tersedia, melewati pembuatan installer$(RESET)"; \
	else \
		$(MAKE) installer-macos-build; \
	fi

installer-macos-build:
	@echo "$(CYAN)📦 Membuat Installer macOS...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos $(RELEASE_DIR)
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/backups
	@cp $(BUILD_DIR)/macos/$(APP_NAME) $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@cp -r web migrations $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/ 2>/dev/null || true
	@printf '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n\t<key>CFBundleExecutable</key>\n\t<string>$(APP_NAME)</string>\n\t<key>CFBundleIdentifier</key>\n\t<string>com.simdokpol.app</string>\n\t<key>CFBundleName</key>\n\t<string>$(APP_NAME)</string>\n\t<key>CFBundlePackageType</key>\n\t<string>APPL</string>\n\t<key>CFBundleShortVersionString</key>\n\t<string>$(VERSION_RAW)</string>\n\t<key>CFBundleVersion</key>\n\t<string>$(VERSION_RAW)</string>\n\t<key>LSMinimumSystemVersion</key>\n\t<string>10.13</string>\n\t<key>NSHighResolutionCapable</key>\n\t<true/>\n\t<key>CFBundleDisplayName</key>\n\t<string>$(APP_NAME)</string>\n</dict>\n</plist>\n' > $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@echo "$(YELLOW)Membuat installer DMG...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos/dmg-contents
	@cp -r $(BUILD_DIR)/installer/macos/$(APP_NAME).app $(BUILD_DIR)/installer/macos/dmg-contents/
	@ln -sf /Applications $(BUILD_DIR)/installer/macos/dmg-contents/Applications 2>/dev/null || true
	@if [ "$(DETECTED_OS)" = "Darwin" ]; then \
		hdiutil create -volname "SIMDOKPOL" -srcfolder $(BUILD_DIR)/installer/macos/dmg-contents -ov -format UDZO $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-$(VERSION_FULL)-installer.dmg; \
	else \
		echo "$(YELLOW)Menggunakan genisoimage untuk pembuatan DMG di Linux...$(RESET)"; \
		genisoimage -V "SIMDOKPOL" -D -R -apple -no-pad -o $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-$(VERSION_FULL)-installer.dmg $(BUILD_DIR)/installer/macos/dmg-contents 2>/dev/null || echo "$(RED)❌ Gagal membuat DMG$(RESET)"; \
	fi
	@cd $(BUILD_DIR)/installer/macos && zip -r $(APP_NAME)-macos-amd64-$(VERSION_FULL)-portable.zip $(APP_NAME).app
	@mv $(BUILD_DIR)/installer/macos/$(APP_NAME)-macos-amd64-$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || true
	@echo "$(GREEN)✅ Installer macOS berhasil dibuat$(RESET)"

installer-all: installer-windows installer-windows-arm64 installer-linux installer-macos
	@echo "$(GREEN)✅ Semua installer berhasil dibuat!$(RESET)"

windows-installer: windows installer-windows
	@echo "$(GREEN)✅ Windows + Installer Selesai!$(RESET)"

windows-arm64-installer: windows-arm64 installer-windows-arm64
	@echo "$(GREEN)✅ Windows ARM64 + Installer Selesai!$(RESET)"

linux-installer: linux installer-linux
	@echo "$(GREEN)✅ Linux + Installer Selesai!$(RESET)"

macos-installer: macos installer-macos
	@echo "$(GREEN)✅ macOS + Installer Selesai!$(RESET)"

# ====================================================================================
# 🛠️ TOOLS & TESTING
# ====================================================================================
tools: validate-build
	@echo "$(CYAN)🛠️  Membangun Admin Tools...$(RESET)"
	@mkdir -p $(BUILD_DIR)/tools
	@echo "   > Membangun License Manager (GUI)..."
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_TOOL) -H=windowsgui" \
	-o $(BUILD_DIR)/tools/LicenseManager.exe cmd/license-manager/main.go || echo "$(YELLOW)⚠️  License Manager tidak tersedia$(RESET)"
	@echo "   > Membangun Signer CLI..."
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_TOOL)" \
	-o $(BUILD_DIR)/tools/SignerCLI.exe cmd/signer/main.go || echo "$(YELLOW)⚠️  Signer CLI tidak tersedia$(RESET)"
	@echo "$(GREEN)✅ Build Admin Tools Selesai$(RESET)"

test-binary:
	@echo "$(YELLOW)🧪 Menguji integritas binary...$(RESET)"
	@if [ -f $(BUILD_DIR)/windows/$(APP_NAME).exe ]; then \
		FILE_SIZE=$(stat -c%s "$(BUILD_DIR)/windows/$(APP_NAME).exe" 2>/dev/null || stat -f%z "$(BUILD_DIR)/windows/$(APP_NAME).exe" 2>/dev/null); \
		if [ $FILE_SIZE -lt 1000000 ]; then \
			echo "$(RED)❌ Ukuran binary Windows terlalu kecil: $FILE_SIZE bytes$(RESET)"; \
			exit 1; \
		fi; \
		echo "$(GREEN)✓ Binary Windows OK ($FILE_SIZE bytes)$(RESET)"; \
	fi
	@if [ -f $(BUILD_DIR)/linux/$(APP_NAME) ]; then \
		FILE_SIZE=$(stat -c%s "$(BUILD_DIR)/linux/$(APP_NAME)" 2>/dev/null || stat -f%z "$(BUILD_DIR)/linux/$(APP_NAME)" 2>/dev/null); \
		if [ $FILE_SIZE -lt 1000000 ]; then \
			echo "$(RED)❌ Ukuran binary Linux terlalu kecil: $FILE_SIZE bytes$(RESET)"; \
			exit 1; \
		fi; \
		echo "$(GREEN)✓ Binary Linux OK ($FILE_SIZE bytes)$(RESET)"; \
	fi
	@echo "$(GREEN)✅ Validasi binary berhasil$(RESET)"

smoke-test: test-binary
	@echo "$(YELLOW)🔥 Menjalankan smoke tests...$(RESET)"
	@if [ -f $(BUILD_DIR)/windows/$(APP_NAME).exe ]; then \
		echo "$(YELLOW)Menguji flag versi binary Windows...$(RESET)"; \
		timeout 5 $(BUILD_DIR)/windows/$(APP_NAME).exe --version 2>/dev/null || echo "$(YELLOW)⚠️  Pemeriksaan versi tidak tersedia atau timeout$(RESET)"; \
	fi
	@if [ -f $(BUILD_DIR)/linux/$(APP_NAME) ]; then \
		echo "$(YELLOW)Menguji flag versi binary Linux...$(RESET)"; \
		timeout 5 $(BUILD_DIR)/linux/$(APP_NAME) --version 2>/dev/null || echo "$(YELLOW)⚠️  Pemeriksaan versi tidak tersedia atau timeout$(RESET)"; \
	fi
	@echo "$(GREEN)✅ Smoke tests selesai$(RESET)"

test: clean deps test-binary smoke-test
	@echo "$(GREEN)✅ Semua testing berhasil$(RESET)"

changelog:
	@echo "$(YELLOW)📝 Menghasilkan Changelog...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@printf "CHANGELOG - $(APP_NAME) $(VERSION_FULL)\n\n" > $(BUILD_DIR)/CHANGELOG.txt
	@git log $(PREV_VERSION)..HEAD --pretty=format:"- %s" >> $(BUILD_DIR)/CHANGELOG.txt 2>/dev/null || printf "- Rilis awal\n" >> $(BUILD_DIR)/CHANGELOG.txt
	@printf "\n" >> $(BUILD_DIR)/CHANGELOG.txt
	@echo "$(GREEN)✅ Changelog dibuat di $(BUILD_DIR)/CHANGELOG.txt$(RESET)"

icon-gen:
	@echo "$(YELLOW)🖼️  Menghasilkan Metadata Windows...$(RESET)"
	@if ! command -v goversioninfo >/dev/null 2>&1; then \
		echo "$(YELLOW)Menginstall goversioninfo...$(RESET)"; \
		go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest || (echo "$(RED)❌ Gagal menginstall goversioninfo$(RESET)" && exit 1); \
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
	@goversioninfo -o $(RESOURCE_SYSO) || (echo "$(RED)❌ Gagal menghasilkan resource$(RESET)" && exit 1)
	@rm -f versioninfo.json
	@echo "$(GREEN)✅ Metadata Windows dihasilkan$(RESET)"

deps:
	@echo "$(YELLOW)🔍 Memeriksa Dependensi...$(RESET)"
	@go mod tidy || (echo "$(RED)❌ Gagal merapikan modul Go$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Dependensi OK$(RESET)"

clean:
	@echo "$(YELLOW)🧹 Membersihkan artifak build...$(RESET)"
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR) $(RESOURCE_SYSO) versioninfo.json
	@echo "$(GREEN)✅ Pembersihan selesai!$(RESET)"

# ====================================================================================
# 📦 INSTALASI DEPENDENSI
# ====================================================================================
install-deps:
	@echo "$(CYAN)📦 Menginstall Dependensi untuk $(DETECTED_OS) ($(PKG_MANAGER))...$(RESET)"
	@if [ "$(PKG_MANAGER)" = "pacman" ]; then \
		echo "$(YELLOW)Menginstall untuk Arch/Manjaro...$(RESET)"; \
		sudo pacman -S --needed base-devel mingw-w64-gcc go git zip unzip gtk3 webkit2gtk dpkg cdrkit rpm-org || (echo "$(RED)❌ Instalasi gagal$(RESET)" && exit 1); \
		if ! command -v makensis >/dev/null 2>&1; then \
			echo "$(YELLOW)⚠️  NSIS tidak ditemukan. Mencoba install via pamac/yay...$(RESET)"; \
			if command -v pamac >/dev/null 2>&1; then \
				pamac install --no-confirm nsis || echo "$(YELLOW)⚠️  Pamac gagal. Install 'nsis' manual via AUR (yay -S nsis)$(RESET)"; \
			elif command -v yay >/dev/null 2>&1; then \
				yay -S --noconfirm nsis || echo "$(YELLOW)⚠️  Yay gagal. Install 'nsis' manual dari AUR$(RESET)"; \
			else \
				echo "$(YELLOW)⚠️  Silakan install 'nsis' manual dari AUR (yay -S nsis)$(RESET)"; \
			fi; \
		fi; \
		echo "$(GREEN)✅ Dependensi Arch/Manjaro terinstall!$(RESET)"; \
	elif [ "$(PKG_MANAGER)" = "apt" ]; then \
		echo "$(YELLOW)Menginstall untuk Ubuntu/Debian...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y build-essential gcc-mingw-w64-x86-64 gcc-mingw-w64-aarch64 golang-go git zip unzip libgtk-3-dev libwebkit2gtk-4.0-dev nsis dpkg rpm genisoimage || (echo "$(RED)❌ Instalasi gagal$(RESET)" && exit 1); \
		echo "$(GREEN)✅ Dependensi Ubuntu/Debian terinstall!$(RESET)"; \
	else \
		echo "$(RED)❌ Package manager tidak didukung: $(PKG_MANAGER)$(RESET)"; \
		echo "$(YELLOW)Silakan install manual: build-essential, mingw-w64, go, git, zip, nsis, dpkg, rpm, genisoimage$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)📦 Menginstall Go tools...$(RESET)"
	@go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest || echo "$(YELLOW)⚠️  Tidak dapat menginstall goversioninfo$(RESET)"
	@echo "$(GREEN)✅ Semua dependensi berhasil terinstall!$(RESET)"

release: clean deps changelog installer-all
	@echo "$(GREEN)✅ RELEASE SELESAI! Semua installer tersedia di '$(RELEASE_DIR)/'$(RESET)"
	@echo "$(YELLOW)📁 File yang dihasilkan:$(RESET)"
	@ls -lh $(RELEASE_DIR)/* 2>/dev/null || echo "$(RED)Tidak ada file yang dihasilkan$(RESET)"

all: clean deps installer-all
	@echo "$(GREEN)🎉 Semua build selesai dengan sukses!$(RESET)"

quick-windows: windows installer-windows
	@echo "$(GREEN)✅ Build cepat Windows selesai!$(RESET)"

quick-linux: linux installer-linux
	@echo "$(GREEN)✅ Build cepat Linux selesai!$(RESET)"

help:
	@echo "$(CYAN)════════════════════════════════════════════════════$(RESET)"
	@echo "$(CYAN)  📖 BANTUAN SIMDOKPOL BUILD SYSTEM v10.0$(RESET)"
	@echo "$(CYAN)════════════════════════════════════════════════════$(RESET)"
	@echo ""
	@echo "$(YELLOW)Target Utama:$(RESET)"
	@echo "  make package           - Tampilkan menu interaktif"
	@echo "  make release           - Build release lengkap (semua platform)"
	@echo "  make all               - Alias untuk release"
	@echo ""
	@echo "$(YELLOW)Build Spesifik Platform dengan Installer:$(RESET)"
	@echo "  make windows-installer      - Build Windows AMD64 + Installer"
	@echo "  make windows-arm64-installer - Build Windows ARM64 + Installer"
	@echo "  make linux-installer        - Build Linux DEB + RPM + Portable"
	@echo "  make macos-installer        - Build macOS DMG + Portable"
	@echo ""
	@echo "$(YELLOW)Binary Saja (Tanpa Installer):$(RESET)"
	@echo "  make windows           - Build binary Windows AMD64"
	@echo "  make windows-arm64     - Build binary Windows ARM64"
	@echo "  make linux             - Build binary Linux AMD64"
	@echo "  make macos             - Build binary macOS"
	@echo ""
	@echo "$(YELLOW)Testing & Validasi:$(RESET)"
	@echo "  make test              - Jalankan test suite lengkap"
	@echo "  make test-binary       - Test integritas binary"
	@echo "  make smoke-test        - Jalankan smoke tests"
	@echo ""
	@echo "$(YELLOW)Utilitas:$(RESET)"
	@echo "  make install-deps      - Install dependensi sistem"
	@echo "  make deps              - Update dependensi Go"
	@echo "  make changelog         - Hasilkan changelog dari git commits"
	@echo "  make clean             - Bersihkan artifak build"
	@echo "  make tools             - Build admin tools"
	@echo ""
	@echo "$(YELLOW)Konfigurasi Saat Ini:$(RESET)"
	@echo "  Versi: $(GREEN)$(VERSION_FULL)$(RESET) (sebelumnya: $(CURRENT_TAG))"
	@echo "  OS: $(GREEN)$(DETECTED_OS)$(RESET)"
	@echo "  Arsitektur: $(GREEN)$(DETECTED_ARCH)$(RESET)"
	@echo "  Package Manager: $(GREEN)$(PKG_MANAGER)$(RESET)"
	@echo ""
	@echo "$(YELLOW)Contoh Penggunaan:$(RESET)"
	@echo "  Build installer Windows:  $(GREEN)make windows-installer$(RESET)"
	@echo "  Build paket Linux:        $(GREEN)make linux-installer$(RESET)"
	@echo "  Release lengkap + tests:  $(GREEN)make clean && make test && make release$(RESET)"
	@echo "  Build development cepat:  $(GREEN)make quick-windows$(RESET)"
	@echo ""
	@echo "$(CYAN)════════════════════════════════════════════════════$(RESET)"
	@echo "Informasi lebih lanjut: https://github.com/muhammad1505/simdokpol"
	@echo "$(CYAN)════════════════════════════════════════════════════$(RESET)"

list-targets: help