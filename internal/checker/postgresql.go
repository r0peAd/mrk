package checker

import (
	"context"
	"fmt"
	"net/url"

	"github.com/example/conncheck/internal/config"
	"github.com/jackc/pgx/v5"
)

func postgresTarget(target config.Target) string {
	return fmt.Sprintf("%s@%s:%d/%s", safeDisplayUser(target.Username), target.Host, target.Port, target.Database)
}

func checkPostgreSQL(ctx context.Context, target config.Target) error {
	dsn := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", target.Host, target.Port),
		Path:   target.Database,
		User:   url.UserPassword(target.Username, target.Password),
	}
	query := url.Values{}
	query.Set("connect_timeout", fmt.Sprintf("%d", max(int(target.Timeout.Seconds()), 1)))
	if target.TLS {
		query.Set("sslmode", "verify-full")
	} else {
		query.Set("sslmode", "disable")
	}
	dsn.RawQuery = query.Encode()

	conn, err := pgx.Connect(ctx, dsn.String())
	if err != nil {
		return fmt.Errorf("connect postgresql: %w", err)
	}
	defer conn.Close(context.Background())

	var version string
	if err := conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		return fmt.Errorf("query postgresql version: %w", err)
	}
	if version == "" {
		return fmt.Errorf("postgresql returned empty server version")
	}
	return nil
}
