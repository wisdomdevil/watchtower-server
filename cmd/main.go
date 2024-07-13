package main

import (
	"flag"
	"os"
	"time"
	"watchtower-server/internal/handlers"
	"watchtower-server/internal/inactivityTimer"
	"watchtower-server/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Флаги для запуска
	port := flag.String("port", "8080", "Port for running webserver")
	timeout := flag.Duration("timeout", 2*time.Minute, "Время после которого виртуальная машина считается недоступной. def: 2m")
	retentionTimeout := flag.Duration("retention", 10800*time.Minute, "Время после которого виртуальная машина считается не актуальной и удаляется из списка метрик")
	address := flag.String("address", "0.0.0.0", "адрес для старта вебсервера")
	help := flag.Bool("help", false, "Отображает доступные флаги")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	handlers.InitMuInfoVM()
	inactivityTimer.InitMuLastActivity()
	r := gin.Default()
	//Метрики промитиуса
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/push", handlers.ParseJSONHandler)
	r.POST("/push", handlers.ParseJSONHandler)
	/*Запуск тикера, тикер раз в минуту запускает метод range, который переопределяет поведение цикла for,
	для добавления mutex  в операции Set 0 (в случае если клиент не отправил сообщения в течении последних timeout минут)
	и операции Delete, в случае если клиент, не отправлял сообщений в течении retentionTimeout минут
	*/
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				inactivityTimer.MuLastActivity.Range(func(name string, value time.Time) {
					readData := handlers.MuInfoVM.Read(name)
					if inactivityTimer.CheckInactivity(name, *timeout) {
						metrics.InfoMetric.With(prometheus.Labels{
							"hostname":   readData.Hostname,
							"os_project": readData.OsProject,
							"project_id": readData.ProjectId,
							"name":       readData.Name,
							"uuid":       readData.Uuid,
						}).Set(0)
					}
					if inactivityTimer.CheckInactivity(name, *retentionTimeout) {
						metrics.InfoMetric.Delete(prometheus.Labels{
							"hostname":   readData.Hostname,
							"os_project": readData.OsProject,
							"project_id": readData.ProjectId,
							"name":       readData.Name,
							"uuid":       readData.Uuid,
						})
						handlers.MuInfoVM.Delete(name)
						inactivityTimer.MuLastActivity.Delete(name)
					}

				})
			}
		}
	}()

	r.Run(*address + ":" + *port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:80)

}
