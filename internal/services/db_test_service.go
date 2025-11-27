package services

import (
	"context"
	"fmt"
	"net"
	"simdokpol/internal/dto"
	"time"

	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBTestService interface {
	TestConnection(req dto.DBTestRequest) error
}

type dbTestService struct{}

func NewDBTestService() DBTestService {
	return &dbTestService{}
}

func isSafeHost(host string) error {
	// Logic security (IP check) - Disederhanakan untuk mengizinkan semua IP
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("gagal resolve host: %v", err)
	}
	for _, ip := range ips {
		if ip.IsLinkLocalMulticast() {
			return fmt.Errorf("koneksi multicast tidak diizinkan")
		}
	}
	return nil
}

func (s *dbTestService) TestConnection(req dto.DBTestRequest) error {
	if req.DBDialect != "sqlite" {
		if err := isSafeHost(req.DBHost); err != nil {
			return err
		}
	}

	var dsn string
	var gormDialector gorm.Dialector

	switch req.DBDialect {
	case "mysql":
		// ✅ PERBAIKAN: Gunakan switch statement untuk SSL mode
		var tlsOption string
		switch req.DBSSLMode {
		case "require":
			tlsOption = "skip-verify"
		case "verify-full":
			tlsOption = "true"
		default:
			tlsOption = "false"
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
			req.DBUser, req.DBPass, req.DBHost, req.DBPort, req.DBName, tlsOption)
		
		gormDialector = gormmysql.Open(dsn)

	case "postgres":
		// ✅ PERBAIKAN: Juga gunakan switch untuk konsistensi
		var sslMode string
		switch req.DBSSLMode {
		case "require", "verify-ca", "verify-full":
			sslMode = req.DBSSLMode
		default:
			sslMode = "disable"
		}
		
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort, sslMode)
		gormDialector = gormpostgres.Open(dsn)
		
	case "sqlite":
		dsn = "simdokpol.db?_foreign_keys=on"
		gormDialector = gormsqlite.Open(dsn)
		
	default:
		return fmt.Errorf("dialek database '%s' tidak didukung", req.DBDialect)
	}

	db, err := gorm.Open(gormDialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("gagal GORM Open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("gagal instance sql.DB: %w", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("gagal Ping DB: %w", err)
	}
	
	return nil
}