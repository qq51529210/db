package web

import (
	"github.com/qq51529210/log"
	"github.com/qq51529210/web/router"
	"net/http"
)

func ListenAndServe(addr string) {
	r := router.NewRouter()
	log.CheckError(http.ListenAndServe(addr, r))
}
