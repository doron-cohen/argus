package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

func (s *APIServer) GetComponentReports(w http.ResponseWriter, r *http.Request, componentId string, params GetComponentReportsParams) {
	ctx := r.Context()

	var reports []storage.CheckReport
	var err error

	if params.LatestOnly != nil && *params.LatestOnly {
		reports, err = s.Repo.GetLatestCheckReportsForComponent(ctx, componentId)
	} else {
		reports, err = s.Repo.GetCheckReportsForComponent(ctx, componentId)
	}

	if err != nil {
		if err == storage.ErrComponentNotFound {
			s.writeNotFoundError(w)
			return
		}
		http.Error(w, "failed to fetch component reports", http.StatusInternalServerError)
		return
	}

	// Convert storage reports to API reports
	var apiReports []CheckReport
	for _, report := range reports {
		apiReport := s.convertToAPICheckReport(&report)
		apiReports = append(apiReports, apiReport)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiReports); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *APIServer) convertToAPICheckReport(report *storage.CheckReport) CheckReport {
	// Convert UUID to openapi_types.UUID
	reportID := openapi_types.UUID(report.ID)

	apiReport := CheckReport{
		Id:        reportID,
		Status:    CheckReportStatus(report.Status),
		Timestamp: report.Timestamp,
		CreatedAt: &report.CreatedAt,
	}

	// Convert check information
	checkID := openapi_types.UUID(report.Check.ID)
	apiReport.Check = Check{
		Id:          checkID,
		Slug:        report.Check.Slug,
		Name:        report.Check.Name,
		Description: &report.Check.Description,
		CreatedAt:   &report.Check.CreatedAt,
	}

	// Convert optional JSONB fields
	if report.Details != nil {
		details := map[string]interface{}(report.Details)
		apiReport.Details = &details
	}
	if report.Metadata != nil {
		metadata := map[string]interface{}(report.Metadata)
		apiReport.Metadata = &metadata
	}

	return apiReport
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
