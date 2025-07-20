# Summary Generation Process Documentation

## Overview
This document outlines the flow involved in the GetSummaries endpoint (`/compat/wakatime/v1/users/{user}/summaries`) which provides WakaTime-compatible summary data for users.

## Endpoint Location
- **File**: `internal/api/compat_summaries.go:41`
- **Method**: `GetSummaries(w http.ResponseWriter, r *http.Request)`

## Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    GetSummaries Endpoint                       │
│               (compat_summaries.go:41)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                User Authentication                              │
│          utilities.CheckEffectiveUser()                        │
│               (compat_summaries.go:42)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│              Time Range Computation                             │
│              a.ComputeTimeRange()                               │
│               (compat_summaries.go:47)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│               Summary Data Loading                              │
│             a.loadUserSummaries()                               │
│               (compat_summaries.go:54)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│            Write Percentage Calculation                         │
│    a.services.Summary().GetHeartbeatsWritePercentage()         │
│               (compat_summaries.go:61)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│             Response Formatting                                 │
│              v1.NewSummariesFrom()                              │
│               (compat_summaries.go:68)                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│               JSON Response                                     │
│             helpers.RespondJSON()                               │
│               (compat_summaries.go:71)                         │
└─────────────────────────────────────────────────────────────────┘
```

## Detailed Process Breakdown

### 1. User Authentication
**Location**: `compat_summaries.go:42`
- Validates the user from the request path parameter
- Supports "current" as user identifier for authenticated user
- Returns error if user is not found or unauthorized

### 2. Time Range Computation
**Location**: `compat_summaries.go:47` → `ComputeTimeRange()` at line 76
- **Input Parameters**: `range`, `start`, `end`, `timezone`
- **Priority**: Range parameter takes precedence over start/end
- **Timezone Handling**: Uses user's timezone or specified timezone parameter
- **Date Parsing**: Supports various formats and interval identifiers
- **WakaTime Compatibility**: Adjusts end date to be inclusive (adds end-of-day)

### 3. Summary Data Loading
**Location**: `compat_summaries.go:54` → `loadUserSummaries()` at line 126

#### 3.1 Daily Interval Splitting
- Splits the overall time range into daily intervals using `utils.SplitRangeByDays()`
- Each day gets processed separately for summary generation

#### 3.2 Filter Parsing
- Extracts filtering parameters from request: `project`, `language`, `editor`, `operating_system`, `machine`, `label`
- Uses `helpers.ParseSummaryFilters()` to parse these into filter objects

#### 3.3 Summary Generation per Day
**Core Logic**: `a.services.Summary().RetrieveWithAliases()` at line 144
- **Service Chain**: `SummaryService.RetrieveWithAliases()` → `SummaryService.Aliased()` → `SummaryService.Retrieve()`
- **Simplified API**: Eliminates the need to pass function parameters
- **Caching**: Implements intelligent caching based on cache keys
- **Alias Resolution**: Resolves entity aliases and project labels
- **Data Sources**: Combines pre-calculated summaries from database with on-demand generation

#### 3.4 Summary Service Flow (`services/summary.go`)

```
SummaryService.RetrieveWithAliases() / SummaryService.SummarizeWithAliases()
    │
    └── SummaryService.Aliased()
        │
        ├── Cache Check (if enabled)
        │
        ├── Call SummaryRetriever (Retrieve or Summarize method)
        │   │
        │   ├── Check for existing summaries in database
    │   │
    │   ├── Identify missing intervals (getMissingIntervals)
    │   │
    │   ├── Generate summaries for missing intervals
    │   │   └── SummaryService.Summarize()
    │   │       ├── Fetch duration data
    │   │       ├── Aggregate by entity types (parallel)
    │   │       └── Create Summary objects
    │   │
    │   └── Merge existing + generated summaries
    │
    ├── Apply alias resolution
    │
    ├── Apply project labels
    │
    └── Cache result (if enabled)
```

### 4. Write Percentage Calculation
**Location**: `compat_summaries.go:61`
- Calculates the percentage of heartbeats that represent write operations
- Delegates to `HeartbeatService.GetHeartbeatsWritePercentage()`
- Used for WakaTime compatibility metrics

### 5. Response Formatting
**Location**: `compat_summaries.go:68` → `v1.NewSummariesFrom()`

#### 5.1 Data Transformation
- Converts internal `models.Summary` format to WakaTime-compatible format
- Processes each summary's entity collections (projects, languages, editors, etc.)
- Calculates cumulative totals and daily averages

#### 5.2 Parallel Processing
- Uses goroutines for parallel conversion of different entity types
- Optimizes performance for large summary datasets

#### 5.3 Response Structure (`SummariesViewModel`)
```json
{
  "data": [
    {
      "grand_total": {...},
      "projects": [...],
      "languages": [...],
      "editors": [...],
      "operating_systems": [...],
      "machines": [...],
      "labels": [...],
      "branches": [...],
      "entities": [...],
      "range": {...}
    }
  ],
  "end": "2021-04-30T23:59:59Z",
  "start": "2021-04-29T00:00:00Z",
  "timezone": "Europe/Berlin",
  "write_percentage": 85.5
}
```

## Key Service Components

### Summary Service Interface (`services/services.go:88-99`)
```go
type ISummaryService interface {
    // Legacy method (still supported for backward compatibility)
    Aliased(time.Time, time.Time, *models.User, SummaryRetriever, *models.Filters, bool) (*models.Summary, error)
    
    // Simplified methods (recommended)
    RetrieveWithAliases(time.Time, time.Time, *models.User, *models.Filters, bool) (*models.Summary, error)
    SummarizeWithAliases(time.Time, time.Time, *models.User, *models.Filters, bool) (*models.Summary, error)
    
    // Core methods
    Retrieve(time.Time, time.Time, *models.User, *models.Filters) (*models.Summary, error)
    Summarize(time.Time, time.Time, *models.User, *models.Filters) (*models.Summary, error)
    GetHeartbeatsWritePercentage(string, time.Time, time.Time) (float64, error)
    // ... CRUD operations
}
```

### Models Summary Structure (`models/summary.go:31-54`)
```go
type Summary struct {
    FromTime         CustomTime      `json:"from"`
    ToTime           CustomTime      `json:"to"`
    Projects         SummaryItems    `json:"projects"`
    Languages        SummaryItems    `json:"languages"`
    Editors          SummaryItems    `json:"editors"`
    OperatingSystems SummaryItems    `json:"operating_systems"`
    Machines         SummaryItems    `json:"machines"`
    Labels           SummaryItems    `json:"labels"`
    Branches         SummaryItems    `json:"branches"`
    Entities         SummaryItems    `json:"entities"`
    Categories       SummaryItems    `json:"categories"`
    NumHeartbeats    int             `json:"num_heartbeats"`
}
```

## API Simplification Changes

### Before: Complex Function Parameter Pattern
```go
// Old ceremonious API requiring function parameters and array indexing
summary, err := a.services.Summary().Aliased(
    interval[0], interval[1], user,  // Cryptic array indexing
    a.services.Summary().Retrieve,   // Function parameter!
    filters, end.After(time.Now())
)
```

### After: Simplified Direct Methods
```go
// New structured API with meaningful field names
request := summarytypes.NewSummaryRequest(interval.Start, interval.End, user).WithFilters(filters)
if end.After(time.Now()) {
    request = request.WithoutCache()
}
options := summarytypes.DefaultProcessingOptions()
summary, err := a.services.Summary().Generate(request, options)
```

### Benefits of Simplification
- **Eliminated ceremonious function parameters**: No more passing `service.Retrieve` as parameter
- **Meaningful interval access**: `interval.Start` and `interval.End` instead of cryptic `[0]` and `[1]`
- **Structured request objects**: Clear, self-documenting API with fluent interface
- **Type safety**: Strong typing prevents array index errors
- **Improved readability**: Code intent is immediately clear
- **Reduced complexity**: Fewer parameters to understand and maintain
- **Maintained functionality**: All existing features (caching, aliasing, project labels) work identically
- **Backward compatibility**: Legacy `Aliased` method still available

## Performance Optimizations

1. **Intelligent Caching**: Service-level caching with configurable TTL
2. **Parallel Processing**: Goroutines for entity type conversions
3. **Hybrid Data Strategy**: Combines pre-calculated and on-demand summaries
4. **Daily Interval Processing**: Efficient handling of large time ranges
5. **Database Optimization**: Batched queries for missing intervals

## Error Handling

- **Authentication Errors**: 401/403 responses for invalid users
- **Parameter Validation**: 400 responses for invalid time ranges or parameters
- **Service Errors**: 500 responses for database or processing failures
- **Graceful Degradation**: Continues processing even if some summaries fail

## WakaTime Compatibility Features

- **Inclusive End Dates**: Adjusts end times to match WakaTime behavior
- **Response Format**: Matches WakaTime API response structure exactly
- **Parameter Support**: Supports WakaTime query parameters and filtering
- **Timezone Handling**: Respects user timezones and parameter overrides