package api

import (
	"encoding/json"
	"net/http"
)

// APIServer implements ServerInterface
// (Component type is generated in api.gen.go)
type APIServer struct{}

func NewAPIServer() ServerInterface {
	return &APIServer{}
}

func (s *APIServer) GetComponents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]Component{})
}
