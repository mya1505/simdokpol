# ====================================================================================
# SIMDOKPOL - ULTIMATE BUILD SYSTEM (DYNAMIC VERSION) 🚀
# Fitur: Auto-Version, Cross-Platform, Installer System
# Support: Windows AMD64/ARM64, Linux AMD64, macOS
# ====================================================================================

# --- ⚙️ KONFIGURASI ---
APP_NAME := simdokpol
BUILD_DIR := build
RELEASE_DIR := release
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

# Warna
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all package menu windows windows-arm64 linux macos tools changelog clean icon-gen deps check-deps install-deps installer-windows installer-linux installer-macos installer-all check-go check-git

# ====================================================================================
# 🎮 MENU
# ====================================================================================
package:
	@clear
	@echo "$(CYAN)==================================================$(RESET)"
	@echo "$(CYAN)   👮 SIMDOKPOL BUILDER v9.0 (Installer Edition) $(RESET)"
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
		*) echo "$(RED)Bye!$(RESET)" ;; \
	esac

# --- PREREQUISITE CHECKS ---
check-go:
	@command -v go >/dev/null 2>&1 || (echo "$(RED)❌ Go is not installed$(RESET)" && exit 1)

check-git:
	@command -v git >/dev/null 2>&1 || (echo "$(RED)❌ Git is not installed$(RESET)" && exit 1)

check-deps: check-go check-git
	@echo "$(GREEN)✅ Basic dependencies OK$(RESET)"

# --- RELEASE PIPELINE ---
release: clean check-deps deps changelog installer-all
	@echo "$(GREEN)✅ RELEASE SELESAI! Semua installer tersedia di '$(RELEASE_DIR)/'$(RESET)"
	@echo "$(YELLOW)📁 File yang dihasilkan:$(RESET)"
	@ls -lh $(RELEASE_DIR)/* 2>/dev/null || echo "No files generated"

# --- INSTALLER TARGETS ---
installer-all: installer-windows installer-windows-arm64 installer-linux installer-macos
	@echo "$(GREEN)✅ Semua installer berhasil dibuat!$(RESET)"

installer-windows: windows check-nsis-deps
	@echo "$(CYAN)📦 Creating Windows Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows $(RELEASE_DIR)
	
	@# Prepare files for installer
	@cp $(BUILD_DIR)/windows/$(APP_NAME).exe $(BUILD_DIR)/installer/windows/
	@cp -r web migrations $(BUILD_DIR)/installer/windows/ 2>/dev/null || echo "$(YELLOW)Warning: web/migrations not found$(RESET)"
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows/icon.ico 2>/dev/null || echo "$(YELLOW)Warning: Icon not found$(RESET)"
	@cp LICENSE $(BUILD_DIR)/installer/windows/ 2>/dev/null || echo "$(YELLOW)Warning: LICENSE not found$(RESET)"
	
	@# Create batch file with proper escaping
	@printf '@echo off\n' > $(BUILD_DIR)/installer/windows/start.bat
	@printf 'title SIMDOKPOL Server\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo ========================================\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo   SIMDOKPOL - System Startup\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo ========================================\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo.\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo [INFO] Memulai SIMDOKPOL Server...\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo [INFO] Aplikasi akan terbuka otomatis di browser\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo.\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'start simdokpol.exe\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo [SUCCESS] Server berjalan di background\n' >> $(BUILD_DIR)/installer/windows/start.bat
	@printf 'echo.\n' >> $(BUILD_DIR)/installer/windows/start.bat
	
	@# Create NSIS script using printf to avoid heredoc issues
	@printf '!define APP_NAME "SIMDOKPOL"\n' > $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define VERSION "%s"\n' "$(VERSION_RAW)" >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define PUBLISHER "SIMDOKPOL Team"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define WEB_SITE "https://github.com/muhammad1505/simdokpol"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define INSTALLER_NAME "%s-windows-x64-v%s-installer.exe"\n' "$(APP_NAME)" "$(VERSION_FULL)" >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'Name "$${APP_NAME} $${VERSION}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'OutFile "$${INSTALLER_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'InstallDir "$$PROGRAMFILES64\\$${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'RequestExecutionLevel admin\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!include "MUI2.nsh"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define MUI_ICON "icon.ico"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define MUI_UNICON "icon.ico"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_PAGE_WELCOME\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@if [ -f "LICENSE" ]; then \
		printf '!insertmacro MUI_PAGE_LICENSE "LICENSE"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; \
	fi
	@printf '!insertmacro MUI_PAGE_DIRECTORY\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_PAGE_INSTFILES\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define MUI_FINISHPAGE_RUN "$$INSTDIR\\%s.exe"\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!define MUI_FINISHPAGE_RUN_TEXT "Jalankan $${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_PAGE_FINISH\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_UNPAGE_CONFIRM\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_UNPAGE_INSTFILES\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_UNPAGE_FINISH\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '!insertmacro MUI_LANGUAGE "English"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'Section "Install"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  SetOutPath "$$INSTDIR"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  File "%s.exe"\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  File "icon.ico"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@if [ -d "web" ]; then printf '  File /r "web"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; fi
	@if [ -d "migrations" ]; then printf '  File /r "migrations"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi; fi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  CreateDirectory "$$INSTDIR\\backups"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteUninstaller "$$INSTDIR\\Uninstall.exe"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}" "DisplayName" "$${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}" "UninstallString" "$$INSTDIR\\Uninstall.exe"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}" "DisplayIcon" "$$INSTDIR\\icon.ico"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}" "Publisher" "$${PUBLISHER}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}" "DisplayVersion" "$${VERSION}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}" "URLInfoAbout" "$${WEB_SITE}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  CreateShortcut "$$DESKTOP\\$${APP_NAME}.lnk" "$$INSTDIR\\%s.exe" "" "$$INSTDIR\\icon.ico" 0\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  CreateDirectory "$$SMPROGRAMS\\$${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  CreateShortcut "$$SMPROGRAMS\\$${APP_NAME}\\$${APP_NAME}.lnk" "$$INSTDIR\\%s.exe" "" "$$INSTDIR\\icon.ico" 0\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  CreateShortcut "$$SMPROGRAMS\\$${APP_NAME}\\Uninstall.lnk" "$$INSTDIR\\Uninstall.exe" "" "$$INSTDIR\\icon.ico" 0\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'SectionEnd\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'Section "Uninstall"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  Delete "$$DESKTOP\\$${APP_NAME}.lnk"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  RMDir /r "$$SMPROGRAMS\\$${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  DeleteRegKey HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  \n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf '  RMDir /r "$$INSTDIR"\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	@printf 'SectionEnd\n' >> $(BUILD_DIR)/installer/windows/installer.nsi
	
	@# Build NSIS installer
	@cd $(BUILD_DIR)/installer/windows && makensis installer.nsi
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-v$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/ 2>/dev/null || echo "$(RED)Failed to create installer$(RESET)"
	
	@# Create portable package
	@cd $(BUILD_DIR)/installer/windows && zip -r $(APP_NAME)-windows-x64-v$(VERSION_FULL)-portable.zip . -x "installer.nsi" "*.exe"
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || echo "$(RED)Failed to create portable package$(RESET)"
	
	@echo "$(GREEN)✅ Windows Installer created!$(RESET)"

installer-windows-arm64: windows-arm64 check-nsis-deps
	@echo "$(CYAN)📦 Creating Windows ARM64 Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows-arm64 $(RELEASE_DIR)
	
	@cp $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(BUILD_DIR)/installer/windows-arm64/
	@cp -r web migrations $(BUILD_DIR)/installer/windows-arm64/ 2>/dev/null || true
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows-arm64/icon.ico 2>/dev/null || true
	
	@# Create NSIS script for ARM64 (similar process as AMD64)
	@printf '!define APP_NAME "SIMDOKPOL"\n' > $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!define VERSION "%s"\n' "$(VERSION_RAW)" >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!define ARCH "ARM64"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!define INSTALLER_NAME "%s-windows-arm64-v%s-installer.exe"\n' "$(APP_NAME)" "$(VERSION_FULL)" >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'Name "$${APP_NAME} $${VERSION} ($${ARCH})"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'OutFile "$${INSTALLER_NAME}"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'InstallDir "$$PROGRAMFILES64\\$${APP_NAME}"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'RequestExecutionLevel admin\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!include "MUI2.nsh"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!define MUI_ICON "icon.ico"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!insertmacro MUI_PAGE_WELCOME\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!insertmacro MUI_PAGE_DIRECTORY\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!insertmacro MUI_PAGE_INSTFILES\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!insertmacro MUI_PAGE_FINISH\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '!insertmacro MUI_LANGUAGE "English"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'Section "Install"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '  SetOutPath "$$INSTDIR"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '  File "%s.exe"\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@if [ -d "web" ]; then printf '  File /r "web"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi; fi
	@printf '  WriteUninstaller "$$INSTDIR\\Uninstall.exe"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'SectionEnd\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'Section "Uninstall"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf '  RMDir /r "$$INSTDIR"\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	@printf 'SectionEnd\n' >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi
	
	@cd $(BUILD_DIR)/installer/windows-arm64 && makensis installer.nsi
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/ 2>/dev/null || true
	
	@cd $(BUILD_DIR)/installer/windows-arm64 && zip -r $(APP_NAME)-windows-arm64-v$(VERSION_FULL)-portable.zip . -x "installer.nsi" "*.exe"
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || true
	
	@echo "$(GREEN)✅ Windows ARM64 Installer created!$(RESET)"

installer-linux: linux check-linux-installer-deps
	@echo "$(CYAN)📦 Creating Linux Installers...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux $(RELEASE_DIR)
	
	@cp $(BUILD_DIR)/linux/$(APP_NAME) $(BUILD_DIR)/installer/linux/
	@cp -r web migrations $(BUILD_DIR)/installer/linux/ 2>/dev/null || true
	@cp LICENSE $(BUILD_DIR)/installer/linux/ 2>/dev/null || true
	
	@# Create start script
	@printf '#!/bin/bash\n' > $(BUILD_DIR)/installer/linux/start.sh
	@printf 'echo "========================================"\n' >> $(BUILD_DIR)/installer/linux/start.sh
	@printf 'echo "  SIMDOKPOL - System Startup"\n' >> $(BUILD_DIR)/installer/linux/start.sh
	@printf 'echo "========================================"\n' >> $(BUILD_DIR)/installer/linux/start.sh
	@printf 'echo ""\n' >> $(BUILD_DIR)/installer/linux/start.sh
	@printf 'echo "[INFO] Memulai SIMDOKPOL Server..."\n' >> $(BUILD_DIR)/installer/linux/start.sh
	@printf './simdokpol\n' >> $(BUILD_DIR)/installer/linux/start.sh
	@chmod +x $(BUILD_DIR)/installer/linux/start.sh
	
	@# Create DEB package
	@echo "$(YELLOW)Building DEB package...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/DEBIAN
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/bin
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/share/applications
	
	@cp -r $(BUILD_DIR)/installer/linux/* $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)/ 2>/dev/null || true
	
	@# Create symlink script instead of direct symlink
	@printf '#!/bin/sh\n/opt/%s/%s "$$@"\n' "$(APP_NAME)" "$(APP_NAME)" > $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	@chmod +x $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	
	@# Create desktop entry
	@printf '[Desktop Entry]\n' > $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Version=1.0\n' >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Type=Application\n' >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Name=SIMDOKPOL\n' >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Comment=Sistem Informasi Manajemen Dokumen Kepolisian\n' >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Exec=/opt/%s/start.sh\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Terminal=false\n' >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	@printf 'Categories=Office;Database;\n' >> $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop
	
	@# Create control file
	@printf 'Package: %s\n' "$(APP_NAME)" > $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf 'Version: %s\n' "$(VERSION_RAW)" >> $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf 'Section: utils\n' >> $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf 'Priority: optional\n' >> $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf 'Architecture: amd64\n' >> $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf 'Maintainer: SIMDOKPOL Team\n' >> $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	@printf 'Description: Sistem Informasi Manajemen Dokumen Kepolisian\n' >> $(BUILD_DIR)/installer/linux/deb/DEBIAN/control
	
	@# Create postinst script
	@printf '#!/bin/bash\nset -e\nchmod +x "/opt/%s/%s"\nchmod +x "/opt/%s/start.sh"\nexit 0\n' "$(APP_NAME)" "$(APP_NAME)" "$(APP_NAME)" > $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	@chmod 755 $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	
	@# Build DEB
	@dpkg-deb --build $(BUILD_DIR)/installer/linux/deb $(RELEASE_DIR)/$(APP_NAME)_$(VERSION_RAW)_amd64.deb 2>/dev/null || echo "$(RED)Failed to create DEB$(RESET)"
	
	@# Create portable package
	@cd $(BUILD_DIR)/installer/linux && tar -czf $(APP_NAME)-linux-amd64-v$(VERSION_FULL)-portable.tar.gz * 2>/dev/null
	@mv $(BUILD_DIR)/installer/linux/$(APP_NAME)-linux-amd64-v$(VERSION_FULL)-portable.tar.gz $(RELEASE_DIR)/ 2>/dev/null || true
	
	@echo "$(GREEN)✅ Linux Installers created!$(RESET)"

installer-macos: macos check-macos-deps
	@echo "$(CYAN)📦 Creating macOS Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos $(RELEASE_DIR)
	
	@# Create app bundle structure
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/backups
	
	@cp $(BUILD_DIR)/macos/$(APP_NAME) $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@cp -r web migrations $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/ 2>/dev/null || true
	
	@# Create Info.plist
	@printf '<?xml version="1.0" encoding="UTF-8"?>\n' > $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '<plist version="1.0">\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '<dict>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundleExecutable</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>%s</string>\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundleIdentifier</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>com.simdokpol.app</string>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundleName</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>%s</string>\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundlePackageType</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>APPL</string>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundleShortVersionString</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>%s</string>\n' "$(VERSION_RAW)" >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundleVersion</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>%s</string>\n' "$(VERSION_RAW)" >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>LSMinimumSystemVersion</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>10.13</string>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>NSHighResolutionCapable</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<true/>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<key>CFBundleDisplayName</key>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '	<string>%s</string>\n' "$(APP_NAME)" >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '</dict>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	@printf '</plist>\n' >> $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist
	
	@# Create DMG
	@echo "$(YELLOW)Creating DMG installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos/dmg-contents
	@cp -r $(BUILD_DIR)/installer/macos/$(APP_NAME).app $(BUILD_DIR)/installer/macos/dmg-contents/
	@ln -sf /Applications $(BUILD_DIR)/installer/macos/dmg-contents/Applications 2>/dev/null || true
	
	@# Create DMG using hdiutil (macOS) or genisoimage (Linux)
	@if [ "$(DETECTED_OS)" = "Darwin" ]; then \
		hdiutil create -volname "SIMDOKPOL" -srcfolder $(BUILD_DIR)/installer/macos/dmg-contents -ov -format UDZO $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-installer.dmg; \
	else \
		echo "$(YELLOW)Using genisoimage for DMG creation on Linux...$(RESET)"; \
		genisoimage -V "SIMDOKPOL" -D -R -apple -no-pad -o $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-installer.dmg $(BUILD_DIR)/installer/macos/dmg-contents 2>/dev/null || echo "$(RED)Failed to create DMG$(RESET)"; \
	fi
	
	@# Create portable package
	@cd $(BUILD_DIR)/installer/macos && zip -r $(APP_NAME)-macos-amd64-v$(VERSION_FULL)-portable.zip $(APP_NAME).app 2>/dev/null
	@mv $(BUILD_DIR)/installer/macos/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/ 2>/dev/null || true
	
	@echo "$(GREEN)✅ macOS Installer created!$(RESET)"

# --- BUILD TARGETS ---
windows: check-deps icon-gen
	@echo "$(CYAN)🔨 Building Windows AMD64 App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows
	@CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE)
	@rm -f $(RESOURCE_SYSO)
	@echo "$(GREEN)✅ Windows AMD64 OK.$(RESET)"

windows-arm64: check-deps icon-gen
	@echo "$(CYAN)🔨 Building Windows ARM64 App (Snapdragon)...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows-arm64
	@CGO_ENABLED=1 CC=aarch64-w64-mingw32-gcc GOOS=windows GOARCH=arm64 \
	go build -ldflags "$(LDFLAGS_APP) -H=windowsgui" -tags sqlite_omit_load_extension \
	-o $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(MAIN_FILE)
	@rm -f $(RESOURCE_SYSO)
	@echo "$(GREEN)✅ Windows ARM64 OK.$(RESET)"

linux: check-deps
	@echo "$(CYAN)🔨 Building Linux AMD64 App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE)
	@chmod +x $(BUILD_DIR)/linux/$(APP_NAME)
	@echo "$(GREEN)✅ Linux AMD64 OK.$(RESET)"

macos: check-deps
	@echo "$(CYAN)🔨 Building macOS App...$(RESET)"
	@mkdir -p $(BUILD_DIR)/macos
	@rm -f $(RESOURCE_SYSO)
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS_APP)" \
	-o $(BUILD_DIR)/macos/$(APP_NAME) $(MAIN_FILE)
	@chmod +x $(BUILD_DIR)/macos/$(APP_NAME)
	@echo "$(GREEN)✅ macOS OK.$(RESET)"

windows-installer: windows installer-windows
	@echo "$(GREEN)✅ Windows + Installer Complete!$(RESET)"

windows-arm64-installer: windows-arm64 installer-windows-arm64
	@echo "$(GREEN)✅ Windows ARM64 + Installer Complete!$(RESET)"

linux-installer: linux installer-linux
	@echo "$(GREEN)✅ Linux + Installer Complete!$(RESET)"

macos-installer: macos installer-macos
	@echo "$(GREEN)✅ macOS + Installer Complete!$(RESET)"

# --- DEPENDENCIES MANAGEMENT ---
install-deps:
	@echo "$(CYAN)📦 Installing Dependencies for $(DETECTED_OS) ($(PKG_MANAGER))...$(RESET)"
	
	@if [ "$(PKG_MANAGER)" = "pacman" ]; then \
		echo "$(YELLOW)Installing for Arch/Manjaro...$(RESET)"; \
		sudo pacman -S --needed base-devel mingw-w64-gcc go git zip unzip gtk3 webkit2gtk nsis dpkg rpm-tools genisoimage || exit 1; \
		echo "$(GREEN)✅ Arch/Manjaro dependencies installed!$(RESET)"; \
	elif [ "$(PKG_MANAGER)" = "apt" ]; then \
		echo "$(YELLOW)Installing for Ubuntu/Debian...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y build-essential gcc-mingw-w64-x86-64 gcc-mingw-w64-aarch64 golang-go git zip unzip libgtk-3-dev libwebkit2gtk-4.0-dev nsis dpkg rpm genisoimage || exit 1; \
		echo "$(GREEN)✅ Ubuntu/Debian dependencies installed!$(RESET)"; \
	else \
		echo "$(RED)❌ Unsupported package manager: $(PKG_MANAGER)$(RESET)"; \
		echo "Please install manually: build-essential, mingw-w64, go, git, zip, nsis, dpkg, rpm, genisoimage"; \
		exit 1; \
	fi
	
	@echo "$(YELLOW)📦 Installing Go tools...$(RESET)"
	@go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest || echo "$(YELLOW)Warning: Could not install goversioninfo$(RESET)"
	@echo "$(GREEN)✅ All dependencies installed!$(RESET)"

check-nsis-deps:
	@if ! command -v makensis >/dev/null 2>&1; then \
		echo "$(RED)❌ NSIS not found. Installing...$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			sudo pacman -S --needed nsis || exit 1; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			sudo apt-get install -y nsis || exit 1; \
		fi; \
	fi

check-linux-installer-deps:
	@if ! command -v dpkg-deb >/dev/null 2>&1; then \
		echo "$(RED)❌ dpkg-deb not found. Installing...$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			sudo pacman -S --needed dpkg || exit 1; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			sudo apt-get install -y dpkg || exit 1; \
		fi; \
	fi

check-macos-deps:
	@if [ "$(DETECTED_OS)" != "Darwin" ] && ! command -v genisoimage >/dev/null 2>&1; then \
		echo "$(RED)❌ genisoimage not found. Installing...$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			sudo pacman -S --needed cdrkit || exit 1; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			sudo apt-get install -y genisoimage || exit 1; \
		fi; \
	fi

# --- UTILITY TARGETS ---
changelog:
	@echo "$(YELLOW)📝 Generating Changelog...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@printf "CHANGELOG - %s %s\n\n" "$(APP_NAME)" "$(VERSION_FULL)" > $(BUILD_DIR)/CHANGELOG.txt
	@git log $(PREV_VERSION)..HEAD --pretty=format:"- %s" >> $(BUILD_DIR)/CHANGELOG.txt 2>/dev/null || printf "- Initial release\n" >> $(BUILD_DIR)/CHANGELOG.txt
	@printf "\n" >> $(BUILD_DIR)/CHANGELOG.txt
	@echo "$(GREEN)✅ Changelog generated at $(BUILD_DIR)/CHANGELOG.txt$(RESET)"

icon-gen:
	@echo "$(YELLOW)🖼️  Generating Windows Metadata...$(RESET)"
	@if ! command -v goversioninfo >/dev/null 2>&1; then \
		echo "   Installing goversioninfo..."; \
		go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest || (echo "$(RED)Failed to install goversioninfo$(RESET)" && exit 1); \
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
	
	@goversioninfo -o $(RESOURCE_SYSO) || (echo "$(RED)Failed to generate resource$(RESET)" && exit 1)
	@rm -f versioninfo.json
	@echo "$(GREEN)✅ Windows metadata generated$(RESET)"

deps:
	@echo "$(YELLOW)🔍 Checking Dependencies...$(RESET)"
	@go mod tidy || (echo "$(RED)Failed to tidy Go modules$(RESET)" && exit 1)
	@echo "$(GREEN)✅ Dependencies OK$(RESET)"

clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR) $(RESOURCE_SYSO) versioninfo.json
	@echo "$(GREEN)✅ Clean complete!$(RESET)"

list-targets:
	@echo "$(CYAN)🎯 Available Build Targets:$(RESET)"
	@echo "  windows-installer      - Build + Windows Installer"
	@echo "  windows-arm64-installer - Build + Windows ARM64 Installer"
	@echo "  linux-installer        - Build + Linux DEB + Portable"
	@echo "  macos-installer        - Build + macOS DMG + Portable"
	@echo "  installer-all          - Build all installers"
	@echo "  release                - Full release with installers"
	@echo "  install-deps           - Install all system dependencies"
	@echo "  clean                  - Remove all build artifacts"
	@echo "  deps                   - Update Go dependencies"

# --- ADMIN TOOLS ---
tools: check-deps
	@echo "$(CYAN)🛠️  Building Admin Tools...$(RESET)"
	@mkdir -p $(BUILD_DIR)/tools
	@echo "$(YELLOW)Admin tools target not yet implemented$(RESET)"
	@echo "$(GREEN)✅ Tools built$(RESET)"

# --- QUICK TARGETS ---
all: clean deps installer-all
	@echo "$(GREEN)🎉 All builds completed!$(RESET)"

quick-windows: windows installer-windows
	@echo "$(GREEN)✅ Quick Windows build complete!$(RESET)"

quick-linux: linux installer-linux
	@echo "$(GREEN)✅ Quick Linux build complete!$(RESET)"

# --- HELP ---
help:
	@echo "$(CYAN)════════════════════════════════════════$(RESET)"
	@echo "$(CYAN)  SIMDOKPOL Build System Help$(RESET)"
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
	@echo "  make linux-installer         - Linux DEB + Portable"
	@echo "  make macos-installer         - macOS DMG + Portable"
	@echo ""
	@echo "$(YELLOW)Binary Only (No Installer):$(RESET)"
	@echo "  make windows           - Windows AMD64 binary"
	@echo "  make windows-arm64     - Windows ARM64 binary"
	@echo "  make linux             - Linux binary"
	@echo "  make macos             - macOS binary"
	@echo ""
	@echo "$(YELLOW)Utilities:$(RESET)"
	@echo "  make install-deps      - Install system dependencies"
	@echo "  make deps              - Update Go dependencies"
	@echo "  make changelog         - Generate changelog"
	@echo "  make list-targets      - List all available targets"
	@echo "  make help              - Show this help"
	@echo ""
	@echo "$(YELLOW)Current Configuration:$(RESET)"
	@echo "  Version: $(GREEN)$(VERSION_FULL)$(RESET)"
	@echo "  OS: $(GREEN)$(DETECTED_OS)$(RESET)"
	@echo "  Arch: $(GREEN)$(DETECTED_ARCH)$(RESET)"
	@echo "$(CYAN)════════════════════════════════════════$(RESET)"