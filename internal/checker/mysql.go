package checker

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/example/conncheck/internal/config"
	"github.com/go-sql-driver/mysql"
)

func mysqlTarget(target config.Target) string {
	return fmt.Sprintf("%s@%s:%d/%s", safeDisplayUser(target.Username), target.Host, target.Port, target.Database)
}

func checkMySQL(ctx context.Context, target config.Target) error {
	cfg := mysql.NewConfig()
	cfg.User = target.Username
	cfg.Passwd = target.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", target.Host, target.Port)
	cfg.DBName = target.Database
	cfg.Timeout = target.Timeout
	cfg.ReadTimeout = target.Timeout
	cfg.ParseTime = true
	if target.TLS {
		cfg.TLSConfig = "true"
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}
	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return fmt.Errorf("query mysql version: %w", err)
	}
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("mysql returned empty server version")
	}
	return nil
}
