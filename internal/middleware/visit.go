package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"strings"

	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// VisitTracker 访问记录中间件：记录每次前台请求"是谁、何时、访问了什么"。
// 访客身份两级识别：
//  1. 前端 localStorage 生成的 UUID，通过 X-Visitor-ID 请求头带上（首选）
//  2. 没带头的，降级用 sha256(IP + User-Agent)
//
// 异步落库：go func 里写 MySQL，不阻塞响应。博客流量下安全；
// 量大了换 buffered channel + 单 worker（对应 Java 的线程池 + 队列）。
func VisitTracker(repo *repository.VisitLogRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 后台接口与部署探活不算访客
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/admin") || path == "/api/v1/healthz" {
			c.Next()
			return
		}

		ua := c.Request.UserAgent()
		ip := normalizeIP(c.ClientIP())

		// 1. 确定访客身份
		key := strings.TrimSpace(c.GetHeader("X-Visitor-ID"))
		if len(key) > 64 {
			key = key[:64] // 防恶意超长头
		}
		if key == "" {
			sum := sha256.Sum256([]byte(ip + "|" + ua))
			key = hex.EncodeToString(sum[:])[:64]
		}
		nickname := strings.TrimSpace(c.GetHeader("X-Visitor-Name"))
		if len(nickname) > 64 {
			nickname = nickname[:64]
		}
		device := strings.TrimSpace(c.GetHeader("X-Visitor-Device"))
		if len(device) > 128 {
			device = device[:128]
		}

		// 2. 先取出需要的数据再起 goroutine！
		//    gin.Context 在请求结束后会被回收复用，异步代码里绝不能再碰 c。
		entry := model.VisitLog{
			VisitorKey: key,
			Nickname:   nickname,
			Device:     device,
			Path:       path,
			IP:         ip,
			UserAgent:  truncate(ua, 512),
			Referer:    truncate(c.Request.Header.Get("Referer"), 512),
		}

		// 3. 异步写库，不阻塞响应
		go func() {
			defer func() {
				if r := recover(); r != nil { // goroutine 里的 panic 不会自动冒泡，必须自己兜住
					log.Printf("记录访问日志失败(panic): %v", r)
				}
			}()
			if err := repo.Create(&entry); err != nil {
				log.Printf("记录访问日志失败: %v", err)
			}
		}()

		// 4. 放行，继续后面的处理
		c.Next()
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// normalizeIP 把 IPv6 回环/映射地址规整成可读的 IPv4（::1 → 127.0.0.1、::ffff:1.2.3.4 → 1.2.3.4），
// 避免后台把 ::1 这类地址当成"乱码"展示。无法解析的原始串原样返回。
func normalizeIP(s string) string {
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return s
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	if ip.IsLoopback() {
		return "127.0.0.1"
	}
	return s
}
