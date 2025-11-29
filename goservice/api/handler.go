package api

import (
	"encoding/json"
	"net/http"
)

// EchoHandler 返回请求数据
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := map[string]interface{}{
		"message": "Hello from Go",
		"data":    payload,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
