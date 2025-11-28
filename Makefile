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
DETECTED_OS := $(shell uname -s)
DETECTED_ARCH := $(shell uname -m)
PKG_MANAGER := $(shell if command -v pacman >/dev/null 2>&1; then echo "pacman"; elif command -v apt-get >/dev/null 2>&1; then echo "apt"; else echo "unknown"; fi)

# Warna
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all package menu windows windows-arm64 linux macos tools changelog clean icon-gen deps check-deps install-deps installer-windows installer-linux installer-macos installer-all

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

# --- RELEASE PIPELINE ---
release: clean deps changelog installer-all
	@echo "$(GREEN)✅ RELEASE SELESAI! Semua installer tersedia di '$(RELEASE_DIR)/'$(RESET)"
	@echo "$(YELLOW)📁 File yang dihasilkan:$(RESET)"
	@ls -la $(RELEASE_DIR)/*

# --- INSTALLER TARGETS ---
installer-all: installer-windows installer-windows-arm64 installer-linux installer-macos
	@echo "$(GREEN)✅ Semua installer berhasil dibuat!$(RESET)"

installer-windows: windows check-nsis-deps
	@echo "$(CYAN)📦 Creating Windows Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows
	
	@# Prepare files for installer
	@cp $(BUILD_DIR)/windows/$(APP_NAME).exe $(BUILD_DIR)/installer/windows/
	@cp -r web migrations $(BUILD_DIR)/installer/windows/
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows/icon.ico 2>/dev/null || echo "Warning: Icon not found"
	@cp LICENSE $(BUILD_DIR)/installer/windows/ 2>/dev/null || echo "Warning: LICENSE not found"
	
	@# Create batch files
	@cat > $(BUILD_DIR)/installer/windows/start.bat << 'EOF'
	@echo off
	title SIMDOKPOL Server
	echo ========================================
	echo   SIMDOKPOL - System Startup
	echo ========================================
	echo.
	echo [INFO] Memulai SIMDOKPOL Server...
	echo [INFO] Aplikasi akan terbuka otomatis di browser
	echo [INFO] Icon aplikasi tersedia di system tray
	echo.
	
	start simdokpol.exe
	
	echo [SUCCESS] Server berjalan di background
	echo [INFO] Klik icon di system tray untuk kontrol aplikasi
	echo.
	EOF
	
	@# Create NSIS installer script
	@cat > $(BUILD_DIR)/installer/windows/installer.nsi << 'NSISEOF'
	!define APP_NAME "SIMDOKPOL"
	!define VERSION "$(VERSION_RAW)"
	!define PUBLISHER "SIMDOKPOL Team"
	!define WEB_SITE "https://github.com/muhammad1505/simdokpol"
	!define INSTALLER_NAME "$(APP_NAME)-windows-x64-v$(VERSION_FULL)-installer.exe"
	
	Name "$${APP_NAME} $${VERSION}"
	OutFile "$${INSTALLER_NAME}"
	InstallDir "$$PROGRAMFILES64\$${APP_NAME}"
	RequestExecutionLevel admin
	
	!include "MUI2.nsh"
	
	!define MUI_ICON "icon.ico"
	!define MUI_UNICON "icon.ico"
	
	!insertmacro MUI_PAGE_WELCOME
	NSISEOF
	
	@# Add license page if exists
	@if [ -f "LICENSE" ]; then \
		cat >> $(BUILD_DIR)/installer/windows/installer.nsi << 'NSISEOF'
	!insertmacro MUI_PAGE_LICENSE "LICENSE"
	NSISEOF
	fi
	
	@cat >> $(BUILD_DIR)/installer/windows/installer.nsi << 'NSISEOF'
	!insertmacro MUI_PAGE_DIRECTORY
	!insertmacro MUI_PAGE_INSTFILES
	
	!define MUI_FINISHPAGE_RUN "$$INSTDIR\$(APP_NAME).exe"
	!define MUI_FINISHPAGE_RUN_TEXT "Jalankan $${APP_NAME}"
	!insertmacro MUI_PAGE_FINISH
	
	!insertmacro MUI_UNPAGE_CONFIRM
	!insertmacro MUI_UNPAGE_INSTFILES
	!insertmacro MUI_UNPAGE_FINISH
	
	!insertmacro MUI_LANGUAGE "English"
	
	Section "Install"
	  SetOutPath "$$INSTDIR"
	  File "$(APP_NAME).exe"
	  File "icon.ico"
	  File /r "web"
	  File /r "migrations"
	  
	  CreateDirectory "$$INSTDIR\backups"
	  
	  WriteUninstaller "$$INSTDIR\Uninstall.exe"
	  
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "DisplayName" "$${APP_NAME}"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "UninstallString" "$$INSTDIR\Uninstall.exe"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "DisplayIcon" "$$INSTDIR\icon.ico"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "Publisher" "$${PUBLISHER}"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "DisplayVersion" "$${VERSION}"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "URLInfoAbout" "$${WEB_SITE}"
	  
	  CreateShortcut "$$DESKTOP\$${APP_NAME}.lnk" "$$INSTDIR\$(APP_NAME).exe" "" "$$INSTDIR\icon.ico" 0
	  
	  CreateDirectory "$$SMPROGRAMS\$${APP_NAME}"
	  CreateShortcut "$$SMPROGRAMS\$${APP_NAME}\$${APP_NAME}.lnk" "$$INSTDIR\$(APP_NAME).exe" "" "$$INSTDIR\icon.ico" 0
	  CreateShortcut "$$SMPROGRAMS\$${APP_NAME}\Uninstall.lnk" "$$INSTDIR\Uninstall.exe" "" "$$INSTDIR\icon.ico" 0
	SectionEnd
	
	Section "Uninstall"
	  Delete "$$DESKTOP\$${APP_NAME}.lnk"
	  RMDir /r "$$SMPROGRAMS\$${APP_NAME}"
	  
	  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}"
	  
	  RMDir /r "$$INSTDIR"
	SectionEnd
	NSISEOF
	
	@# Build NSIS installer
	@cd $(BUILD_DIR)/installer/windows && makensis installer.nsi
	@mkdir -p $(RELEASE_DIR)
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-v$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/
	
	@# Create portable package
	@cd $(BUILD_DIR)/installer/windows && zip -r $(APP_NAME)-windows-x64-v$(VERSION_FULL)-portable.zip . -x "installer.nsi"
	@mv $(BUILD_DIR)/installer/windows/$(APP_NAME)-windows-x64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/
	
	@echo "$(GREEN)✅ Windows Installer created!$(RESET)"

installer-windows-arm64: windows-arm64 check-nsis-deps
	@echo "$(CYAN)📦 Creating Windows ARM64 Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/windows-arm64
	
	@# Similar process as Windows AMD64 but for ARM64
	@cp $(BUILD_DIR)/windows-arm64/$(APP_NAME).exe $(BUILD_DIR)/installer/windows-arm64/
	@cp -r web migrations $(BUILD_DIR)/installer/windows-arm64/
	@cp $(ICON_PATH) $(BUILD_DIR)/installer/windows-arm64/icon.ico 2>/dev/null || echo "Warning: Icon not found"
	
	@# Create NSIS installer for ARM64
	@cat > $(BUILD_DIR)/installer/windows-arm64/installer.nsi << 'NSISEOF'
	!define APP_NAME "SIMDOKPOL"
	!define VERSION "$(VERSION_RAW)"
	!define ARCH "ARM64"
	!define INSTALLER_NAME "$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-installer.exe"
	
	Name "$${APP_NAME} $${VERSION} ($${ARCH})"
	OutFile "$${INSTALLER_NAME}"
	InstallDir "$$PROGRAMFILES64\$${APP_NAME}"
	RequestExecutionLevel admin
	
	!include "MUI2.nsh"
	
	!define MUI_ICON "icon.ico"
	!define MUI_UNICON "icon.ico"
	
	!insertmacro MUI_PAGE_WELCOME
	NSISEOF
	
	@if [ -f "LICENSE" ]; then \
		cat >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi << 'NSISEOF'
	!insertmacro MUI_PAGE_LICENSE "LICENSE"
	NSISEOF
	fi
	
	@cat >> $(BUILD_DIR)/installer/windows-arm64/installer.nsi << 'NSISEOF'
	!insertmacro MUI_PAGE_DIRECTORY
	!insertmacro MUI_PAGE_INSTFILES
	!insertmacro MUI_PAGE_FINISH
	
	!insertmacro MUI_UNPAGE_CONFIRM
	!insertmacro MUI_UNPAGE_INSTFILES
	
	!insertmacro MUI_LANGUAGE "English"
	
	Section "Install"
	  SetOutPath "$$INSTDIR"
	  File "$(APP_NAME).exe"
	  File "icon.ico"
	  File /r "web"
	  File /r "migrations"
	  
	  CreateDirectory "$$INSTDIR\backups"
	  
	  WriteUninstaller "$$INSTDIR\Uninstall.exe"
	  
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "DisplayName" "$${APP_NAME} ($${ARCH})"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "UninstallString" "$$INSTDIR\Uninstall.exe"
	  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}" "DisplayVersion" "$${VERSION}"
	  
	  CreateShortcut "$$DESKTOP\$${APP_NAME}.lnk" "$$INSTDIR\$(APP_NAME).exe" "" "$$INSTDIR\icon.ico" 0
	  CreateDirectory "$$SMPROGRAMS\$${APP_NAME}"
	  CreateShortcut "$$SMPROGRAMS\$${APP_NAME}\$${APP_NAME}.lnk" "$$INSTDIR\$(APP_NAME).exe" "" "$$INSTDIR\icon.ico" 0
	SectionEnd
	
	Section "Uninstall"
	  Delete "$$DESKTOP\$${APP_NAME}.lnk"
	  RMDir /r "$$SMPROGRAMS\$${APP_NAME}"
	  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\$${APP_NAME}"
	  RMDir /r "$$INSTDIR"
	SectionEnd
	NSISEOF
	
	@cd $(BUILD_DIR)/installer/windows-arm64 && makensis installer.nsi
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-installer.exe $(RELEASE_DIR)/
	
	@# Create portable package for ARM64
	@cd $(BUILD_DIR)/installer/windows-arm64 && zip -r $(APP_NAME)-windows-arm64-v$(VERSION_FULL)-portable.zip . -x "installer.nsi"
	@mv $(BUILD_DIR)/installer/windows-arm64/$(APP_NAME)-windows-arm64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/
	
	@echo "$(GREEN)✅ Windows ARM64 Installer created!$(RESET)"

installer-linux: linux check-linux-installer-deps
	@echo "$(CYAN)📦 Creating Linux Installers...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux
	
	@# Prepare files
	@cp $(BUILD_DIR)/linux/$(APP_NAME) $(BUILD_DIR)/installer/linux/
	@cp -r web migrations $(BUILD_DIR)/installer/linux/
	@cp LICENSE $(BUILD_DIR)/installer/linux/ 2>/dev/null || echo "Warning: LICENSE not found"
	
	@# Create start script
	@cat > $(BUILD_DIR)/installer/linux/start.sh << 'EOF'
	#!/bin/bash
	echo "========================================"
	echo "  SIMDOKPOL - System Startup"
	echo "========================================"
	echo ""
	echo "[INFO] Memulai SIMDOKPOL Server..."
	echo "[INFO] Aplikasi akan terbuka otomatis di browser"
	echo "[INFO] Icon aplikasi tersedia di system tray"
	echo "[INFO] Tekan Ctrl+C untuk menghentikan server"
	echo ""
	./simdokpol
	EOF
	@chmod +x $(BUILD_DIR)/installer/linux/start.sh
	
	@# Create DEB package
	@echo "$(YELLOW)Building DEB package...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/DEBIAN
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/bin
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/share/applications
	@mkdir -p $(BUILD_DIR)/installer/linux/deb/usr/share/icons/hicolor/256x256/apps
	
	@cp -r $(BUILD_DIR)/installer/linux/* $(BUILD_DIR)/installer/linux/deb/opt/$(APP_NAME)/
	@cp web/static/img/icon.png $(BUILD_DIR)/installer/linux/deb/usr/share/icons/hicolor/256x256/apps/$(APP_NAME).png 2>/dev/null || echo "Warning: Icon not found"
	
	@# Create symlink
	@ln -sf /opt/$(APP_NAME)/$(APP_NAME) $(BUILD_DIR)/installer/linux/deb/usr/bin/$(APP_NAME)
	
	@# Create desktop entry
	@cat > $(BUILD_DIR)/installer/linux/deb/usr/share/applications/$(APP_NAME).desktop << 'EOF'
	[Desktop Entry]
	Version=1.0
	Type=Application
	Name=SIMDOKPOL
	Comment=Sistem Informasi Manajemen Dokumen Kepolisian
	Exec=/opt/$(APP_NAME)/start.sh
	Icon=$(APP_NAME)
	Terminal=false
	Categories=Office;Database;
	StartupWMClass=SIMDOKPOL
	EOF
	
	@# Create control file
	@cat > $(BUILD_DIR)/installer/linux/deb/DEBIAN/control << 'EOF'
	Package: $(APP_NAME)
	Version: $(VERSION_RAW)
	Section: utils
	Priority: optional
	Architecture: amd64
	Depends: libgtk-3-0, libayatana-appindicator3-1
	Maintainer: SIMDOKPOL Team
	Description: Sistem Informasi Manajemen Dokumen Kepolisian
	 SIMDOKPOL adalah aplikasi desktop untuk manajemen dokumen
	 kepolisian berbasis web yang dapat diakses melalui browser.
	EOF
	
	@# Create postinst script
	@cat > $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst << 'EOF'
	#!/bin/bash
	set -e
	chmod +x "/opt/$(APP_NAME)/$(APP_NAME)"
	chmod +x "/opt/$(APP_NAME)/start.sh"
	echo "SIMDOKPOL berhasil diinstall!"
	echo "Jalankan dengan perintah: $(APP_NAME)"
	exit 0
	EOF
	@chmod 755 $(BUILD_DIR)/installer/linux/deb/DEBIAN/postinst
	
	@# Build DEB
	@dpkg-deb --build $(BUILD_DIR)/installer/linux/deb $(RELEASE_DIR)/$(APP_NAME)_$(VERSION_RAW)_amd64.deb
	
	@# Create portable package
	@cd $(BUILD_DIR)/installer/linux && tar -czf $(APP_NAME)-linux-amd64-v$(VERSION_FULL)-portable.tar.gz *
	@mv $(BUILD_DIR)/installer/linux/$(APP_NAME)-linux-amd64-v$(VERSION_FULL)-portable.tar.gz $(RELEASE_DIR)/
	
	@echo "$(GREEN)✅ Linux Installers created!$(RESET)"

installer-macos: macos check-macos-deps
	@echo "$(CYAN)📦 Creating macOS Installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos
	
	@# Create app bundle structure
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources
	@mkdir -p $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/backups
	
	@cp $(BUILD_DIR)/macos/$(APP_NAME) $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@cp -r web migrations $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Resources/
	
	@# Create Info.plist
	@cat > $(BUILD_DIR)/installer/macos/$(APP_NAME).app/Contents/Info.plist << 'EOF'
	<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
	<dict>
		<key>CFBundleExecutable</key>
		<string>$(APP_NAME)</string>
		<key>CFBundleIdentifier</key>
		<string>com.simdokpol.app</string>
		<key>CFBundleName</key>
		<string>$(APP_NAME)</string>
		<key>CFBundlePackageType</key>
		<string>APPL</string>
		<key>CFBundleShortVersionString</key>
		<string>$(VERSION_RAW)</string>
		<key>CFBundleVersion</key>
		<string>$(VERSION_RAW)</string>
		<key>LSMinimumSystemVersion</key>
		<string>10.13</string>
		<key>NSHighResolutionCapable</key>
		<true/>
		<key>CFBundleDisplayName</key>
		<string>$(APP_NAME)</string>
	</dict>
	</plist>
	EOF
	
	@# Create DMG
	@echo "$(YELLOW)Creating DMG installer...$(RESET)"
	@mkdir -p $(BUILD_DIR)/installer/macos/dmg-contents
	@cp -r $(BUILD_DIR)/installer/macos/$(APP_NAME).app $(BUILD_DIR)/installer/macos/dmg-contents/
	@ln -s /Applications $(BUILD_DIR)/installer/macos/dmg-contents/Applications
	
	@# Create DMG using hdiutil (macOS) or genisoimage (Linux)
	@if [ "$(DETECTED_OS)" = "Darwin" ]; then \
		hdiutil create -volname "SIMDOKPOL" -srcfolder $(BUILD_DIR)/installer/macos/dmg-contents -ov -format UDZO $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-installer.dmg; \
	else \
		echo "$(YELLOW)Using genisoimage for DMG creation on Linux...$(RESET)"; \
		genisoimage -V "SIMDOKPOL" -D -R -apple -no-pad -o $(RELEASE_DIR)/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-installer.dmg $(BUILD_DIR)/installer/macos/dmg-contents; \
	fi
	
	@# Create portable package
	@cd $(BUILD_DIR)/installer/macos && zip -r $(APP_NAME)-macos-amd64-v$(VERSION_FULL)-portable.zip $(APP_NAME).app
	@mv $(BUILD_DIR)/installer/macos/$(APP_NAME)-macos-amd64-v$(VERSION_FULL)-portable.zip $(RELEASE_DIR)/
	
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
		sudo pacman -S --needed base-devel mingw-w64-gcc go git zip unzip gtk3 webkit2gtk nsis dpkg rpm-tools genisoimage; \
		echo "$(GREEN)✅ Arch/Manjaro dependencies installed!$(RESET)"; \
	elif [ "$(PKG_MANAGER)" = "apt" ]; then \
		echo "$(YELLOW)Installing for Ubuntu/Debian...$(RESET)"; \
		sudo apt-get update && sudo apt-get install -y build-essential gcc-mingw-w64-x86-64 gcc-mingw-w64-arm64 golang-go git zip unzip libgtk-3-dev libwebkit2gtk-4.0-dev nsis dpkg rpm genisoimage; \
		echo "$(GREEN)✅ Ubuntu/Debian dependencies installed!$(RESET)"; \
	else \
		echo "$(RED)❌ Unsupported package manager: $(PKG_MANAGER)$(RESET)"; \
		echo "Please install manually: build-essential, mingw-w64, go, git, zip, nsis, dpkg, rpm, genisoimage"; \
		exit 1; \
	fi
	
	@echo "$(YELLOW)📦 Installing Go tools...$(RESET)"
	@go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	@echo "$(GREEN)✅ All dependencies installed!$(RESET)"

check-nsis-deps:
	@if ! command -v makensis >/dev/null 2>&1; then \
		echo "$(RED)❌ NSIS not found. Installing...$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			sudo pacman -S --needed nsis; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			sudo apt-get install -y nsis; \
		fi; \
	fi

check-linux-installer-deps:
	@if ! command -v dpkg-deb >/dev/null 2>&1; then \
		echo "$(RED)❌ dpkg-deb not found. Installing...$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			sudo pacman -S --needed dpkg; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			sudo apt-get install -y dpkg; \
		fi; \
	fi

check-macos-deps:
	@if [ "$(DETECTED_OS)" != "Darwin" ] && ! command -v genisoimage >/dev/null 2>&1; then \
		echo "$(RED)❌ genisoimage not found. Installing...$(RESET)"; \
		if [ "$(PKG_MANAGER)" = "pacman" ]; then \
			sudo pacman -S --needed cdrkit; \
		elif [ "$(PKG_MANAGER)" = "apt" ]; then \
			sudo apt-get install -y genisoimage; \
		fi; \
	fi

# --- UTILITY TARGETS ---
changelog:
	@echo "$(YELLOW)📝 Generating Changelog...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@echo "CHANGELOG - $(APP_NAME) $(VERSION_FULL)" > $(BUILD_DIR)/CHANGELOG.txt
	@echo "" >> $(BUILD_DIR)/CHANGELOG.txt
	@git log $(PREV_VERSION)..HEAD --pretty=format:"- %s" >> $(BUILD_DIR)/CHANGELOG.txt 2>/dev/null || echo "- Initial release" >> $(BUILD_DIR)/CHANGELOG.txt
	@echo "" >> $(BUILD_DIR)/CHANGELOG.txt

icon-gen:
	@echo "$(YELLOW)🖼️  Generating Windows Metadata...$(RESET)"
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
	@echo '        }' >> versioninfo.json
	@echo '    }' >> versioninfo.json
	@echo '}' >> versioninfo.json
	
	@goversioninfo -o $(RESOURCE_SYSO)
	@rm -f versioninfo.json

deps:
	@echo "$(YELLOW)🔍 Checking Dependencies...$(RESET)"
	@go mod tidy

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

# --- QUICK TARGETS ---
all: clean deps installer-all
	@echo "$(GREEN)🎉 All builds completed!$(RESET)"

quick-windows: windows installer-windows
	@echo "$(GREEN)✅ Quick Windows build complete!$(RESET)"

quick-linux: linux installer-linux
	@echo "$(GREEN)✅ Quick Linux build complete!$(RESET)"