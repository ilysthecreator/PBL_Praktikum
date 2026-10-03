package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"UTS/helper"
)

type FailedLoginTracker struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

var GlobalLoginTracker = &FailedLoginTracker{
	attempts: make(map[string][]time.Time),
}

func (t *FailedLoginTracker) IsBlocked(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// Bersihkan riwayat yang lebih dari 1 menit
	valid := []time.Time{}
	for _, ts := range t.attempts[ip] {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	t.attempts[ip] = valid

	return len(valid) >= 5
}

func (t *FailedLoginTracker) RecordFailure(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	valid := []time.Time{}
	for _, ts := range t.attempts[ip] {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	valid = append(valid, now)
	t.attempts[ip] = valid
}

func (t *FailedLoginTracker) RecordSuccess(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, ip)
}

func LoginRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if GlobalLoginTracker.IsBlocked(ip) {
			c.Set("Retry-After", "60")
			return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan gagal login, coba lagi setelah 1 menit")
		}
		return c.Next()
	}
}
