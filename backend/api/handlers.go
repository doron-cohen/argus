package api

import (
	"encoding/json"
	"net/http"

	"github.com/doron-cohen/argus/backend/internal/storage"
)

type APIServer struct {
	Repo *storage.Repository
}

func NewAPIServer(repo *storage.Repository) ServerInterface {
	return &APIServer{Repo: repo}
}

func (s *APIServer) GetComponents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	components, err := s.Repo.GetComponents(ctx)
	if err != nil {
		http.Error(w, "failed to fetch components", http.StatusInternalServerError)
		return
	}

	var apiComponents []Component
	for _, c := range components {
		id := c.ID.String()
		name := c.Name
		apiComponents = append(apiComponents, Component{
			Id:   &id,
			Name: &name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiComponents)
}
