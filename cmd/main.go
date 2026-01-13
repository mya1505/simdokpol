package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	shutdownChan    = make(chan struct{})
)

func main() {
	setupEnvironment()
	initializeSecrets()
	systray.Run(onReady, onExit)
}

func onReady() {
	setupLogging()

	appData := utils.GetAppDataDir()

	h := sha256.Sum256([]byte(services.AppSecretKeyString))
	keyHash := fmt.Sprintf("%x", h[:4])

	log.Println("==========================================")
	log.Printf("🚀 SIMDOKPOL DESKTOP - v%s", version)
	log.Printf("💻 Mode: DESKTOP (GUI/SYSTRAY)")
	log.Printf("📂 Data Dir: %s", appData)
	log.Printf("🔑 Secret Hash: %s...", keyHash)
	log.Println("==========================================")

	systray.SetIcon(web.GetIconBytes())
	systray.SetTitle("SIMDOKPOL")
	systray.SetTooltip("Sistem Informasi Manajemen Dokumen Kepolisian")

	mOpen := systray.AddMenuItem("Buka Aplikasi", "Buka di Browser")
	mQuit := systray.AddMenuItem("Keluar", "Hentikan Server")

	vhost := utils.NewVHostSetup()
	isVhostSetup, _ := vhost.IsSetup()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var appURL string
	if isVhostSetup {
		appURL = vhost.GetURL(port)
	} else {
		appURL = fmt.Sprintf("http://localhost:%s", port)
	}

	cfg := config.LoadConfig()

	db, err := setupDatabase(cfg)
	if err != nil {
		log.Printf("❌ GAGAL KONEKSI DATABASE: %v. Cek config/restart.", err)
		beeep.Alert("SIMDOKPOL Error", "Gagal koneksi database. Cek log.", "assets/warning.png")
	} else {
		seedDefaultTemplates(db)
	}

	var userRepo repositories.UserRepository
	var docRepo repositories.LostDocumentRepository
	var residentRepo repositories.ResidentRepository
	var configRepo repositories.ConfigRepository
	var auditRepo repositories.AuditLogRepository
	var licenseRepo repositories.LicenseRepository
	var itemTemplateRepo repositories.ItemTemplateRepository
	var jobPositionRepo repositories.JobPositionRepository

	if db != nil {
		userRepo = repositories.NewUserRepository(db)
		docRepo = repositories.NewLostDocumentRepository(db)
		residentRepo = repositories.NewResidentRepository(db)
		configRepo = repositories.NewConfigRepository(db)
		auditRepo = repositories.NewAuditLogRepository(db)
		licenseRepo = repositories.NewLicenseRepository(db)
		itemTemplateRepo = repositories.NewItemTemplateRepository(db)
		jobPositionRepo = repositories.NewJobPositionRepository(db)
	}

	auditService := services.NewAuditLogService(auditRepo)
	configService := services.NewConfigService(configRepo, db)
	backupService := services.NewBackupService(db, cfg, configService, auditService)
	licenseService := services.NewLicenseService(licenseRepo, configService, auditService)
	userService := services.NewUserService(userRepo, auditService, cfg)
	authService := services.NewAuthService(userRepo, configService)
	migrationService := services.NewDataMigrationService(db, auditService, configService)

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	docService := services.NewLostDocumentService(db, docRepo, residentRepo, userRepo, auditService, configService, configRepo, exeDir)
	dashboardService := services.NewDashboardService(docRepo, userRepo, configService)
	reportService := services.NewReportService(docRepo, configService, exeDir)
	itemTemplateService := services.NewItemTemplateService(itemTemplateRepo)
	jobPositionService := services.NewJobPositionService(jobPositionRepo, auditService)
	dbTestService := services.NewDBTestService()
	updateService := services.NewUpdateService()

	authController := controllers.NewAuthController(authService, configService, version)
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
	jobPositionController := controllers.NewJobPositionController(jobPositionService)
	dbTestController := controllers.NewDBTestController(dbTestService)
	updateController := controllers.NewUpdateController(updateService, version)
	systemController := controllers.NewSystemController(db)

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.MaxMultipartMemory = 8 << 20

	funcMap := template.FuncMap{
		"ToUpper":                strings.ToUpper,
		"FormatTanggalIndonesia": utils.FormatTanggalIndonesia,
	}
	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(web.Assets, "templates/*.html", "templates/partials/*.html"))
	r.SetHTMLTemplate(templ)
	r.StaticFS("/static", web.GetStaticFS())

	r.Use(func(c *gin.Context) {
		c.Set("AppVersion", version)
		decodedChangelog, _ := utils.DecodeBase64(changelogBase64)
		c.Set("AppChangelog", decodedChangelog)
		c.Next()
	})

	r.GET("/login", authController.ShowLoginPage)
	r.POST("/api/login", middleware.LoginRateLimiter.GetLimiterMiddleware(), authController.Login)
	r.POST("/api/logout", authController.Logout)
	r.GET("/setup", configController.ShowSetupPage)
	r.POST("/api/setup", configController.SaveSetup)
	r.POST("/api/setup/restore", configController.RestoreSetup)
	r.POST("/api/db/test", dbTestController.TestConnection)
	r.GET("/api/healthz", systemController.Healthz)

	authorized := r.Group("/")
	authorized.Use(middleware.SetupMiddleware(configService))
	if userRepo != nil {
		authorized.Use(middleware.AuthMiddleware(userRepo))
	}

	authorized.GET("/", func(c *gin.Context) {
		controllers.RenderHTML(c, "dashboard.html", gin.H{"Title": "Beranda", "Config": mustGetConfig(configService)})
	})
	authorized.GET("/api/config/limits", configController.GetLimits)
	authorized.GET("/api/stats", dashboardController.GetStats)
	authorized.GET("/api/stats/monthly-issuance", dashboardController.GetMonthlyChart)
	authorized.GET("/api/stats/item-composition", dashboardController.GetItemCompositionChart)
	authorized.GET("/api/notifications/expiring-documents", dashboardController.GetExpiringDocuments)
	authorized.GET("/api/updates/check", updateController.CheckUpdate)

	authorized.GET("/documents", func(c *gin.Context) {
		controllers.RenderHTML(c, "document_list.html", gin.H{"Title": "Daftar Dokumen", "PageType": "active"})
	})
	authorized.GET("/documents/archived", func(c *gin.Context) {
		controllers.RenderHTML(c, "document_list.html", gin.H{"Title": "Arsip Dokumen", "PageType": "archived"})
	})
	authorized.GET("/documents/new", func(c *gin.Context) {
		controllers.RenderHTML(c, "document_form.html", gin.H{"Title": "Buat Surat Baru", "IsEdit": false, "DocID": 0})
	})
	authorized.GET("/documents/:id/edit", func(c *gin.Context) {
		controllers.RenderHTML(c, "document_form.html", gin.H{"Title": "Edit Surat", "IsEdit": true, "DocID": c.Param("id")})
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
		controllers.RenderHTML(c, "print_preview.html", gin.H{
			"Document":         doc,
			"Config":           conf,
			"ArchiveDays":      archiveDays,
			"ArchiveDaysWords": utils.IntToIndonesianWords(archiveDays),
		})
	})

	authorized.POST("/api/documents", docController.Create)
	authorized.GET("/api/documents", docController.FindAll)
	authorized.GET("/api/documents/:id", docController.FindByID)
	authorized.GET("/api/documents/:id/pdf", docController.GetPDF)
	authorized.PUT("/api/documents/:id", docController.Update)
	authorized.DELETE("/api/documents/:id", docController.Delete)
	authorized.GET("/api/search", docController.SearchGlobal)
	authorized.GET("/search", func(c *gin.Context) {
		controllers.RenderHTML(c, "search_results.html", gin.H{"Title": "Hasil Pencarian"})
	})

	authorized.GET("/api/item-templates/active", itemTemplateController.FindAllActive)
	authorized.GET("/profile", func(c *gin.Context) {
		controllers.RenderHTML(c, "profile.html", gin.H{"Title": "Profil Saya"})
	})
	authorized.PUT("/api/profile", userController.UpdateProfile)
	authorized.PUT("/api/profile/password", userController.ChangePassword)
	authorized.GET("/panduan", func(c *gin.Context) {
		controllers.RenderHTML(c, "panduan.html", gin.H{"Title": "Panduan", "ActiveTab": "overview"})
	})
	authorized.GET("/panduan/setup", func(c *gin.Context) {
		controllers.RenderHTML(c, "panduan_setup.html", gin.H{"Title": "Panduan", "ActiveTab": "setup"})
	})
	authorized.GET("/panduan/dokumen", func(c *gin.Context) {
		controllers.RenderHTML(c, "panduan_dokumen.html", gin.H{"Title": "Panduan", "ActiveTab": "dokumen"})
	})
	authorized.GET("/panduan/admin", func(c *gin.Context) {
		controllers.RenderHTML(c, "panduan_admin.html", gin.H{"Title": "Panduan", "ActiveTab": "admin"})
	})
	authorized.GET("/upgrade", func(c *gin.Context) {
		conf, _ := configService.GetConfig()
		controllers.RenderHTML(c, "upgrade.html", gin.H{"Title": "Upgrade ke Pro", "Config": conf})
	})
	authorized.GET("/tentang", func(c *gin.Context) {
		conf, _ := configService.GetConfig()
		controllers.RenderHTML(c, "tentang.html", gin.H{"Title": "Tentang", "Config": conf})
	})

	authorized.GET("/api/users/operators", userController.FindOperators)

	admin := authorized.Group("/")
	admin.Use(middleware.AdminAuthMiddleware())
	admin.GET("/users", func(c *gin.Context) {
		controllers.RenderHTML(c, "user_list.html", gin.H{"Title": "Manajemen Pengguna"})
	})
	admin.GET("/jabatan", func(c *gin.Context) {
		controllers.RenderHTML(c, "jabatan_list.html", gin.H{"Title": "Master Jabatan"})
	})
	admin.GET("/users/new", func(c *gin.Context) {
		controllers.RenderHTML(c, "user_form.html", gin.H{"Title": "Tambah Pengguna", "IsEdit": false, "UserID": 0})
	})
	admin.GET("/users/:id/edit", func(c *gin.Context) {
		controllers.RenderHTML(c, "user_form.html", gin.H{"Title": "Edit Pengguna", "IsEdit": true, "UserID": c.Param("id")})
	})
	admin.POST("/api/users", userController.Create)
	admin.GET("/api/users", userController.FindAll)
	admin.GET("/api/users/:id", userController.FindByID)
	admin.PUT("/api/users/:id", userController.Update)
	admin.DELETE("/api/users/:id", userController.Delete)
	admin.POST("/api/users/:id/activate", userController.Activate)
	admin.GET("/api/jabatans", jobPositionController.FindAll)
	admin.GET("/api/jabatans/active", jobPositionController.FindAllActive)
	admin.POST("/api/jabatans", jobPositionController.Create)
	admin.PUT("/api/jabatans/:id", jobPositionController.Update)
	admin.DELETE("/api/jabatans/:id", jobPositionController.Delete)
	admin.POST("/api/jabatans/:id/restore", jobPositionController.Restore)
	admin.GET("/settings", func(c *gin.Context) {
		controllers.RenderHTML(c, "settings.html", gin.H{"Title": "Pengaturan Sistem"})
	})
	admin.GET("/api/settings", settingsController.GetSettings)
	admin.PUT("/api/settings", settingsController.UpdateSettings)
	admin.POST("/api/backups", backupController.CreateBackup)
	admin.POST("/api/restore", backupController.RestoreBackup)
	admin.POST("/api/settings/migrate", configController.MigrateDatabase)
	admin.GET("/api/settings/download-cert", settingsController.DownloadCertificate)
	admin.POST("/api/settings/install-cert", settingsController.InstallCertificate)
	admin.GET("/api/audit-logs", auditController.FindAll)
	admin.GET("/api/audit-logs/export", auditController.Export)
	admin.GET("/api/metrics", systemController.Metrics)
	admin.GET("/audit-logs", func(c *gin.Context) {
		controllers.RenderHTML(c, "audit_log_list.html", gin.H{"Title": "Log Audit"})
	})
	admin.GET("/api/documents/export", docController.Export)
	admin.POST("/api/license/activate", licenseController.ActivateLicense)
	admin.GET("/api/license/hwid", licenseController.GetHardwareID)
	admin.GET("/api/license/hwid/qr", licenseController.GetHardwareIDQR)

	pro := admin.Group("/")
	pro.Use(middleware.LicenseMiddleware(licenseService))
	pro.GET("/reports/aggregate", reportController.ShowReportPage)
	pro.GET("/api/reports/aggregate/pdf", reportController.GenerateReportPDF)
	pro.GET("/templates", func(c *gin.Context) {
		controllers.RenderHTML(c, "item_template_list.html", gin.H{"Title": "Template Barang"})
	})
	pro.GET("/templates/new", func(c *gin.Context) {
		controllers.RenderHTML(c, "item_template_form.html", gin.H{"Title": "Tambah Template", "IsEdit": false, "TemplateID": 0})
	})
	pro.GET("/templates/:id/edit", func(c *gin.Context) {
		controllers.RenderHTML(c, "item_template_form.html", gin.H{"Title": "Edit Template", "IsEdit": true, "TemplateID": c.Param("id")})
	})
	pro.GET("/api/item-templates", itemTemplateController.FindAll)
	pro.GET("/api/item-templates/:id", itemTemplateController.FindByID)
	pro.POST("/api/item-templates", itemTemplateController.Create)
	pro.PUT("/api/item-templates/:id", itemTemplateController.Update)
	pro.DELETE("/api/item-templates/:id", itemTemplateController.Delete)

	srv := &http.Server{Addr: ":" + port, Handler: r}

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	isHTTPS := os.Getenv("ENABLE_HTTPS") == "true"
	certFile, keyFile := "", ""
	if isHTTPS {
		var errCert error
		certFile, keyFile, errCert = utils.EnsureCertificates()
		if errCert != nil {
			log.Printf("⚠️ ERROR CERT: %v. Fallback ke HTTP.", errCert)
			isHTTPS = false
		}
	}

	if isHTTPS {
		if isVhostSetup {
			appURL = strings.Replace(appURL, "http://", "https://", 1)
		} else {
			appURL = fmt.Sprintf("https://localhost:%s", port)
		}
	}

	go func() {
		if isHTTPS {
			log.Printf("🔒 Server berjalan di %s (HTTPS)", appURL)
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server Error: %s\n", err)
			}
		} else {
			log.Printf("🌐 Server berjalan di %s (HTTP)", appURL)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server Error: %s\n", err)
			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("✨ Membuka browser...")
		utils.OpenBrowser(appURL)
		beeep.Notify("SIMDOKPOL Berjalan", fmt.Sprintf("Akses di %s", appURL), "web/static/img/icon.png")
	}()

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				utils.OpenBrowser(appURL)
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-shutdownChan:
				return
			}
		}
	}()

	go func() {
		<-quitChan
		log.Println("🛑 Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("⚠️ Shutdown error: %v", err)
		}
		log.Println("✅ Server stopped.")
		close(shutdownChan)
		systray.Quit()
	}()
}

func onExit() {
	log.Println("👋 SIMDOKPOL Desktop ditutup.")
}

func setupEnvironment() {
	envPath := filepath.Join(utils.GetAppDataDir(), ".env")
	_ = godotenv.Overload(envPath)
}

func initializeSecrets() {
	if services.JWTSecretKeyString != "" {
		services.JWTSecretKey = []byte(services.JWTSecretKeyString)
	}

	if services.AppSecretKeyString == "" {
		services.AppSecretKeyString = os.Getenv("APP_SECRET_KEY")
	}
	if len(services.JWTSecretKey) == 0 {
		if jwtStr := os.Getenv("JWT_SECRET_KEY"); jwtStr != "" {
			services.JWTSecretKey = []byte(jwtStr)
		}
	}

	updates := make(map[string]string)

	if services.AppSecretKeyString == "" {
		log.Println("🔑 Generating new APP_SECRET_KEY...")
		b := make([]byte, 32)
		rand.Read(b)
		services.AppSecretKeyString = hex.EncodeToString(b)
		updates["APP_SECRET_KEY"] = services.AppSecretKeyString
		os.Setenv("APP_SECRET_KEY", services.AppSecretKeyString)
	}

	if len(services.JWTSecretKey) == 0 {
		log.Println("🔑 Generating new JWT_SECRET_KEY...")
		b := make([]byte, 32)
		rand.Read(b)
		jwtStr := hex.EncodeToString(b)
		services.JWTSecretKey = []byte(jwtStr)
		updates["JWT_SECRET_KEY"] = jwtStr
		os.Setenv("JWT_SECRET_KEY", jwtStr)
	}

	if len(updates) > 0 {
		if err := utils.UpdateEnvFile(updates); err != nil {
			log.Printf("⚠️ Gagal menyimpan secrets ke .env: %v", err)
		} else {
			log.Println("✅ Secrets berhasil disimpan permanen ke .env")
		}
	}
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}

	switch cfg.DBDialect {
	case "mysql":
		var tlsOption string
		switch cfg.DBSSLMode {
		case "require", "verify-full":
			tlsOption = "true"
		default:
			tlsOption = "false"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, tlsOption)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		sslMode := cfg.DBSSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, sslMode)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default:
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormConfig)
		if err == nil {
			db.Exec("PRAGMA journal_mode = WAL;")
			db.Exec("PRAGMA synchronous = NORMAL;")
			db.Exec("PRAGMA foreign_keys = ON;")
		}
	}
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	err = db.AutoMigrate(&models.User{}, &models.Resident{}, &models.LostDocument{}, &models.LostItem{}, &models.AuditLog{}, &models.Configuration{}, &models.ItemTemplate{}, &models.License{}, &models.JobPosition{})
	if err != nil {
		return nil, fmt.Errorf("migrasi gagal: %w", err)
	}
	if err := utils.NormalizeLegacyJabatanRegu(db); err != nil {
		log.Printf("WARN: gagal normalisasi jabatan regu: %v", err)
	}
	if err := utils.EnsureDefaultJobPositions(db); err != nil {
		log.Printf("WARN: gagal seed jabatan default: %v", err)
	}
	return db, nil
}

func setupLogging() {
	logPath := filepath.Join(utils.GetAppDataDir(), "logs", "simdokpol.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	fileLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	mw := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(mw)
}

func mustGetConfig(s services.ConfigService) *dto.AppConfig {
	c, _ := s.GetConfig()
	return c
}

func seedDefaultTemplates(db *gorm.DB) {
	var count int64
	db.Model(&models.ItemTemplate{}).Unscoped().Count(&count)
	if count > 0 {
		return
	}

	log.Println("🔹 Seeding templates...")

	bankOptions := []string{
		"BRI", "BCA", "Mandiri", "BNI", "BSI", "BTN", "CIMB Niaga",
		"Danamon", "Permata", "Panin", "Maybank", "Mega",
		"BTPN / Jenius", "Bank Daerah (BPD)", "Lainnya",
	}

	templates := []models.ItemTemplate{
		{
			NamaBarang: "KTP",
			Urutan:     1,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:          "NIK",
					Type:           "text",
					DataLabel:      "NIK",
					Regex:          "^[0-9]{16}$",
					RequiredLength: 16,
					IsNumeric:      true,
					Placeholder:    "16 Digit NIK",
				},
			},
		},
		{
			NamaBarang: "SIM",
			Urutan:     2,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Golongan SIM",
					Type:      "select",
					DataLabel: "Gol",
					Options:   []string{"A", "B I", "B II", "C", "D"},
				},
				{
					Label:     "Nomor SIM",
					Type:      "text",
					DataLabel: "No. SIM",
					Regex:     "^[0-9]{12,14}$",
					MinLength: 12,
					MaxLength: 14,
					IsNumeric: true,
				},
			},
		},
		{
			NamaBarang: "STNK",
			Urutan:     3,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Nomor Polisi",
					Type:      "text",
					DataLabel: "No. Pol",
				},
				{
					Label:     "Nomor Rangka",
					Type:      "text",
					DataLabel: "No. Rangka",
				},
				{
					Label:     "Nomor Mesin",
					Type:      "text",
					DataLabel: "No. Mesin",
				},
			},
		},
		{
			NamaBarang: "BPKB",
			Urutan:     4,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Nomor BPKB",
					Type:      "text",
					DataLabel: "No. BPKB",
				},
				{
					Label:     "Atas Nama",
					Type:      "text",
					DataLabel: "a.n.",
				},
			},
		},
		{
			NamaBarang: "IJAZAH",
			Urutan:     5,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Tingkat",
					Type:      "select",
					DataLabel: "Tingkat",
					Options:   []string{"SD", "SMP", "SMA", "D3", "S1", "S2"},
				},
				{
					Label:     "Nomor Ijazah",
					Type:      "text",
					DataLabel: "No. Ijazah",
				},
			},
		},
		{
			NamaBarang: "ATM",
			Urutan:     6,
			IsActive:   true,
			FieldsConfig: models.JSONFieldArray{
				{
					Label:     "Nama Bank",
					Type:      "select",
					DataLabel: "Bank",
					Options:   bankOptions,
				},
				{
					Label:     "Nomor Rekening",
					Type:      "text",
					DataLabel: "No. Rek",
				},
			},
		},
		{
			NamaBarang:   "LAINNYA",
			Urutan:       99,
			IsActive:     true,
			FieldsConfig: models.JSONFieldArray{},
		},
	}

	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "nama_barang"}},
		DoNothing: true,
	}).Create(&templates)
}
