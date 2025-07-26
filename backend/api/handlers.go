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
	if err := json.NewEncoder(w).Encode(apiComponents); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *APIServer) GetComponentById(w http.ResponseWriter, r *http.Request, componentId string) {
	ctx := r.Context()
	component, err := s.Repo.GetComponentByID(ctx, componentId)
	if err != nil {
		if err == storage.ErrComponentNotFound {
			s.writeNotFoundError(w)
			return
		}
		http.Error(w, "failed to fetch component", http.StatusInternalServerError)
		return
	}

	// Convert storage component to API component
	apiComponent := s.convertToAPIComponent(component)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiComponent); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// convertToAPIComponent converts a storage component to an API component
func (s *APIServer) convertToAPIComponent(component *storage.Component) Component {
	apiComponent := Component{
		Name: component.Name,
	}

	// Set ID if available (use ComponentID from storage)
	if component.ComponentID != "" {
		id := component.ComponentID
		apiComponent.Id = &id
	}

	// Set description if available
	if component.Description != "" {
		description := component.Description
		apiComponent.Description = &description
	}

	// Set owners if available
	if len(component.Maintainers) > 0 || component.Team != "" {
		owners := &Owners{}

		if len(component.Maintainers) > 0 {
			maintainers := []string(component.Maintainers)
			owners.Maintainers = &maintainers
		}

		if component.Team != "" {
			team := component.Team
			owners.Team = &team
		}

		apiComponent.Owners = owners
	}

	return apiComponent
}

// writeNotFoundError writes a not found error response
func (s *APIServer) writeNotFoundError(w http.ResponseWriter) {
	code := "NOT_FOUND"
	errorResponse := Error{
		Error: "Component not found",
		Code:  &code,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
		return
	}
}

func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := Health{
		Status:    Healthy,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
