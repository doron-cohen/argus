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

func (s *APIServer) GetComponentReports(w http.ResponseWriter, r *http.Request, componentId string, params GetComponentReportsParams) {
	ctx := r.Context()

	// Get reports for the component
	reports, err := s.Repo.GetCheckReportsForComponent(ctx, componentId)
	if err != nil {
		if err == storage.ErrComponentNotFound {
			s.writeNotFoundError(w)
			return
		}
		http.Error(w, "failed to fetch component reports", http.StatusInternalServerError)
		return
	}

	// Apply filters
	filteredReports := s.filterReports(reports, params)

	// Apply pagination
	limit := 50 // default
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 100 {
		limit = *params.Limit
	}

	offset := 0 // default
	if params.Offset != nil && *params.Offset >= 0 {
		offset = *params.Offset
	}

	total := len(filteredReports)
	hasMore := offset+limit < total

	// Apply offset and limit
	end := offset + limit
	if end > total {
		end = total
	}

	var paginatedReports []storage.CheckReport
	if offset < total {
		paginatedReports = filteredReports[offset:end]
	}

	// Convert storage reports to API reports
	var apiReports []CheckReport
	for _, report := range paginatedReports {
		apiReport := s.convertToAPICheckReport(report)
		apiReports = append(apiReports, apiReport)
	}

	// Create response
	response := ComponentReportsResponse{
		Reports: apiReports,
		Pagination: Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// filterReports applies the query parameters to filter the reports
func (s *APIServer) filterReports(reports []storage.CheckReport, params GetComponentReportsParams) []storage.CheckReport {
	var filtered []storage.CheckReport

	for _, report := range reports {
		// Filter by status
		if params.Status != nil {
			var reportStatus storage.CheckStatus
			switch *params.Status {
			case GetComponentReportsParamsStatusPass:
				reportStatus = storage.CheckStatusPass
			case GetComponentReportsParamsStatusFail:
				reportStatus = storage.CheckStatusFail
			case GetComponentReportsParamsStatusDisabled:
				reportStatus = storage.CheckStatusDisabled
			case GetComponentReportsParamsStatusSkipped:
				reportStatus = storage.CheckStatusSkipped
			case GetComponentReportsParamsStatusUnknown:
				reportStatus = storage.CheckStatusUnknown
			case GetComponentReportsParamsStatusError:
				reportStatus = storage.CheckStatusError
			case GetComponentReportsParamsStatusCompleted:
				reportStatus = storage.CheckStatusCompleted
			default:
				continue // skip if status doesn't match
			}

			if report.Status != reportStatus {
				continue
			}
		}

		// Filter by check slug
		if params.CheckSlug != nil && *params.CheckSlug != "" {
			if report.Check.Slug != *params.CheckSlug {
				continue
			}
		}

		// Filter by since timestamp
		if params.Since != nil {
			if report.Timestamp.Before(*params.Since) {
				continue
			}
		}

		filtered = append(filtered, report)
	}

	return filtered
}

// convertToAPICheckReport converts a storage check report to an API check report
func (s *APIServer) convertToAPICheckReport(report storage.CheckReport) CheckReport {
	// Convert status to CheckReportStatus
	var status CheckReportStatus
	switch report.Status {
	case storage.CheckStatusPass:
		status = CheckReportStatusPass
	case storage.CheckStatusFail:
		status = CheckReportStatusFail
	case storage.CheckStatusDisabled:
		status = CheckReportStatusDisabled
	case storage.CheckStatusSkipped:
		status = CheckReportStatusSkipped
	case storage.CheckStatusUnknown:
		status = CheckReportStatusUnknown
	case storage.CheckStatusError:
		status = CheckReportStatusError
	case storage.CheckStatusCompleted:
		status = CheckReportStatusCompleted
	default:
		status = CheckReportStatusUnknown
	}

	apiReport := CheckReport{
		Id:        report.ID.String(),
		CheckSlug: report.Check.Slug,
		Status:    status,
		Timestamp: report.Timestamp,
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
