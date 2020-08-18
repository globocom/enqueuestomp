package enqueuestomp

import (
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
)

func StartDashboard(addr string) {
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()

	logger.Printf("start dashboard http://%s", addr)
	go func() {
		logger.Fatal(http.ListenAndServe(addr, hystrixStreamHandler))
	}()
}
