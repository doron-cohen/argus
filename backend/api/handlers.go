package api

import (
	"encoding/json"
	"net/http"
	"time"

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
		component := Component{
			Name: c.Name,
		}

		// Set ID if available (use ComponentID from storage)
		if c.ComponentID != "" {
			id := c.ComponentID
			component.Id = &id
		}

		// Set description if available
		if c.Description != "" {
			description := c.Description
			component.Description = &description
		}

		// Set owners if available
		if len(c.Maintainers) > 0 || c.Team != "" {
			owners := &Owners{}

			if len(c.Maintainers) > 0 {
				maintainers := []string(c.Maintainers)
				owners.Maintainers = &maintainers
			}

			if c.Team != "" {
				team := c.Team
				owners.Team = &team
			}

			component.Owners = owners
		}

		apiComponents = append(apiComponents, component)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiComponents)
}

func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := Health{
		Status:    Healthy,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}
