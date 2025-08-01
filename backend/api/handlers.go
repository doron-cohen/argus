package api

import (
	"encoding/json"
	"fmt"
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

// convertAPISStatusToStorageStatus converts API status to storage status
func (s *APIServer) convertAPISStatusToStorageStatus(status GetComponentReportsParamsStatus) (*storage.CheckStatus, error) {
	var statusValue storage.CheckStatus
	switch status {
	case GetComponentReportsParamsStatusPass:
		statusValue = storage.CheckStatusPass
	case GetComponentReportsParamsStatusFail:
		statusValue = storage.CheckStatusFail
	case GetComponentReportsParamsStatusDisabled:
		statusValue = storage.CheckStatusDisabled
	case GetComponentReportsParamsStatusSkipped:
		statusValue = storage.CheckStatusSkipped
	case GetComponentReportsParamsStatusUnknown:
		statusValue = storage.CheckStatusUnknown
	case GetComponentReportsParamsStatusError:
		statusValue = storage.CheckStatusError
	case GetComponentReportsParamsStatusCompleted:
		statusValue = storage.CheckStatusCompleted
	default:
		return nil, fmt.Errorf("invalid status: %v", status)
	}
	return &statusValue, nil
}

func (s *APIServer) GetComponentReports(w http.ResponseWriter, r *http.Request, componentId string, params GetComponentReportsParams) {
	ctx := r.Context()

	// Convert API parameters to storage types for database filtering
	var status *storage.CheckStatus
	if params.Status != nil {
		var err error
		status, err = s.convertAPISStatusToStorageStatus(*params.Status)
		if err != nil {
			// Invalid status, return 400 Bad Request
			http.Error(w, fmt.Sprintf("Invalid status parameter: %v", *params.Status), http.StatusBadRequest)
			return
		}
	}

	// Get pagination parameters
	limit := s.getLimit(params)
	offset := s.getOffset(params)
	latestPerCheck := params.LatestPerCheck != nil && *params.LatestPerCheck

	// Get reports with database-level filtering, pagination, and latest per check
	reports, total, err := s.Repo.GetCheckReportsForComponentWithPagination(ctx, componentId, status, params.CheckSlug, params.Since, limit, offset, latestPerCheck)
	if err != nil {
		if err == storage.ErrComponentNotFound {
			s.writeNotFoundError(w)
			return
		}
		http.Error(w, "failed to fetch component reports", http.StatusInternalServerError)
		return
	}

	// Convert storage reports to API reports
	apiReports := s.convertToAPICheckReports(reports)

	// Create pagination metadata
	hasMore := offset+limit < int(total)
	pagination := Pagination{
		Total:   int(total),
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}

	// Create response
	response := ComponentReportsResponse{
		Reports:    apiReports,
		Pagination: pagination,
	}

	s.writeJSONResponse(w, response)
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

// getLimit returns the limit parameter with validation
func (s *APIServer) getLimit(params GetComponentReportsParams) int {
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 100 {
		return *params.Limit
	}
	return 50 // default
}

// getOffset returns the offset parameter with validation
func (s *APIServer) getOffset(params GetComponentReportsParams) int {
	if params.Offset != nil && *params.Offset >= 0 {
		return *params.Offset
	}
	return 0 // default
}

// convertToAPICheckReports converts a slice of storage check reports to API check reports
func (s *APIServer) convertToAPICheckReports(reports []storage.CheckReport) []CheckReport {
	apiReports := make([]CheckReport, len(reports))
	for i, report := range reports {
		apiReports[i] = s.convertToAPICheckReport(report)
	}
	return apiReports
}

// writeJSONResponse writes a JSON response to the HTTP response writer
func (s *APIServer) writeJSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
