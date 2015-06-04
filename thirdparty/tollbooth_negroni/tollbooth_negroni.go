package tollbooth_negroni

import (
	"github.com/codegangsta/negroni"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/config"
	"net/http"
)

func LimitHandler(limiter *config.Limiter) negroni.HandlerFunc {
	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		httpError := tollbooth.LimitByRequest(limiter, r)
		if httpError != nil {
			w.Header().Add("Content-Type", limiter.MessageContentType)
			w.Write([]byte(httpError.Message))
			w.WriteHeader(httpError.StatusCode)
			return

		} else {
			next(w, r)
		}

	})
}
