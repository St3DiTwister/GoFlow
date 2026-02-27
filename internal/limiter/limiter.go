package limiter

import (
	"context"
	"golang.org/x/time/rate"
	"sync"
	"sync/atomic"
	"time"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen int64
}

type IPRateLimiter struct {
	clients map[string]*client
	mu      sync.RWMutex
	r       float64
	b       int
}

func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.RLock()
	cl, ok := l.clients[ip]
	if ok {
		atomic.StoreInt64(&cl.lastSeen, time.Now().Unix())
		l.mu.RUnlock()
		return cl.limiter
	}
	l.mu.RUnlock()
	l.mu.Lock()
	cl, ok = l.clients[ip]
	if ok {
		l.mu.Unlock()
		return cl.limiter
	}
	limiter := rate.NewLimiter(rate.Limit(l.r), l.b)
	defer l.mu.Unlock()
	l.clients[ip] = &client{limiter: limiter, lastSeen: time.Now().Unix()}
	return limiter
}

func NewIPRateLimiter(ctx context.Context, r float64, b int) *IPRateLimiter {
	rateLimiter := &IPRateLimiter{
		clients: make(map[string]*client),
		r:       r,
		b:       b,
		mu:      sync.RWMutex{},
	}

	go func() {
		tick := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-tick.C:
				keysToDelete := make([]string, 0)
				rateLimiter.mu.RLock()
				for k, v := range rateLimiter.clients {
					if time.Since(time.Unix(v.lastSeen, 0)) > 10*time.Minute {
						keysToDelete = append(keysToDelete, k)
					}
				}
				rateLimiter.mu.RUnlock()

				rateLimiter.mu.Lock()
				for _, k := range keysToDelete {
					data := rateLimiter.clients[k]
					if time.Since(time.Unix(data.lastSeen, 0)) > 10*time.Minute {
						delete(rateLimiter.clients, k)
					}
				}
				rateLimiter.mu.Unlock()
			case <-ctx.Done():
				tick.Stop()
				return
			}
		}
	}()

	return rateLimiter
}
