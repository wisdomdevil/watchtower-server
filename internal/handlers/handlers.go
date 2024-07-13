package handlers

import (
	"net/http"
	"sync"
	"watchtower-server/internal/inactivityTimer"
	"watchtower-server/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type MetaVM struct {
	OsProject string `json:"os_project"`
}
type InfoVM struct {
	Hostname  string `json:"hostname"`
	Name      string `json:"name"`
	ProjectId string `json:"project_id"`
	Uuid      string `json:"uuid"`
	MetaVM    `json:"meta"`
}

type muInfoVMMap struct {
	infoVMMap map[string]InfoVM
	mu        sync.RWMutex
}

var (
	//Инициализируем мапу содержающую структуру InfoVM для каждого клиента.
	MuInfoVM muInfoVMMap
)

// Функция создания мапы при первом вызове хендлера
func InitMuInfoVM() {
	MuInfoVM.infoVMMap = make(map[string]InfoVM)
}

func (m *muInfoVMMap) Add(name string, info InfoVM) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoVMMap[name] = info
}

func (m *muInfoVMMap) Read(name string) InfoVM {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.infoVMMap[name]
}

func (m *muInfoVMMap) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.infoVMMap, name)
}

func ParseJSONHandler(c *gin.Context) {
	var info InfoVM

	if c.Request.Method != "POST" {
		c.String(http.StatusOK, "Invalid method")
	} else {
		// Парсинг JSON в структуру Person
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Записываем в мапу
		MuInfoVM.Add(info.Name, info)
		//При входящем POST запросе, создаем новый key-value в мапе или обновляем существующий

		inactivityTimer.MuLastActivity.UpdateTime(info.Name)

		metrics.InfoMetric.With(prometheus.Labels{
			"hostname":   info.Hostname,
			"os_project": info.OsProject,
			"project_id": info.ProjectId,
			"name":       info.Name,
			"uuid":       info.Uuid,
		}).Set(1)

	}
}
