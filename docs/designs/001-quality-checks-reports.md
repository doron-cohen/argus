# Design: Quality Checks & Reports System

## Context
**Users**: Maintainers and managers of software components
**Pain**: Lack of visibility into how well software components are maintained, leading to reactive responses to incidents rather than proactive maintenance
**Solution**: A reporting system that tracks quality metrics for each component, providing dashboards to understand component health and maintenance status

## Goals
- Enable maintainers and managers to understand component health at a glance
- Provide historical tracking of quality metrics to identify trends and maintenance needs
- Support proactive maintenance decisions based on quality data
- Allow external systems to report check results via REST API

## Constraints
- Push-only model (no polling from Argus)
- Low scale (10 requests/second max)
- Simple check types initially (unit tests, build, linter)
- Metadata is execution-focused, not component-static

## Design

### Data Model

#### Check Definition
```go
type Check struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
    Slug        string    `gorm:"not null;uniqueIndex;size:100"` // e.g., "unit-tests", "build", "linter"
    Name        string    `gorm:"not null;size:255"`             // "Unit Tests", "Build Process", "Code Linting"
    Description string    `gorm:"type:text"`
    CreatedAt   time.Time `gorm:"autoCreateTime"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
```

#### Check Report
```go
type CheckReport struct {
    ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
    CheckID     uuid.UUID   `gorm:"type:uuid;not null;index:idx_check_timestamp"`
    ComponentID uuid.UUID   `gorm:"type:uuid;not null;index:idx_component_check"`
    Status      CheckStatus `gorm:"type:varchar(20);not null;index:idx_check_status"`
    Timestamp   time.Time   `gorm:"not null;index:idx_check_timestamp"`
    Details     JSONB       `gorm:"type:jsonb;index:,type:gin"`     // Check-specific data (coverage %, warnings, etc.)
    Metadata    JSONB       `gorm:"type:jsonb;index:,type:gin"`     // Execution context (CI job, environment, duration)
    CreatedAt   time.Time   `gorm:"autoCreateTime"`
    UpdatedAt   time.Time   `gorm:"autoUpdateTime"`

    // Relationships
    Check     Check     `gorm:"foreignKey:CheckID;constraint:OnDelete:CASCADE"`
    Component Component `gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
}
```

**Note**: Checks are auto-created when reports are submitted with new check slugs. The system looks up the check by slug and creates it if it doesn't exist, then uses the CheckID for the relationship.

type CheckStatus string

const (
    CheckStatusPass      CheckStatus = "pass"
    CheckStatusFail      CheckStatus = "fail"
    CheckStatusDisabled  CheckStatus = "disabled"
    CheckStatusSkipped   CheckStatus = "skipped"
    CheckStatusUnknown   CheckStatus = "unknown"    // Report submission error
    CheckStatusError     CheckStatus = "error"      // Check evaluation failed
    CheckStatusCompleted CheckStatus = "completed"  // Check ran, but pass/fail unclear
)
```

### API Design

#### Reports Module (New)
```
POST   /reports                          # Submit check report
```

#### API Module (Extended)
```
GET    /api/v1/components/{id}/reports   # Get reports for component
```

### Report Submission Format
```json
{
  "check_slug": "unit-tests",
  "component_id": "auth-service",
  "status": "pass",
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "coverage_percentage": 85.5,
    "tests_passed": 150,
    "tests_failed": 0,
    "duration_seconds": 45
  },
  "metadata": {
    "ci_job_id": "12345",
    "environment": "staging",
    "branch": "main",
    "commit_sha": "abc123",
    "execution_duration_ms": 45000
  }
}
```

**Note**: When a report is submitted with a new `check_slug`, the system automatically creates a check definition with the slug as the identifier and a default name/description.

### Query Patterns (Phase 1)

#### Basic Status Counts
```sql
SELECT 
    status,
    COUNT(*) as count
FROM check_reports 
WHERE check_id = ? AND timestamp >= ?
GROUP BY status
```

#### Latest Status per Component/Check
```sql
SELECT DISTINCT ON (component_id, check_id)
    component_id, check_id, status, timestamp
FROM check_reports 
ORDER BY component_id, check_id, timestamp DESC
```

### Component Health Dashboard (Phase 1)
- **Component Reports**: View check reports for a specific component
- **Basic Status**: See latest check results and status history

### Implementation Scope (Phase 1 Only)

#### Reports Module
1. OpenAPI specification for report submission
2. Generated server/client code for reports module
3. Report submission handler (POST /reports)
4. Database schema for check reports
5. Repository methods for report storage and retrieval

#### API Module Extension
1. New endpoint to get reports for a component (GET /api/v1/components/{id}/reports)
2. Integration with existing component API

## Tradeoffs

### Simplicity vs. Flexibility
- **Chosen**: JSONB for metadata/details allows flexible reporting without schema changes
- **Rejected**: Strict schema would require migrations for new check types

### Push vs. Pull Model
- **Chosen**: Push-only model simplifies architecture and reduces complexity
- **Rejected**: Polling would require scheduling infrastructure and state management

### Check Configuration
- **Chosen**: Simple pass/fail status with details in JSONB
- **Rejected**: Complex rule engine would add significant complexity for minimal benefit

### Data Retention
- **Chosen**: Keep all historical data for trend analysis
- **Rejected**: Automatic cleanup would lose valuable historical insights

### API Design
- **Chosen**: RESTful API following existing patterns
- **Rejected**: GraphQL would add complexity without clear benefits at this scale
