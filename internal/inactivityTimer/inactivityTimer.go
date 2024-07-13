package inactivityTimer

import (
	"sync"
	"time"
)

type muLastActivityMap struct {
	LastActivity map[string]time.Time
	mu           sync.RWMutex
}

var (
	MuLastActivity muLastActivityMap
)

func InitMuLastActivity() {
	MuLastActivity.LastActivity = make(map[string]time.Time)
}

func (m *muLastActivityMap) UpdateTime(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastActivity[name] = time.Now()
}

func (m *muLastActivityMap) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.LastActivity, name)
}

func (m *muLastActivityMap) Read(name string) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.LastActivity[name]
}

func (m *muLastActivityMap) Range(f func(key string, value time.Time)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.LastActivity {
		f(k, v)
	}
}

// Проверяет прошло ли время таймаута начиная с последнего сообщения, если прошло - true, если нет - false
func CheckInactivity(name string, timeout time.Duration) (result bool) {
	return time.Since(MuLastActivity.Read(name)) > timeout
}
