package main

import (
	"context"
	"fmt"
	"html/template"
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
	"gorm.io/gorm/logger"
)

var (
	version         = "dev"
	changelogBase64 = ""
)

func main() {
	// 1. Setup Environment & Logging
	setupEnvironment()
	setupLogging()

	log.Println("=== MEMULAI SIMDOKPOL ===")
	
	// Load Config (Secure)
	cfg := config.LoadConfig()

	// 2. Setup Database Connection (Dengan Error Handling)
	db, err := setupDatabase(cfg)
	if err != nil {
		// Jika GUI, tampilkan alert native OS
		msg := fmt.Sprintf("Gagal koneksi database: %v", err)
		_ = beeep.Alert("SIMDOKPOL Error", msg, "")
		// Log fatal agar aplikasi berhenti
		log.Fatalf("Gagal koneksi database: %v", err)
	}

	// 3. Setup Layers (Dependency Injection)
	
	// --- REPOSITORIES ---
	userRepo := repositories.NewUserRepository(db)
	docRepo := repositories.NewLostDocumentRepository(db)
	residentRepo := repositories.NewResidentRepository(db)
	configRepo := repositories.NewConfigRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	licenseRepo := repositories.NewLicenseRepository(db)
	itemTemplateRepo := repositories.NewItemTemplateRepository(db)

	// --- SERVICES ---
	auditService := services.NewAuditLogService(auditRepo)
	configService := services.NewConfigService(configRepo)
	backupService := services.NewBackupService(db, cfg, configService, auditService)
	licenseService := services.NewLicenseService(licenseRepo, configService, auditService)
	userService := services.NewUserService(userRepo, auditService, cfg)
	dataMigrationService := services.NewDataMigrationService(db, auditService, configService)
	
	// Dapatkan path executable untuk keperluan logging/backup path fallback
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	
	docService := services.NewLostDocumentService(db, docRepo, residentRepo, userRepo, auditService, configService, configRepo, exeDir)
	dashboardService := services.NewDashboardService(docRepo, userRepo, configService)
	reportService := services.NewReportService(docRepo, configService, exeDir)
	itemTemplateService := services.NewItemTemplateService(itemTemplateRepo)
	dbTestService := services.NewDBTestService()
	updateService := services.NewUpdateService()
	authService := services.NewAuthService(userRepo)

	// --- CONTROLLERS ---
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	docController := controllers.NewLostDocumentController(docService)
	dashboardController := controllers.NewDashboardController(dashboardService)
	configController := controllers.NewConfigController(configService, userService, backupService, dataMigrationService)
	auditController := controllers.NewAuditLogController(auditService)
	backupController := controllers.NewBackupController(backupService)
	settingsController := controllers.NewSettingsController(configService, auditService)
	licenseController := controllers.NewLicenseController(licenseService, auditService)
	reportController := controllers.NewReportController(reportService, configService)
	itemTemplateController := controllers.NewItemTemplateController(itemTemplateService)
	dbTestController := controllers.NewDBTestController(dbTestService)
	updateController := controllers.NewUpdateController(updateService, version)

	// 4. Setup Router Gin
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Middleware Global
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.MaxMultipartMemory = 8 << 20 // Max upload 8MB

	// Setup Templates dari Embed FS
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	// Parse template dari memory (embed)
	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(web.Assets, "templates/*.html", "templates/partials/*.html"))
	r.SetHTMLTemplate(templ)

	// Setup Static Files dari Embed FS
	staticFS := web.GetStaticFS()
	r.StaticFS("/static", staticFS)

	// Middleware Inject Variabel Global ke Template
	r.Use(func(c *gin.Context) {
		c.Set("AppVersion", version)
		decodedChangelog, _ := utils.DecodeBase64(changelogBase64)
		c.Set("AppChangelog", decodedChangelog)
		c.Next()
	})

	// --- ROUTES ---
	
	// Public Routes
	r.GET("/login", func(c *gin.Context) { c.HTML(200, "login.html", gin.H{"Title": "Login"}) })
	// Rate limiting khusus login
	r.POST("/api/login", middleware.LoginRateLimiter.GetLimiterMiddleware(), authController.Login)
	r.POST("/api/logout", authController.Logout)
	
	// Setup Wizard Routes
	r.GET("/setup", configController.ShowSetupPage)
	r.POST("/api/setup", configController.SaveSetup)
	r.POST("/api/setup/restore", configController.RestoreSetup)
	r.POST("/api/db/test", dbTestController.TestConnection)

	// Protected Routes (Butuh Login)
	authorized := r.Group("/")
	authorized.Use(middleware.SetupMiddleware(configService)) // Cek setup dulu
	authorized.Use(middleware.AuthMiddleware(userRepo))       // Baru cek login

	// Dashboard
	authorized.GET("/", func(c *gin.Context) { 
		c.HTML(200, "dashboard.html", gin.H{
			"Title": "Dasbor", 
			"CurrentUser": c.MustGet("currentUser"),
			"Config": mustGetConfig(configService), // Helper untuk ambil config di template
		}) 
	})
	
	// API Statistik
	authorized.GET("/api/stats", dashboardController.GetStats)
	authorized.GET("/api/stats/monthly-issuance", dashboardController.GetMonthlyChart)
	authorized.GET("/api/stats/item-composition", dashboardController.GetItemCompositionChart)
	authorized.GET("/api/notifications/expiring-documents", dashboardController.GetExpiringDocuments)
	authorized.GET("/api/updates/check", updateController.CheckUpdate)

	// Dokumen Routes
	authorized.GET("/documents", func(c *gin.Context) { c.HTML(200, "document_list.html", gin.H{"Title": "Daftar Dokumen", "CurrentUser": c.MustGet("currentUser"), "PageType": "active"}) })
	authorized.GET("/documents/archived", func(c *gin.Context) { c.HTML(200, "document_list.html", gin.H{"Title": "Arsip Dokumen", "CurrentUser": c.MustGet("currentUser"), "PageType": "archived"}) })
	authorized.GET("/documents/new", func(c *gin.Context) { c.HTML(200, "document_form.html", gin.H{"Title": "Buat Surat Baru", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "DocID": 0}) })
	authorized.GET("/documents/:id/edit", func(c *gin.Context) { c.HTML(200, "document_form.html", gin.H{"Title": "Edit Surat", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "DocID": c.Param("id")}) })
	
	// Print Preview
	authorized.GET("/documents/:id/print", func(c *gin.Context) {
		docID := c.Param("id")
		var id uint
		fmt.Sscanf(docID, "%d", &id)
		userID := c.GetUint("userID")
		doc, err := docService.FindByID(id, userID)
		if err != nil {
			c.String(404, "Dokumen tidak ditemukan")
			return
		}
		conf, _ := configService.GetConfig()
		archiveDays := 15
		if conf.ArchiveDurationDays > 0 { archiveDays = conf.ArchiveDurationDays }
		// Convert angka ke terbilang
		archiveDaysWords := utils.IntToIndonesianWords(archiveDays)
		c.HTML(200, "print_preview.html", gin.H{"Document": doc, "Config": conf, "ArchiveDaysWords": archiveDaysWords})
	})

	// API Dokumen (CRUD)
	authorized.POST("/api/documents", docController.Create)
	authorized.GET("/api/documents", docController.FindAll) // Server-side paging
	authorized.GET("/api/documents/:id", docController.FindByID)
	authorized.GET("/api/documents/:id/pdf", docController.GetPDF)
	authorized.PUT("/api/documents/:id", docController.Update)
	authorized.DELETE("/api/documents/:id", docController.Delete)
	authorized.GET("/api/search", docController.SearchGlobal)
	authorized.GET("/search", func(c *gin.Context) { c.HTML(200, "search_results.html", gin.H{"Title": "Hasil Pencarian", "CurrentUser": c.MustGet("currentUser")}) })

	// User Profile Routes
	authorized.GET("/profile", func(c *gin.Context) { c.HTML(200, "profile.html", gin.H{"Title": "Profil Saya", "CurrentUser": c.MustGet("currentUser")}) })
	authorized.PUT("/api/profile", userController.UpdateProfile)
	authorized.PUT("/api/profile/password", userController.ChangePassword)
	
	// Static Info Pages
	authorized.GET("/panduan", func(c *gin.Context) { c.HTML(200, "panduan.html", gin.H{"Title": "Panduan", "CurrentUser": c.MustGet("currentUser")}) })
	authorized.GET("/tentang", func(c *gin.Context) { 
		conf, _ := configService.GetConfig()
		c.HTML(200, "tentang.html", gin.H{"Title": "Tentang", "CurrentUser": c.MustGet("currentUser"), "AppVersion": version, "Config": conf}) 
	})

	// --- ADMIN ROUTES ---
	admin := authorized.Group("/")
	admin.Use(middleware.AdminAuthMiddleware())
	
	// User Management
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

	// Settings & Maintenance
	admin.GET("/settings", func(c *gin.Context) { c.HTML(200, "settings.html", gin.H{"Title": "Pengaturan Sistem", "CurrentUser": c.MustGet("currentUser")}) })
	admin.GET("/api/settings", settingsController.GetSettings)
	admin.PUT("/api/settings", settingsController.UpdateSettings)
	admin.POST("/api/backups", backupController.CreateBackup)
	admin.POST("/api/restore", backupController.RestoreBackup)
	
	// Audit Logs
	admin.GET("/api/audit-logs", auditController.FindAll)
	admin.GET("/api/audit-logs/export", auditController.Export)
	admin.GET("/audit-logs", func(c *gin.Context) { c.HTML(200, "audit_log_list.html", gin.H{"Title": "Log Audit", "CurrentUser": c.MustGet("currentUser")}) })
	
	// Export Data Dokumen
	admin.GET("/api/documents/export", docController.Export)
	
	// License Management
	admin.POST("/api/license/activate", licenseController.ActivateLicense)
	admin.GET("/api/license/hwid", licenseController.GetHardwareID)

	// --- PRO ROUTES (Butuh Lisensi Valid) ---
	pro := admin.Group("/")
	pro.Use(middleware.LicenseMiddleware(licenseService))
	
	// Reports
	pro.GET("/reports/aggregate", reportController.ShowReportPage)
	pro.GET("/api/reports/aggregate/pdf", reportController.GenerateReportPDF)
	
	// Template Editor
	pro.GET("/templates", func(c *gin.Context) { c.HTML(200, "item_template_list.html", gin.H{"Title": "Template Barang", "CurrentUser": c.MustGet("currentUser")}) })
	pro.GET("/templates/new", func(c *gin.Context) { c.HTML(200, "item_template_form.html", gin.H{"Title": "Tambah Template", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "TemplateID": 0}) })
	pro.GET("/templates/:id/edit", func(c *gin.Context) { c.HTML(200, "item_template_form.html", gin.H{"Title": "Edit Template", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "TemplateID": c.Param("id")}) })
	pro.GET("/api/item-templates", itemTemplateController.FindAll)
	pro.GET("/api/item-templates/active", itemTemplateController.FindAllActive)
	pro.GET("/api/item-templates/:id", itemTemplateController.FindByID)
	pro.POST("/api/item-templates", itemTemplateController.Create)
	pro.PUT("/api/item-templates/:id", itemTemplateController.Update)
	pro.DELETE("/api/item-templates/:id", itemTemplateController.Delete)

	// 5. Server Start with Graceful Shutdown
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Jalankan server di Goroutine
	go func() {
		// Notifikasi Desktop
		_ = beeep.Notify("SIMDOKPOL Berjalan", fmt.Sprintf("Akses di: http://localhost:%s", port), "")
		log.Printf("Server running on port %s", port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Systray Logic
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Goroutine untuk listen sinyal Ctrl+C (Terminal)
	go func() {
		<-quitChan
		log.Println("Signal received, shutting down systray...")
		systray.Quit() // Ini akan memicu onExit
	}()

	// Main thread diblokir oleh Systray (Required for GUI on Mac/Win)
	systray.Run(onReady, func() {
		// onExit: Shutdown server Gin
		log.Println("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}
		log.Println("Server exited")
	})
}

// setupDatabase: Mengembalikan error agar bisa di-handle main
func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	switch cfg.DBDialect {
	case "mysql":
		// MySQL DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		// PostgreSQL DSN
		sslMode := os.Getenv("DB_SSLMODE")
		if sslMode == "" { sslMode = "disable" }
		
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, sslMode)
			
		if rootCert := os.Getenv("DB_ROOT_CERT"); rootCert != "" {
			dsn += fmt.Sprintf(" sslrootcert=%s", rootCert)
		}
		
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default: // sqlite (Default)
		dir := filepath.Dir(cfg.DBDSN)
		if dir != "." && dir != "" { _ = os.MkdirAll(dir, 0755) }
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil { 
			// Enable Foreign Keys for SQLite
			db.Exec("PRAGMA foreign_keys = ON") 
		}
	}

	if err != nil {
		return nil, err
	}

	// Connection Pooling
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto Migrate HANYA untuk SQLite (Portable Mode)
	// Untuk MySQL/Postgres, gunakan file SQL migration manual
	if cfg.DBDialect == "sqlite" {
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
			return nil, fmt.Errorf("gagal migrasi database: %w", err)
		}
	}
	return db, nil
}

func setupEnvironment() {
	// Load dari AppData folder dulu (prioritas)
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Load(envPath)
	// Load dari folder lokal (fallback dev)
	_ = godotenv.Load()
}

func setupLogging() {
	logPath := filepath.Join(utils.GetAppDataDir(), "logs", "simdokpol.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)

	// Log ke file dengan rotasi + konsol
	log.SetOutput(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     28, // Hari
		Compress:   true,
	})
	// Note: lumberjack replace stdout, jika mau dua-duanya butuh io.MultiWriter (tapi io.MultiWriter tidak support rotasi native)
	// Untuk production desktop, log file lebih penting.
}

func onReady() {
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

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				port := os.Getenv("PORT")
				if port == "" { port = "8080" }
				// Cek apakah vhost aktif untuk buka URL yg cantik
				vhost := utils.NewVHostSetup()
				isVhost, _ := vhost.IsSetup()
				if isVhost {
					openBrowser(vhost.GetURL(port))
				} else {
					openBrowser(fmt.Sprintf("http://localhost:%s", port))
				}
			case <-mVhost.ClickedCh:
				vhost := utils.NewVHostSetup()
				if err := vhost.Setup(); err != nil {
					_ = beeep.Alert("Gagal Setup VHost", "Pastikan Anda menjalankan aplikasi sebagai Administrator/Root.", "")
				} else {
					_ = beeep.Notify("Sukses", "Domain simdokpol.local berhasil dikonfigurasi!", "")
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux": err = exec.Command("xdg-open", url).Start()
	case "windows": err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": err = exec.Command("open", url).Start()
	default: err = fmt.Errorf("unsupported platform")
	}
	if err != nil { log.Printf("Gagal membuka browser: %v", err) }
}

// Helper untuk ambil config di template dashboard
func mustGetConfig(s services.ConfigService) *dto.AppConfig {
	c, _ := s.GetConfig()
	return c
}