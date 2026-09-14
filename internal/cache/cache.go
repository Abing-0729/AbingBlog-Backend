package cache

import (
	"context"
	"log"
	"time"

	"abingblog-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// Init 连接 Redis 并 Ping 探活。返回的 *redis.Client 是并发安全的，
// 全局单例即可，注入到需要缓存的 service 层。
//
// 注意：这里只负责"连上 Redis"这一件事。具体缓存怎么用——
// 读时先查缓存后查库（Cache Aside）、写时如何让缓存失效、
// key 怎么设计、TTL 定多久、穿透/雪崩怎么防——都是 service 层的缓存策略，
// 不在客户端封装的职责范围内（那部分是面试重点，留给业务代码去写）。
func Init(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 启动即探活：连不上直接 Fatal，避免带病启动后请求时才发现 Redis 挂了
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}
	log.Printf("Redis connected at %s (db=%d)", cfg.Redis.Addr(), cfg.Redis.DB)

	return client
}
