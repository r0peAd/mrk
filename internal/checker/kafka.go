package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/example/conncheck/internal/config"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func kafkaTarget(target config.Target) string {
	return strings.Join(target.Brokers, ",")
}

func checkKafka(ctx context.Context, target config.Target) error {
	dialer := &kafkago.Dialer{
		Timeout:   target.Timeout,
		DualStack: true,
	}
	if target.TLS {
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	if target.SASL.Enabled {
		mechanism, err := kafkaSASL(target)
		if err != nil {
			return err
		}
		dialer.SASLMechanism = mechanism
	}

	conn, err := dialer.DialContext(ctx, "tcp", target.Brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker %s: %w", target.Brokers[0], err)
	}
	defer conn.Close()

	brokers, err := conn.Brokers()
	if err != nil {
		return fmt.Errorf("list kafka brokers: %w", err)
	}
	if len(brokers) == 0 {
		return fmt.Errorf("kafka returned no broker metadata")
	}
	return nil
}

func kafkaSASL(target config.Target) (sasl.Mechanism, error) {
	switch strings.ToLower(target.SASL.Mechanism) {
	case "", "plain":
		return plain.Mechanism{
			Username: target.SASL.Username,
			Password: target.SASL.Password,
		}, nil
	case "scram-sha-256":
		return scram.Mechanism(scram.SHA256, target.SASL.Username, target.SASL.Password)
	case "scram-sha-512":
		return scram.Mechanism(scram.SHA512, target.SASL.Username, target.SASL.Password)
	default:
		return nil, fmt.Errorf("unsupported kafka sasl mechanism %q", target.SASL.Mechanism)
	}
}
