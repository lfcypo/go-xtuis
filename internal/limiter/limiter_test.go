package limiter

import (
	"sync"
	"testing"
	"time"
)

type mockClock struct {
	currentTime time.Time
	mu          sync.RWMutex
}

func newMockClock() *mockClock {
	return &mockClock{
		currentTime: time.Now(),
	}
}

func (m *mockClock) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTime
}

func (m *mockClock) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTime = m.currentTime.Add(d)
}

func (m *mockClock) Set(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTime = t
}

type testLimiter struct {
	maxTimes  int
	window    time.Duration
	count     int
	windowEnd time.Time
	mu        sync.Mutex
	nowFunc   func() time.Time
}

func (l *testLimiter) Limit() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
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

func newTestLimiter(maxTimes int, window time.Duration, nowFunc func() time.Time) *testLimiter {
	now := nowFunc()
	return &testLimiter{
		maxTimes:  maxTimes,
		window:    window,
		count:     0,
		windowEnd: now.Add(window),
		nowFunc:   nowFunc,
	}
}

func TestDayLimiter(t *testing.T) {
	testMaxTimes := 3
	testWindow := 24 * time.Hour
	mockClock := newMockClock()
	nowFunc := mockClock.Now
	limiter := newTestLimiter(testMaxTimes, testWindow, nowFunc)
	mockClock.Set(time.Now())

	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(25 * time.Hour)
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}
}

func TestMinuteLimiter(t *testing.T) {
	testMaxTimes := 2
	testWindow := time.Minute
	mockClock := newMockClock()
	nowFunc := mockClock.Now
	limiter := newTestLimiter(testMaxTimes, testWindow, nowFunc)
	mockClock.Set(time.Now())

	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(30 * time.Second)
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(31 * time.Second)
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(61 * time.Second)
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
}

func TestLimiterConcurrent(t *testing.T) {
	testMaxTimes := 100
	testWindow := time.Minute
	mockClock := newMockClock()
	nowFunc := mockClock.Now
	limiter := newTestLimiter(testMaxTimes, testWindow, nowFunc)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Limit() {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != testMaxTimes {
		t.Errorf("期望%d, 得到%d", testMaxTimes, successCount)
	}
}

func TestLimiterZeroLimit(t *testing.T) {
	mockClock := newMockClock()
	limiter := newTestLimiter(0, time.Minute, mockClock.Now)

	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(2 * time.Minute)
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}
}

func TestLimiterLargeWindow(t *testing.T) {
	mockClock := newMockClock()
	limiter := newTestLimiter(1, 365*24*time.Hour, mockClock.Now)

	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}

	mockClock.Advance(364 * 24 * time.Hour)
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(2 * 24 * time.Hour)
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}
}

func TestLimiterTimeBackward(t *testing.T) {
	mockClock := newMockClock()
	initialTime := mockClock.Now()
	limiter := newTestLimiter(2, time.Hour, mockClock.Now)

	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}

	mockClock.Set(initialTime.Add(-30 * time.Minute))
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}

	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}
}

func TestLimiterResetBehavior(t *testing.T) {
	mockClock := newMockClock()
	limiter := newTestLimiter(2, time.Hour, mockClock.Now)

	limiter.Limit()
	limiter.Limit()

	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(59 * time.Minute)
	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}

	mockClock.Advance(2 * time.Minute)
	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}

	if !limiter.Limit() {
		t.Errorf("期望true, 得到false")
	}

	if limiter.Limit() {
		t.Errorf("期望false, 得到true")
	}
}
