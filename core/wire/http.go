package wire

import (
	"encoding/json"
	"net/http"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func durationSeconds(value int) time.Duration {
	if value <= 0 {
		value = 120
	}
	return time.Duration(value) * time.Second
}
