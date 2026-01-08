package tracker

import (
	"encoding/json"
	"net/http"
)

func (s *Server) respond(rw http.ResponseWriter, status int, v any) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	encoder := json.NewEncoder(rw)
	if err := encoder.Encode(v); err != nil {
		panic(err)
	}
}
