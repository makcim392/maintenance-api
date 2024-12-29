package handlers

import (
	"net/http"
	"time"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Version 3 - " + time.Now().String()))
}
