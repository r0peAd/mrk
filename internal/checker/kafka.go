package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"
	"net"

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
		dialer.TLS =  &tls.Config{InsecureSkipVerify: true}
		//&tls.Config{MinVersion: tls.VersionTLS12}
	}
	if target.SASL.Enabled {
		//mechanism, err := kafkaSASL(target)
		//if err != nil {
		//	return err
		//}
		//dialer.SASLMechanism = mechanism
	}


	conn, err := dialer.DialContext(ctx, "tcp", target.Brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker %s: %w", target.Brokers[0], err)
	}
	defer conn.Close()

	// 设置短超时，尝试读 1 字节
// - 如果返回 timeout 错误 → 连接大概率健康（对端没关）
// - 如果返回 EOF → 对端已关闭
// - 如果读到数据 → 正常
conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
buf := make([]byte, 1)
n, err := conn.Read(buf)
if err != nil {
    if ne, ok := err.(net.Error); ok && ne.Timeout() {
        // 超时 = 连接仍然存活（Kafka 不会主动推送数据）
    } else {
        // EOF 或其他错误 = 连接已断开
        //log.Fatalf("conn dead: %v", err)
        return fmt.Errorf("dial kafka broker %s: %w", target.Brokers[0], err)

    }
} else {
    _ = n // 意外收到数据
}
conn.SetReadDeadline(time.Time{}) // 清除 deadline

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
