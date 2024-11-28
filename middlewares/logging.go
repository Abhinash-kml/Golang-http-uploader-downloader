package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Chain(final http.Handler, handlers ...Middleware) http.Handler {
	for i := range handlers {
		final = handlers[i](final)
	}

	return final
}

func Logging(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time := time.Now()

		err := os.MkdirAll("logs", os.ModePerm)
		if err != nil {
			log.Fatal("Could not create logs directory")
			return
		}

		logsFilehandler, err := os.OpenFile("./logs/resquestlog.txt", os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			log.Fatal("Could not create request log file")
			return
		}

		text := fmt.Sprintf("IP: %s - Time: %s\n", r.RemoteAddr, time.String())
		fmt.Fprint(logsFilehandler, text)

		next.ServeHTTP(w, r)
	})
}
