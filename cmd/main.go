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
	"gorm.io/gorm/logger"
)

var (
	version         = "dev"
	changelogBase64 = ""
)

func main() {
	setupEnvironment()
	setupLogging()
	
	appData := utils.GetAppDataDir()
	log.Println("==========================================")
	log.Printf("🚀 SIMDOKPOL STARTUP - v%s", version)
	log.Printf("📂 App Data Dir: %s", appData)
	log.Println("==========================================")

	cfg := config.LoadConfig()

	db, err := setupDatabase(cfg)
	if err != nil {
		msg := fmt.Sprintf("FATAL: Gagal koneksi database: %v", err)
		_ = beeep.Alert("SIMDOKPOL Error", msg, "")
		log.Fatal(msg) // ✅ PERBAIKAN: gunakan log.Fatal bukan log.Fatalf
	}

	if cfg.DBDialect == "sqlite" {
		seedDefaultTemplates(db)
	}

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
	
	// FIX: Inject ConfigService ke Auth (untuk timeout session)
	authService := services.NewAuthService(userRepo, configService)
	
	migrationService := services.NewDataMigrationService(db, auditService, configService)

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	docService := services.NewLostDocumentService(db, docRepo, residentRepo, userRepo, auditService, configService, configRepo, exeDir)
	dashboardService := services.NewDashboardService(docRepo, userRepo, configService)
	reportService := services.NewReportService(docRepo, configService, exeDir)
	itemTemplateService := services.NewItemTemplateService(itemTemplateRepo)
	dbTestService := services.NewDBTestService()
	updateService := services.NewUpdateService()

	// --- CONTROLLERS ---
	// FIX: Inject ConfigService ke AuthController (untuk cookie max age)
	authController := controllers.NewAuthController(authService, configService)
	
	userController := controllers.NewUserController(userService)
	docController := controllers.NewLostDocumentController(docService)
	dashboardController := controllers.NewDashboardController(dashboardService)
	configController := controllers.NewConfigController(configService, userService, backupService, migrationService)
	auditController := controllers.NewAuditLogController(auditService)
	backupController := controllers.NewBackupController(backupService)
	settingsController := controllers.NewSettingsController(configService, auditService)
	licenseController := controllers.NewLicenseController(licenseService, auditService)
	reportController := controllers.NewReportController(reportService, configService)
	itemTemplateController := controllers.NewItemTemplateController(itemTemplateService)
	dbTestController := controllers.NewDBTestController(dbTestService)
	updateController := controllers.NewUpdateController(updateService, version)

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.MaxMultipartMemory = 8 << 20

	funcMap := template.FuncMap{"ToUpper": strings.ToUpper}
	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(web.Assets, "templates/*.html", "templates/partials/*.html"))
	r.SetHTMLTemplate(templ)
	r.StaticFS("/static", web.GetStaticFS())

	r.Use(func(c *gin.Context) {
		c.Set("AppVersion", version)
		decodedChangelog, _ := utils.DecodeBase64(changelogBase64)
		c.Set("AppChangelog", decodedChangelog)
		c.Next()
	})

	// --- ROUTES ---
	r.GET("/login", func(c *gin.Context) { c.HTML(200, "login.html", gin.H{"Title": "Login"}) })
	r.POST("/api/login", middleware.LoginRateLimiter.GetLimiterMiddleware(), authController.Login)
	r.POST("/api/logout", authController.Logout)
	r.GET("/setup", configController.ShowSetupPage)
	r.POST("/api/setup", configController.SaveSetup)
	r.POST("/api/setup/restore", configController.RestoreSetup)
	r.POST("/api/db/test", dbTestController.TestConnection)

	authorized := r.Group("/")
	authorized.Use(middleware.SetupMiddleware(configService))
	authorized.Use(middleware.AuthMiddleware(userRepo))

	authorized.GET("/", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{"Title": "Beranda", "CurrentUser": c.MustGet("currentUser"), "Config": mustGetConfig(configService)})
	})
	
	// FIX: Route Limit Config (Public untuk Frontend)
	authorized.GET("/api/config/limits", configController.GetLimits)

	authorized.GET("/api/stats", dashboardController.GetStats)
	authorized.GET("/api/stats/monthly-issuance", dashboardController.GetMonthlyChart)
	authorized.GET("/api/stats/item-composition", dashboardController.GetItemCompositionChart)
	authorized.GET("/api/notifications/expiring-documents", dashboardController.GetExpiringDocuments)
	authorized.GET("/api/updates/check", updateController.CheckUpdate)

	authorized.GET("/documents", func(c *gin.Context) {
		c.HTML(200, "document_list.html", gin.H{"Title": "Daftar Dokumen", "CurrentUser": c.MustGet("currentUser"), "PageType": "active"})
	})
	authorized.GET("/documents/archived", func(c *gin.Context) {
		c.HTML(200, "document_list.html", gin.H{"Title": "Arsip Dokumen", "CurrentUser": c.MustGet("currentUser"), "PageType": "archived"})
	})
	authorized.GET("/documents/new", func(c *gin.Context) {
		c.HTML(200, "document_form.html", gin.H{"Title": "Buat Surat Baru", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "DocID": 0})
	})
	authorized.GET("/documents/:id/edit", func(c *gin.Context) {
		c.HTML(200, "document_form.html", gin.H{"Title": "Edit Surat", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "DocID": c.Param("id")})
	})

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
		if conf.ArchiveDurationDays > 0 {
			archiveDays = conf.ArchiveDurationDays
		}
		c.HTML(200, "print_preview.html", gin.H{"Document": doc, "Config": conf, "ArchiveDaysWords": utils.IntToIndonesianWords(archiveDays)})
	})

	authorized.POST("/api/documents", docController.Create)
	authorized.GET("/api/documents", docController.FindAll)
	authorized.GET("/api/documents/:id", docController.FindByID)
	authorized.GET("/api/documents/:id/pdf", docController.GetPDF)
	authorized.PUT("/api/documents/:id", docController.Update)
	authorized.DELETE("/api/documents/:id", docController.Delete)
	authorized.GET("/api/search", docController.SearchGlobal)
	authorized.GET("/search", func(c *gin.Context) {
		c.HTML(200, "search_results.html", gin.H{"Title": "Hasil Pencarian", "CurrentUser": c.MustGet("currentUser")})
	})

	authorized.GET("/api/item-templates/active", itemTemplateController.FindAllActive)

	authorized.GET("/profile", func(c *gin.Context) {
		c.HTML(200, "profile.html", gin.H{"Title": "Profil Saya", "CurrentUser": c.MustGet("currentUser")})
	})
	authorized.PUT("/api/profile", userController.UpdateProfile)
	authorized.PUT("/api/profile/password", userController.ChangePassword)

	authorized.GET("/panduan", func(c *gin.Context) {
		c.HTML(200, "panduan.html", gin.H{"Title": "Panduan", "CurrentUser": c.MustGet("currentUser")})
	})
	authorized.GET("/tentang", func(c *gin.Context) {
		conf, _ := configService.GetConfig()
		c.HTML(200, "tentang.html", gin.H{"Title": "Tentang", "CurrentUser": c.MustGet("currentUser"), "AppVersion": version, "Config": conf})
	})

	admin := authorized.Group("/")
	admin.Use(middleware.AdminAuthMiddleware())

	admin.GET("/users", func(c *gin.Context) {
		c.HTML(200, "user_list.html", gin.H{"Title": "Manajemen Pengguna", "CurrentUser": c.MustGet("currentUser")})
	})
	admin.GET("/users/new", func(c *gin.Context) {
		c.HTML(200, "user_form.html", gin.H{"Title": "Tambah Pengguna", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "UserID": 0})
	})
	admin.GET("/users/:id/edit", func(c *gin.Context) {
		c.HTML(200, "user_form.html", gin.H{"Title": "Edit Pengguna", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "UserID": c.Param("id")})
	})
	admin.POST("/api/users", userController.Create)
	admin.GET("/api/users", userController.FindAll)
	admin.GET("/api/users/operators", userController.FindOperators)
	admin.GET("/api/users/:id", userController.FindByID)
	admin.PUT("/api/users/:id", userController.Update)
	admin.DELETE("/api/users/:id", userController.Delete)
	admin.POST("/api/users/:id/activate", userController.Activate)

	admin.GET("/settings", func(c *gin.Context) {
		c.HTML(200, "settings.html", gin.H{"Title": "Pengaturan Sistem", "CurrentUser": c.MustGet("currentUser")})
	})
	admin.GET("/api/settings", settingsController.GetSettings)
	admin.PUT("/api/settings", settingsController.UpdateSettings)
	admin.POST("/api/backups", backupController.CreateBackup)
	admin.POST("/api/restore", backupController.RestoreBackup)
	admin.POST("/api/settings/migrate", configController.MigrateDatabase)

	admin.GET("/api/audit-logs", auditController.FindAll)
	admin.GET("/api/audit-logs/export", auditController.Export)
	admin.GET("/audit-logs", func(c *gin.Context) {
		c.HTML(200, "audit_log_list.html", gin.H{"Title": "Log Audit", "CurrentUser": c.MustGet("currentUser")})
	})
	admin.GET("/api/documents/export", docController.Export)

	admin.POST("/api/license/activate", licenseController.ActivateLicense)
	admin.GET("/api/license/hwid", licenseController.GetHardwareID)

	pro := admin.Group("/")
	pro.Use(middleware.LicenseMiddleware(licenseService))
	pro.GET("/reports/aggregate", reportController.ShowReportPage)
	pro.GET("/api/reports/aggregate/pdf", reportController.GenerateReportPDF)
	pro.GET("/templates", func(c *gin.Context) {
		c.HTML(200, "item_template_list.html", gin.H{"Title": "Template Barang", "CurrentUser": c.MustGet("currentUser")})
	})
	pro.GET("/templates/new", func(c *gin.Context) {
		c.HTML(200, "item_template_form.html", gin.H{"Title": "Tambah Template", "CurrentUser": c.MustGet("currentUser"), "IsEdit": false, "TemplateID": 0})
	})
	pro.GET("/templates/:id/edit", func(c *gin.Context) {
		c.HTML(200, "item_template_form.html", gin.H{"Title": "Edit Template", "CurrentUser": c.MustGet("currentUser"), "IsEdit": true, "TemplateID": c.Param("id")})
	})
	pro.GET("/api/item-templates", itemTemplateController.FindAll)
	pro.GET("/api/item-templates/:id", itemTemplateController.FindByID)
	pro.POST("/api/item-templates", itemTemplateController.Create)
	pro.PUT("/api/item-templates/:id", itemTemplateController.Update)
	pro.DELETE("/api/item-templates/:id", itemTemplateController.Delete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	isHTTPS := os.Getenv("ENABLE_HTTPS") == "true"
	log.Printf("🔍 Konfigurasi HTTPS: %v", isHTTPS)
	
	var certFile, keyFile string
	var errHTTPS error

	if isHTTPS {
		certFile, keyFile, errHTTPS = utils.EnsureCertificates()
		if errHTTPS != nil {
			log.Printf("⚠️ ERROR CERT: Gagal membuat sertifikat: %v", errHTTPS)
			isHTTPS = false
		} else {
			log.Printf("✅ Cert Path: %s", certFile)
		}
	}

	go func() {
		if isHTTPS {
			log.Printf("🔒 Server berjalan di port %s (HTTPS Enabled)", port)
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenTLS error: %s\n", err)
			}
		} else {
			log.Printf("🌐 Server berjalan di port %s (HTTP Standar)", port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}
	}()

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-quitChan; systray.Quit() }()

	systray.Run(func() { onReady(isHTTPS) }, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Shutdown error:", err)
		}
	})
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}

	switch cfg.DBDialect {
	case "mysql":
		// FIX: Support SSL di MySQL
		tlsOption := "false"
		if cfg.DBSSLMode == "require" {
			tlsOption = "skip-verify"
		}
		if cfg.DBSSLMode == "verify-full" {
			tlsOption = "true"
		}
		
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", 
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, tlsOption)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		sslMode := cfg.DBSSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", 
			cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, sslMode)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default: 
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil {
			db.Exec("PRAGMA foreign_keys = ON")
		}
	}

	if err != nil {
		return nil, err
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if cfg.DBDialect == "sqlite" {
		err = db.AutoMigrate(&models.User{}, &models.Resident{}, &models.LostDocument{}, &models.LostItem{}, &models.AuditLog{}, &models.Configuration{}, &models.ItemTemplate{}, &models.License{})
		if err != nil {
			return nil, fmt.Errorf("migrasi gagal: %w", err)
		}
		seedDefaultTemplates(db)
	}
	return db, nil
}

func seedDefaultTemplates(db *gorm.DB) {
	var count int64
	db.Model(&models.ItemTemplate{}).Count(&count)
	if count > 0 {
		return
	}
	
	log.Println("🔹 Seeding default templates...")
	templates := []models.ItemTemplate{
		{
			NamaBarang: "KTP", 
			Urutan: 1, 
			IsActive: true,
			FieldsConfig: models.JSONFieldArray{
				{Label: "NIK", Type: "text", DataLabel: "NIK", RequiredLength: 16, IsNumeric: true},
			},
		},
		{
			NamaBarang: "SIM", 
			Urutan: 2, 
			IsActive: true, 
			FieldsConfig: models.JSONFieldArray{
				{Label: "Golongan", Type: "select", DataLabel: "Gol", Options: []string{"A", "C", "B I", "B II", "D"}},
				{Label: "Nomor SIM", Type: "text", DataLabel: "No. SIM", MinLength: 12, IsNumeric: true},
			},
		},
		{
			NamaBarang: "STNK", 
			Urutan: 3, 
			IsActive: true, 
			FieldsConfig: models.JSONFieldArray{},
		},
		{
			NamaBarang: "BPKB", 
			Urutan: 4, 
			IsActive: true, 
			FieldsConfig: models.JSONFieldArray{},
		},
		{
			NamaBarang: "ATM", 
			Urutan: 5, 
			IsActive: true, 
			FieldsConfig: models.JSONFieldArray{},
		},
		{
			NamaBarang: "LAINNYA", 
			Urutan: 99, 
			IsActive: true, 
			FieldsConfig: models.JSONFieldArray{},
		},
	}
	
	if err := db.Create(&templates).Error; err != nil {
		log.Printf("⚠️ Gagal seeding templates: %v", err)
	} else {
		log.Println("✅ Default templates berhasil di-seed")
	}
}

func setupEnvironment() {
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️ Tidak bisa load .env dari AppData: %v", err)
	}
	
	// Fallback ke .env di current directory
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ Tidak bisa load .env dari current directory: %v", err)
	}
}

func setupLogging() {
	logPath := filepath.Join(utils.GetAppDataDir(), "logs", "simdokpol.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		log.Printf("⚠️ Gagal buat directory logs: %v", err)
	}
	
	fileLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}
	
	// Dual Log (File + Terminal)
	mw := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func onReady(isHTTPS bool) {
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

	go func() {
		time.Sleep(1 * time.Second)
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		
		vhost := utils.NewVHostSetup()
		isVhost, _ := vhost.IsSetup()
		url := fmt.Sprintf("%s://localhost:%s", protocol, port)
		
		if isVhost {
			url = vhost.GetURL(port)
			if isHTTPS {
				url = strings.Replace(url, "http://", "https://", 1)
			}
		}
		
		openBrowser(url)
	}()

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				port := os.Getenv("PORT")
				if port == "" {
					port = "8080"
				}
				openBrowser(fmt.Sprintf("%s://localhost:%s", protocol, port))
				
			case <-mVhost.ClickedCh:
				vhost := utils.NewVHostSetup()
				if err := vhost.Setup(); err != nil {
					_ = beeep.Alert("Gagal", "Butuh Administrator.", "")
				} else {
					_ = beeep.Notify("Sukses", "Domain dikonfigurasi!", "")
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
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("sistem operasi tidak didukung: %s", runtime.GOOS)
	}
	
	if err != nil {
		log.Printf("Gagal buka browser: %v", err)
	}
}

func mustGetConfig(s services.ConfigService) *dto.AppConfig {
	c, _ := s.GetConfig()
	return c
}