package checker

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/example/conncheck/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func mongoTarget(target config.Target) string {
	if len(target.Addresses) > 0 {
		return strings.Join(target.Addresses, ",")
	}
	return fmt.Sprintf("%s@%s:%d/%s", safeDisplayUser(target.Username), target.Host, target.Port, target.Database)
}

func checkMongoDB(ctx context.Context, target config.Target) error {
	uri := mongoURI(target)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(target.Timeout).SetServerSelectionTimeout(target.Timeout))
	if err != nil {
		return fmt.Errorf("create mongodb client: %w", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	if err := client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return fmt.Errorf("ping mongodb: %w", err)
	}
	return nil
}

func mongoURI(target config.Target) string {
	if strings.HasPrefix(target.Database, "mongodb://") || strings.HasPrefix(target.Database, "mongodb+srv://") {
		return target.Database
	}
	credentials := ""
	if target.Username != "" || target.Password != "" {
		credentials = url.UserPassword(target.Username, target.Password).String() + "@"
	}
	hosts := target.Addresses
	if len(hosts) == 0 {
		hosts = []string{fmt.Sprintf("%s:%d", target.Host, target.Port)}
	}
	authDatabase := target.Attributes["auth_database"]
	if authDatabase == "" {
		authDatabase = "admin"
	}
	query := url.Values{}
	if target.TLS {
		query.Set("tls", "true")
	}
	if len(query) > 0 {
		return fmt.Sprintf("mongodb://%s%s/%s?%s", credentials, strings.Join(hosts, ","), target.Database, query.Encode())
	}
	return fmt.Sprintf("mongodb://%s%s/%s", credentials, strings.Join(hosts, ","), target.Database)
}
