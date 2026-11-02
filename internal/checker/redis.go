package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/example/conncheck/internal/config"
	"github.com/redis/go-redis/v9"
)

func redisTarget(target config.Target) string {
	return fmt.Sprintf("%s:%d", target.Host, target.Port)
}

func checkRedis(ctx context.Context, target config.Target) error {
	options := &redis.Options{
		Addr:            fmt.Sprintf("%s:%d", target.Host, target.Port),
		Username:        target.Username,
		Password:        target.Password,
		DB:              redisDatabase(target),
		ConnMaxLifetime: target.Timeout,
		DialTimeout:     target.Timeout,
		ReadTimeout:     target.Timeout,
		WriteTimeout:    target.Timeout,
		MaxRetries:      -1,
	}
	if target.TLS {
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	if target.SASL.Enabled {
		options.CredentialsProvider = func() (string, string) {
			return target.SASL.Username, target.SASL.Password
		}
	}

	client := redis.NewClient(options)
	defer client.Close()

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	if !strings.EqualFold(pong, "PONG") {
		return fmt.Errorf("unexpected redis response: %q", pong)
	}
	return nil
}

func redisDatabase(target config.Target) int {
	if target.Database == "" {
		return 0
	}
	var database int
	if _, err := fmt.Sscanf(target.Database, "%d", &database); err != nil {
		return 0
	}
	return database
}
