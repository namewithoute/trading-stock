package initialize

import (
	"context"
	"fmt"
	"time"
	"trading-stock/internal/config"
	"trading-stock/internal/global"
	"trading-stock/pkg/utils"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// InitKafka: Chỉ chịu trách nhiệm kết nối và tạo Global Writer
func InitKafka(ctx context.Context, cfg config.KafkaConfig) error {
	// 1. Fail Fast: Ping thử tới Brokers
	err := utils.DoWithRetry(ctx, global.Logger, "Kafka Connection", 2*time.Second, func() error {
		for _, broker := range cfg.Brokers {
			conn, err := kafka.Dial("tcp", broker)
			if err == nil {
				conn.Close()
				return nil
			}
		}
		return fmt.Errorf("failed to connect to any kafka broker")
	})

	if err != nil {
		return err
	}

	// 2. Init Producer (Writer) với cấu hình tối ưu
	// Topic "orders" có thể lấy từ config thay vì hardcode
	global.Kafka = &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    "orders",            // Nên đưa vào Config: cfg.Topic
		Balancer: &kafka.LeastBytes{}, // Phân phối load thông minh dựa trên dung lượng

		// --- PERFORMANCE TUNING ---
		// Gom 100 tin nhắn hoặc chờ 10ms rồi mới gửi 1 lần -> Tăng throughput cực mạnh
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,

		// Nén data giúp giảm tải Network (mạnh hơn Gzip, ít tốn CPU hơn)
		Compression: kafka.Snappy,

		// Timeout cho việc write (tránh treo app nếu Kafka chết)
		WriteTimeout: 10 * time.Second,

		// Retry policy: Thử lại 3 lần nếu lỗi mạng nhẹ, mỗi lần cách nhau 2-10s
		MaxAttempts: 3,
	}

	return nil
}

// Hàm đóng Kafka chuẩn (đặt ở bootstrap/startup.go)
func CloseKafka() {
	if global.Kafka != nil {
		// Close() sẽ tự động flush các tin nhắn còn tồn đọng trong buffer
		if err := global.Kafka.Close(); err != nil {
			global.Logger.Error("Failed to close Kafka writer", zap.Error(err))
		} else {
			global.Logger.Info("Kafka writer closed successfully")
		}
	}
}
