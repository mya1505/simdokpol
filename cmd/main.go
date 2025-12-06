package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"simdokpol/internal/config"
	"simdokpol/internal/controllers"
	"simdokpol/internal/dto"
	"simdokpol/internal/middleware"
	"simdokpol/internal/models"
	"simdokpol/internal/repositories"
	"simdokpol/internal/services"
	"simdokpol/internal/utils"
	"simdokpol/web"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var (
	version         = "dev"
	changelogBase64 = ""
)

func main() {
  // --- 1. INISIALISASI SECRETS (PENTING!) ---
  // Jika dijalankan via 'go run', variabel ini kosong. Kita isi default DEV.
  // Jika build production, variabel ini akan diisi via -ldflags, jadi logic ini di-skip.
  if services.AppSecretKeyString == "" {
      // Cek ENV dulu
      if envKey := os.Getenv("APP_SECRET_KEY"); envKey != "" {
          services.AppSecretKeyString = envKey
      } else {
          services.AppSecretKeyString = "DEV_SECRET_KEY_JANGAN_DIPAKAI_PROD"
      }
  }

  if len(services.JWTSecretKey) == 0 {
      if envJwt := os.Getenv("JWT_SECRET_KEY"); envJwt != "" {
          services.JWTSecretKey = []byte(envJwt)
      } else {
          services.JWTSecretKey = []byte("DEV_JWT_SECRET_KEY")
      }
  }
	// Initialize environment and logging
	setupEnvironment()
	setupLogging()

	appData := utils.GetAppDataDir()
	log.Println("==========================================")
	log.Printf("🚀 SIMDOKPOL DESKTOP - v%s", version)
	log.Printf("💻 Mode: PC / GUI ENABLED")
	log.Printf("📂 Data Dir: %s", appData)
	log.Println("==========================================")

	// Load configuration
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("❌ GAGAL MEMUAT KONFIGURASI: Configuration tidak dapat dimuat")
		return
	}

	// Setup database with proper error handling
	log.Println("🔧 Menginisialisasi database...")
	db, err := setupDatabase(cfg)
	
	if err != nil {
		msg := fmt.Sprintf("GAGAL KONEKSI DATABASE: %v", err)
		log.Printf("❌ %s", msg)
		
		// Show alert to user
		_ = beeep.Alert("SIMDOKPOL Error", msg, "")
		
		// Log detailed error information
		log.Printf("❌ Database Error Details:")
		log.Printf("   - Dialect: %s", cfg.DBDialect)
		log.Printf("   - DSN: %s", cfg.DBDSN)
		log.Printf("   - Error: %v", err)
		
		// Exit application - cannot continue without database
		log.Fatal("❌ APLIKASI TIDAK DAPAT DILANJUTKAN: Database diperlukan untuk menjalankan aplikasi")
		return
	}

	// Verify database connection is not nil
	if db == nil {
		log.Fatal("❌ KONEKSI DATABASE NULL: Database connection tidak dapat dibuat")
		return
	}

	// Test database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ GAGAL MENDAPATKAN DATABASE INSTANCE: %v", err)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ GAGAL PING DATABASE: %v", err)
		return
	}

	log.Println("✅ Database berhasil terhubung dan siap digunakan")

	// Seed default templates only after successful DB connection
	seedDefaultTemplates(db)

	// Initialize repositories with nil checks
	log.Println("🔧 Menginisialisasi repositories...")
	
	userRepo := repositories.NewUserRepository(db)
	if userRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT USER REPOSITORY")
		return
	}

	docRepo := repositories.NewLostDocumentRepository(db)
	if docRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT DOCUMENT REPOSITORY")
		return
	}

	residentRepo := repositories.NewResidentRepository(db)
	if residentRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT RESIDENT REPOSITORY")
		return
	}

	configRepo := repositories.NewConfigRepository(db)
	if configRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT CONFIG REPOSITORY")
		return
	}

	auditRepo := repositories.NewAuditLogRepository(db)
	if auditRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT AUDIT REPOSITORY")
		return
	}

	licenseRepo := repositories.NewLicenseRepository(db)
	if licenseRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT LICENSE REPOSITORY")
		return
	}

	itemTemplateRepo := repositories.NewItemTemplateRepository(db)
	if itemTemplateRepo == nil {
		log.Fatal("❌ GAGAL MEMBUAT ITEM TEMPLATE REPOSITORY")
		return
	}

	log.Println("✅ Repositories berhasil dibuat")

	// Initialize services with proper error handling
	log.Println("🔧 Menginisialisasi services...")

	auditService := services.NewAuditLogService(auditRepo)
	if auditService == nil {
		log.Fatal("❌ GAGAL MEMBUAT AUDIT SERVICE")
		return
	}

	configService := services.NewConfigService(configRepo, db)
	if configService == nil {
		log.Fatal("❌ GAGAL MEMBUAT CONFIG SERVICE")
		return
	}

	backupService := services.NewBackupService(db, cfg, configService, auditService)
	if backupService == nil {
		log.Fatal("❌ GAGAL MEMBUAT BACKUP SERVICE")
		return
	}

	// This is the critical line 92 where the original error occurred
	licenseService := services.NewLicenseService(licenseRepo, configService, auditService)
	if licenseService == nil {
		log.Fatal("❌ GAGAL MEMBUAT LICENSE SERVICE")
		return
	}

	userService := services.NewUserService(userRepo, auditService, cfg)
	if userService == nil {
		log.Fatal("❌ GAGAL MEMBUAT USER SERVICE")
		return
	}

	authService := services.NewAuthService(userRepo, configService)
	if authService == nil {
		log.Fatal("❌ GAGAL MEMBUAT AUTH SERVICE")
		return
	}

	migrationService := services.NewDataMigrationService(db, auditService, configService)
	if migrationService == nil {
		log.Fatal("❌ GAGAL MEMBUAT MIGRATION SERVICE")
		return
	}

	// Get executable path for services that need it
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("⚠️  Warning: Cannot get executable path: %v", err)
		exePath = ""
	}
	exeDir := filepath.Dir(exePath)

	docService := services.NewLostDocumentService(db, docRepo, residentRepo, userRepo, auditService, configService, configRepo, exeDir)
	if docService == nil {
		log.Fatal("❌ GAGAL MEMBUAT DOCUMENT SERVICE")
		return
	}

	dashboardService := services.NewDashboardService(docRepo, userRepo, configService)
	if dashboardService == nil {
		log.Fatal("❌ GAGAL MEMBUAT DASHBOARD SERVICE")
		return
	}

	reportService := services.NewReportService(docRepo, configService, exeDir)
	if reportService == nil {
		log.Fatal("❌ GAGAL MEMBUAT REPORT SERVICE")
		return
	}

	itemTemplateService := services.NewItemTemplateService(itemTemplateRepo)
	if itemTemplateService == nil {
		log.Fatal("❌ GAGAL MEMBUAT ITEM TEMPLATE SERVICE")
		return
	}

	dbTestService := services.NewDBTestService()
	if dbTestService == nil {
		log.Fatal("❌ GAGAL MEMBUAT DB TEST SERVICE")
		return
	}

	updateService := services.NewUpdateService()
	if updateService == nil {
		log.Fatal("❌ GAGAL MEMBUAT UPDATE SERVICE")
		return
	}

	log.Println("✅ Services berhasil dibuat")

	// Initialize controllers with proper error handling
	log.Println("🔧 Menginisialisasi controllers...")

	authController := controllers.NewAuthController(authService, configService)
	if authController == nil {
		log.Fatal("❌ GAGAL MEMBUAT AUTH CONTROLLER")
		return
	}

	userController := controllers.NewUserController(userService)
	if userController == nil {
		log.Fatal("❌ GAGAL MEMBUAT USER CONTROLLER")
		return
	}

	docController := controllers.NewLostDocumentController(docService)
	if docController == nil {
		log.Fatal("❌ GAGAL MEMBUAT DOCUMENT CONTROLLER")
		return
	}

	dashboardController := controllers.NewDashboardController(dashboardService)
	if dashboardController == nil {
		log.Fatal("❌ GAGAL MEMBUAT DASHBOARD CONTROLLER")
		return
	}

	configController := controllers.NewConfigController(configService, userService, backupService, migrationService)
	if configController == nil {
		log.Fatal("❌ GAGAL MEMBUAT CONFIG CONTROLLER")
		return
	}

	auditController := controllers.NewAuditLogController(auditService)
	if auditController == nil {
		log.Fatal("❌ GAGAL MEMBUAT AUDIT CONTROLLER")
		return
	}

	backupController := controllers.NewBackupController(backupService)
	if backupController == nil {
		log.Fatal("❌ GAGAL MEMBUAT BACKUP CONTROLLER")
		return
	}

	settingsController := controllers.NewSettingsController(configService, auditService)
	if settingsController == nil {
		log.Fatal("❌ GAGAL MEMBUAT SETTINGS CONTROLLER")
		return
	}

	licenseController := controllers.NewLicenseController(licenseService, auditService)
	if licenseController == nil {
		log.Fatal("❌ GAGAL MEMBUAT LICENSE CONTROLLER")
		return
	}

	reportController := controllers.NewReportController(reportService, configService)
	if reportController == nil {
		log.Fatal("❌ GAGAL MEMBUAT REPORT CONTROLLER")
		return
	}

	itemTemplateController := controllers.NewItemTemplateController(itemTemplateService)
	if itemTemplateController == nil {
		log.Fatal("❌ GAGAL MEMBUAT ITEM TEMPLATE CONTROLLER")
		return
	}

	dbTestController := controllers.NewDBTestController(dbTestService)
	if dbTestController == nil {
		log.Fatal("❌ GAGAL MEMBUAT DB TEST CONTROLLER")
		return
	}

	updateController := controllers.NewUpdateController(updateService, version)
	if updateController == nil {
		log.Fatal("❌ GAGAL MEMBUAT UPDATE CONTROLLER")
		return
	}

	log.Println("✅ Controllers berhasil dibuat")

	// Setup Gin router
	log.Println("🔧 Menginisialisasi web server...")

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Setup templates
	funcMap := template.FuncMap{"ToUpper": strings.ToUpper}
	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(web.Assets, "templates/*.html", "templates/partials/*.html"))
	r.SetHTMLTemplate(templ)
	r.StaticFS("/static", web.GetStaticFS())

	// Set global template variables
	r.Use(func(c *gin.Context) {
		c.Set("AppVersion", version)
		decodedChangelog, _ := utils.DecodeBase64(changelogBase64)
		c.Set("AppChangelog", decodedChangelog)
		c.Next()
	})

	// Setup routes
	setupRoutes(r, authController, configController, dbTestController, userRepo, configService, 
		dashboardController, updateController, docService, docController, itemTemplateController, 
		userController, backupController, auditController, licenseController, licenseService, reportController, settingsController)

	// Setup HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Setup graceful shutdown
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup HTTPS if enabled
	isHTTPS := os.Getenv("ENABLE_HTTPS") == "true"
	certFile, keyFile := "", ""
	if isHTTPS {
		var err error
		certFile, keyFile, err = utils.EnsureCertificates()
		if err != nil {
			log.Printf("⚠️  HTTPS setup failed: %v", err)
			isHTTPS = false
		}
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Starting server on port %s (HTTPS: %v)", port, isHTTPS)
		
		if isHTTPS {
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("❌ HTTPS Server Error: %s\n", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("❌ HTTP Server Error: %s\n", err)
			}
		}
	}()

	// Setup graceful shutdown handler
	go func() {
		<-quitChan
		log.Println("🛑 Shutdown signal received, stopping server...")
		systray.Quit()
	}()

	log.Println("✅ SIMDOKPOL berhasil diinisialisasi, menjalankan system tray...")

	// Run system tray
	systray.Run(func() { 
		onReady(isHTTPS, port) 
	}, func() {
		log.Println("🛑 Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
		} else {
			log.Println("✅ Server shutdown successfully")
		}

		// Close database connection
		if sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("❌ Database close error: %v", err)
			} else {
				log.Println("✅ Database connection closed")
			}
		}
	})
}

// FIX: Tambahkan parameter settingsController di sini
func setupRoutes(r *gin.Engine, authController *controllers.AuthController, configController *controllers.ConfigController, dbTestController *controllers.DBTestController, userRepo repositories.UserRepository, configService services.ConfigService, dashboardController *controllers.DashboardController, updateController *controllers.UpdateController, docService services.LostDocumentService, docController *controllers.LostDocumentController, itemTemplateController *controllers.ItemTemplateController, userController *controllers.UserController, backupController *controllers.BackupController, auditController *controllers.AuditLogController, licenseController *controllers.LicenseController, licenseService services.LicenseService, reportController *controllers.ReportController, settingsController *controllers.SettingsController) {
	r.GET("/login", func(c *gin.Context) { c.HTML(200, "login.html", gin.H{"Title": "Login"}) })
	r.POST("/api/login", middleware.LoginRateLimiter.GetLimiterMiddleware(), authController.Login)
	r.POST("/api/logout", authController.Logout)
	r.GET("/setup", configController.ShowSetupPage)
	r.POST("/api/setup", configController.SaveSetup)
	r.POST("/api/setup/restore", configController.RestoreSetup)
	r.POST("/api/db/test", dbTestController.TestConnection)

	authorized := r.Group("/")
	authorized.Use(middleware.SetupMiddleware(configService))
	if userRepo != nil { authorized.Use(middleware.AuthMiddleware(userRepo)) }

	authorized.GET("/", func(c *gin.Context) { c.HTML(200, "dashboard.html", gin.H{"Title": "Beranda", "CurrentUser": c.MustGet("currentUser"), "Config": mustGetConfig(configService)}) })
	authorized.GET("/api/config/limits", configController.GetLimits)
	authorized.GET("/api/stats", dashboardController.GetStats)
	authorized.GET("/api/stats/monthly-issuance", dashboardController.GetMonthlyChart)
	authorized.GET("/api/stats/item-composition", dashboardController.GetItemCompositionChart)
	authorized.GET("/api/notifications/expiring-documents", dashboardController.GetExpiringDocuments)
	authorized.GET("/api/updates/check", updateController.CheckUpdate)
	authorized.GET("/documents", func(c *gin.Context) { c.HTML(200, "document_list.html", gin.H{"Title": "Daftar Dokumen", "CurrentUser": c.MustGet("currentUser"), "PageType": "active"}) })
	authorized.GET("/documents/archived", func(c *gin.Context) { c.HTML(200, "document_list.html", gin.H{"Title": "Arsip Dokumen", "CurrentUser": c.MustGet("currentUser"), "PageType": "archived"}) })
	authorized.GET("/documents/new", func(c *gin.Context) { c.HTML(200, "document_form.html", gin.H{"Title": "Buat Surat Baru", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "DocID": 0}) })
	authorized.GET("/documents/:id/edit", func(c *gin.Context) { c.HTML(200, "document_form.html", gin.H{"Title": "Edit Surat", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "DocID": c.Param("id")}) })
	
	authorized.GET("/documents/:id/print", func(c *gin.Context) {
		docID := c.Param("id")
		var id uint
		fmt.Sscanf(docID, "%d", &id)
		userID := c.GetUint("userID")
		doc, err := docService.FindByID(id, userID)
		if err != nil { c.String(404, "Dokumen tidak ditemukan"); return }
		conf, _ := configService.GetConfig()
		archiveDays := 15
		if conf.ArchiveDurationDays > 0 { archiveDays = conf.ArchiveDurationDays }
		c.HTML(200, "print_preview.html", gin.H{"Document": doc, "Config": conf, "ArchiveDaysWords": utils.IntToIndonesianWords(archiveDays)})
	})

	authorized.POST("/api/documents", docController.Create)
	authorized.GET("/api/documents", docController.FindAll)
	authorized.GET("/api/documents/:id", docController.FindByID)
	authorized.GET("/api/documents/:id/pdf", docController.GetPDF)
	authorized.PUT("/api/documents/:id", docController.Update)
	authorized.DELETE("/api/documents/:id", docController.Delete)
	authorized.GET("/api/search", docController.SearchGlobal)
	authorized.GET("/search", func(c *gin.Context) { c.HTML(200, "search_results.html", gin.H{"Title": "Hasil Pencarian", "CurrentUser": c.MustGet("currentUser")}) })
	authorized.GET("/api/item-templates/active", itemTemplateController.FindAllActive)
	authorized.GET("/profile", func(c *gin.Context) { c.HTML(200, "profile.html", gin.H{"Title": "Profil Saya", "CurrentUser": c.MustGet("currentUser")}) })
	authorized.PUT("/api/profile", userController.UpdateProfile)
	authorized.PUT("/api/profile/password", userController.ChangePassword)
	authorized.GET("/panduan", func(c *gin.Context) { c.HTML(200, "panduan.html", gin.H{"Title": "Panduan", "CurrentUser": c.MustGet("currentUser")}) })
	authorized.GET("/upgrade", func(c *gin.Context) {
		conf, _ := configService.GetConfig()
		c.HTML(200, "upgrade.html", gin.H{"Title": "Upgrade ke Pro", "CurrentUser": c.MustGet("currentUser"), "Config": conf, "AppVersion": version})
	})
	authorized.GET("/tentang", func(c *gin.Context) {
		conf, _ := configService.GetConfig()
		c.HTML(200, "tentang.html", gin.H{"Title": "Tentang", "CurrentUser": c.MustGet("currentUser"), "AppVersion": version, "Config": conf})
	})

	admin := authorized.Group("/")
	admin.Use(middleware.AdminAuthMiddleware())
	admin.GET("/users", func(c *gin.Context) { c.HTML(200, "user_list.html", gin.H{"Title": "Manajemen Pengguna", "CurrentUser": c.MustGet("currentUser")}) })
	admin.GET("/users/new", func(c *gin.Context) { c.HTML(200, "user_form.html", gin.H{"Title": "Tambah Pengguna", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "UserID": 0}) })
	admin.GET("/users/:id/edit", func(c *gin.Context) { c.HTML(200, "user_form.html", gin.H{"Title": "Edit Pengguna", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "UserID": c.Param("id")}) })
	admin.POST("/api/users", userController.Create)
	admin.GET("/api/users", userController.FindAll)
	admin.GET("/api/users/operators", userController.FindOperators)
	admin.GET("/api/users/:id", userController.FindByID)
	admin.PUT("/api/users/:id", userController.Update)
	admin.DELETE("/api/users/:id", userController.Delete)
	admin.POST("/api/users/:id/activate", userController.Activate)
	
	admin.GET("/settings", func(c *gin.Context) { c.HTML(200, "settings.html", gin.H{"Title": "Pengaturan Sistem", "CurrentUser": c.MustGet("currentUser")}) })
	admin.GET("/api/settings", settingsController.GetSettings)
	admin.PUT("/api/settings", settingsController.UpdateSettings)
	admin.POST("/api/backups", backupController.CreateBackup)
	admin.POST("/api/restore", backupController.RestoreBackup)
	admin.POST("/api/settings/migrate", configController.MigrateDatabase)
	admin.GET("/api/settings/download-cert", settingsController.DownloadCertificate)
	
	admin.GET("/api/audit-logs", auditController.FindAll)
	admin.GET("/api/audit-logs/export", auditController.Export)
	admin.GET("/audit-logs", func(c *gin.Context) { c.HTML(200, "audit_log_list.html", gin.H{"Title": "Log Audit", "CurrentUser": c.MustGet("currentUser")}) })
	admin.GET("/api/documents/export", docController.Export)
	admin.POST("/api/license/activate", licenseController.ActivateLicense)
	admin.GET("/api/license/hwid", licenseController.GetHardwareID)

	pro := admin.Group("/")
	pro.Use(middleware.LicenseMiddleware(licenseService))
	pro.GET("/reports/aggregate", reportController.ShowReportPage)
	pro.GET("/api/reports/aggregate/pdf", reportController.GenerateReportPDF)
	pro.GET("/templates", func(c *gin.Context) { c.HTML(200, "item_template_list.html", gin.H{"Title": "Template Barang", "CurrentUser": c.MustGet("currentUser")}) })
	pro.GET("/templates/new", func(c *gin.Context) { c.HTML(200, "item_template_form.html", gin.H{"Title": "Tambah Template", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "TemplateID": 0}) })
	pro.GET("/templates/:id/edit", func(c *gin.Context) { c.HTML(200, "item_template_form.html", gin.H{"Title": "Edit Template", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "TemplateID": c.Param("id")}) })
	pro.GET("/api/item-templates", itemTemplateController.FindAll)
	pro.GET("/api/item-templates/:id", itemTemplateController.FindByID)
	pro.POST("/api/item-templates", itemTemplateController.Create)
	pro.PUT("/api/item-templates/:id", itemTemplateController.Update)
	pro.DELETE("/api/item-templates/:id", itemTemplateController.Delete)
}

// Helper Functions with improved error handling
func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	var db *gorm.DB
	var err error
	
	// Enhanced GORM configuration
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// Add more resilient configurations
		DisableForeignKeyConstraintWhenMigrating: false,
		SkipDefaultTransaction: false,
	}

	log.Printf("🔧 Connecting to database (dialect: %s)", cfg.DBDialect)

	switch cfg.DBDialect {
	case "mysql":
		tlsOption := "false"
		switch cfg.DBSSLMode {
		case "require":
			tlsOption = "skip-verify"
		case "verify-full":
			tlsOption = "true"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", 
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, tlsOption)
		
		log.Printf("🔧 MySQL DSN: %s:***@tcp(%s:%s)/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		
	case "postgres":
		sslMode := cfg.DBSSLMode
		if sslMode == "" { 
			sslMode = "disable" 
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", 
			cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, sslMode)
		
		log.Printf("🔧 PostgreSQL DSN: host=%s user=%s dbname=%s port=%s", cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		
	default:
		// SQLite - Enhanced error handling
		log.Printf("🔧 SQLite DSN: %s", cfg.DBDSN)
		
		// Ensure directory exists for SQLite
		if strings.Contains(cfg.DBDSN, "/") || strings.Contains(cfg.DBDSN, "\\") {
			dbDir := filepath.Dir(cfg.DBDSN)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create SQLite directory %s: %w", dbDir, err)
			}
		}
		
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil {
			// Enable foreign keys for SQLite
			db.Exec("PRAGMA foreign_keys = ON")
			db.Exec("PRAGMA journal_mode = WAL")
			db.Exec("PRAGMA synchronous = NORMAL")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Improved connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("🔧 Running database migrations...")

	// Enhanced AutoMigrate with better error handling
	err = db.AutoMigrate(
		&models.User{},
		&models.Resident{},
		&models.LostDocument{},
		&models.LostItem{},
		&models.AuditLog{},
		&models.Configuration{},
		&models.ItemTemplate{},
		&models.License{},
	)

	if err != nil {
		// Log detailed migration error
		log.Printf("❌ Migration failed: %v", err)
		
		// Check if it's the specific FOREIGN column issue
		if strings.Contains(err.Error(), "FOREIGN") && strings.Contains(err.Error(), "lost_documents") {
			log.Println("🔧 Detected FOREIGN column issue in lost_documents table")
			log.Println("💡 Suggested fix: Delete the corrupted database file and restart the application")
			log.Printf("💡 Database location: %s", cfg.DBDSN)
			
			return nil, fmt.Errorf("database migration failed due to corrupted schema (FOREIGN column issue): %w\n" +
				"Solution: Delete the database file and restart the application to create a fresh schema", err)
		}
		
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return db, nil
}

func setupEnvironment() {
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	
	// Load environment files with error handling
	if err := godotenv.Overload(envPath); err != nil {
		log.Printf("⚠️  Cannot load %s: %v", envPath, err)
	}
	
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Cannot load default .env: %v", err)
	}
	
	log.Printf("🔧 Environment loaded from: %s", envPath)
}

func setupLogging() {
	logPath := filepath.Join(utils.GetAppDataDir(), "logs", "simdokpol.log")
	
	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		log.Printf("⚠️  Cannot create log directory: %v", err)
		return
	}
	
	// Setup log rotation
	fileLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,   // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true,
	}
	
	// Multi-writer to both console and file
	mw := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(mw)
	
	// Set log flags for better debugging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	log.Printf("🔧 Logging configured: %s", logPath)
}

func mustGetConfig(s services.ConfigService) *dto.AppConfig {
	if s == nil {
		log.Printf("⚠️  ConfigService is nil in mustGetConfig")
		return &dto.AppConfig{} // Return empty config instead of nil
	}
	
	c, err := s.GetConfig()
	if err != nil {
		log.Printf("⚠️  Error getting config: %v", err)
		return &dto.AppConfig{} // Return empty config instead of nil
	}
	
	return c
}

func seedDefaultTemplates(db *gorm.DB) {
	if db == nil {
		log.Printf("⚠️  Database is nil, skipping template seeding")
		return
	}

	var count int64
	result := db.Model(&models.ItemTemplate{}).Unscoped().Count(&count)
	if result.Error != nil {
		log.Printf("⚠️  Error counting templates: %v", result.Error)
		return
	}

	if count > 0 {
		log.Printf("🔹 Templates already exist (%d), skipping seeding", count)
		return
	}

	log.Println("🔹 Seeding default templates...")
	
	templates := []models.ItemTemplate{
		{NamaBarang: "KTP", Urutan: 1, IsActive: true, FieldsConfig: models.JSONFieldArray{{Label: "NIK", Type: "text", DataLabel: "NIK", Regex: "^[0-9]{16}$", RequiredLength: 16, IsNumeric: true, Placeholder: "16 Digit NIK"}}},
		{NamaBarang: "SIM", Urutan: 2, IsActive: true, FieldsConfig: models.JSONFieldArray{{Label: "Golongan SIM", Type: "select", DataLabel: "Gol", Options: []string{"A", "B I", "B II", "C", "D"}}, {Label: "Nomor SIM", Type: "text", DataLabel: "No. SIM", Regex: "^[0-9]{12,14}$", MinLength: 12, MaxLength: 14, IsNumeric: true}}},
		{NamaBarang: "STNK", Urutan: 3, IsActive: true, FieldsConfig: models.JSONFieldArray{{Label: "Nomor Polisi", Type: "text", DataLabel: "No. Pol"}, {Label: "Nomor Rangka", Type: "text", DataLabel: "No. Rangka"}, {Label: "Nomor Mesin", Type: "text", DataLabel: "No. Mesin"}}},
		{NamaBarang: "BPKB", Urutan: 4, IsActive: true, FieldsConfig: models.JSONFieldArray{{Label: "Nomor BPKB", Type: "text", DataLabel: "No. BPKB"}, {Label: "Atas Nama", Type: "text", DataLabel: "a.n."}}},
		{NamaBarang: "IJAZAH", Urutan: 5, IsActive: true, FieldsConfig: models.JSONFieldArray{{Label: "Tingkat", Type: "select", DataLabel: "Tingkat", Options: []string{"SD", "SMP", "SMA", "D3", "S1", "S2"}}, {Label: "Nomor Ijazah", Type: "text", DataLabel: "No. Ijazah"}}},
		{NamaBarang: "ATM", Urutan: 6, IsActive: true, FieldsConfig: models.JSONFieldArray{{Label: "Nama Bank", Type: "select", DataLabel: "Bank", Options: []string{"BRI", "BCA", "Mandiri"}}, {Label: "Nomor Rekening", Type: "text", DataLabel: "No. Rek"}}},
		{NamaBarang: "LAINNYA", Urutan: 99, IsActive: true, FieldsConfig: models.JSONFieldArray{}},
	}

	result = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "nama_barang"}},
		DoNothing: true,
	}).Create(&templates)

	if result.Error != nil {
		log.Printf("⚠️  Error seeding templates: %v", result.Error)
		return
	}

	log.Printf("✅ Successfully seeded %d templates", len(templates))
}

func onReady(isHTTPS bool, port string) {
	iconData := web.GetIconBytes()
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	} else {
		systray.SetTitle("SIMDOKPOL")
	}
	
	systray.SetTooltip("SIMDOKPOL Berjalan")
	
	mOpen := systray.AddMenuItem("Buka Aplikasi", "Buka di Browser")
	mVhost := systray.AddMenuItem("Setup Domain (simdokpol.local)", "Konfigurasi Virtual Host")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Keluar", "Hentikan Server")

	protocol := "http"
	if isHTTPS {
		protocol = "https"
	}
	
	// Auto-open browser after server starts
	go func() {
		time.Sleep(2 * time.Second)
		
		vhost := utils.NewVHostSetup()
		isVhost, err := vhost.IsSetup()
		if err != nil {
			log.Printf("⚠️  VHost setup check error: %v", err)
		}
		
		url := fmt.Sprintf("%s://localhost:%s", protocol, port)
		if isVhost {
			url = vhost.GetURL(port)
			if isHTTPS {
				url = strings.Replace(url, "http://", "https://", 1)
			}
		}
		
		log.Printf("🌐 Opening browser: %s", url)
		openBrowser(url)
	}()

	// Handle system tray clicks
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				url := fmt.Sprintf("%s://localhost:%s", protocol, port)
				log.Printf("🌐 Manual browser open: %s", url)
				openBrowser(url)
				
			case <-mVhost.ClickedCh:
				log.Println("🔧 Setting up virtual host...")
				vhost := utils.NewVHostSetup()
				if err := vhost.Setup(); err != nil {
					log.Printf("❌ VHost setup failed: %v", err)
					_ = beeep.Alert("Gagal", "Butuh hak akses Administrator.", "")
				} else {
					log.Println("✅ VHost setup successful")
					_ = beeep.Notify("Sukses", "Domain dikonfigurasi!", "")
				}
				
			case <-mQuit.ClickedCh:
				log.Println("🛑 User requested quit from system tray")
				systray.Quit()
			}
		}
	}()
}

func openBrowser(url string) {
	var err error
	
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	
	if err != nil {
		log.Printf("❌ Browser open error: %v", err)
	} else {
		log.Printf("✅ Browser opened successfully")
	}
}