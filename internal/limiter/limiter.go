package limiter

import (
	"sync"
	"time"
)

// Limiter 限流器
type Limiter interface {
	Limit() bool
}

// Factory 创建限流器
type Factory func() Limiter

type limiterImpl struct {
	maxTimes  int
	window    time.Duration
	count     int
	windowEnd time.Time
	mu        sync.Mutex
}

func (l *limiterImpl) Limit() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if now.After(l.windowEnd) {
		l.count = 0
		l.windowEnd = now.Add(l.window)
	}

	if l.count >= l.maxTimes {
		return false
	}

	l.count++
	return true
}

func newLimiter(maxTimes int, window time.Duration) Limiter {
	return &limiterImpl{
		maxTimes:  maxTimes,
		window:    window,
		count:     0,
		windowEnd: time.Now().Add(window),
	}
}

var (
	// DayLimiter 每日限流器
	DayLimiter = newLimiter(MaxTimesPerDays, 24*time.Hour)
	// MinuteLimiter 每分钟限流器
	MinuteLimiter = newLimiter(MaxTimesPerMinutes, time.Minute)
)
