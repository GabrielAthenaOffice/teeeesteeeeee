package handler

import "net/http"

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeError(
		w,
		http.StatusNotFound,
		CodeNotFound,
		"rota não encontrada: "+r.Method+" "+r.URL.Path,
	)
}
