# Searchable Project Filter & Branch Filter with Server-Side Search

## Overview
- Add a **searchable Project filter** (client-side type-to-filter combobox) and a new
  **Branch filter** with debounced, server-side search on the summary/project pages.
- Solves: the native `<select>` Project dropdown is hard to scan with many projects, and there
  is no way to filter a project's stats down to a single branch from the UI (the backend
  already supports `?branch=`).
- Integration: reuses the existing "select a value → reload page with a query param" filter
  mechanism. Backend `ParseSummaryFilters` already maps `?branch=` to `SummaryBranch`
  (`helpers/summary.go:68`), so only a search endpoint + UI are new.
- Design doc: `docs/2026-07-13-search-by-branch-design.md`.

## Context (from discovery)
- Files/components involved:
  - Frontend: `views/entity-filter.tpl.html` (petite-vue template), `views/summary.tpl.html`
    (wires filters + injects `wakapiData`), `static/assets/js/components/entity-filter.js`
    (existing native-select component). New: `static/assets/js/components/combobox-filter.js`.
  - Backend: `routes/api/heartbeat.go` (+ `routes/api/heartbeat_test.go`), `services/services.go`
    (`IHeartbeatService`), `services/heartbeat.go`, `repositories/heartbeat.go`,
    `mocks/heartbeat_service.go`.
- Related patterns found:
  - `EntityFilter` (petite-vue via `v-scope`) applies filters by mutating
    `window.location.search` and reloading; includes an `unknown` → `-` mapping.
  - `HeartbeatRepository.GetEntitySetByUser` (`repositories/heartbeat.go:216`) is the closest
    existing query (distinct entity column for a user); the new search adds project + LIKE + limit.
  - API handlers register routes in `RegisterRoutes` and are wired in `main.go` (~line 304).
    `SummaryApiHandler` uses `NewAuthenticateMiddleware(h.userSrvc).Handler` (cookie + API key),
    which is what the new endpoint needs.
  - Handler tests use `mocks/heartbeat_service.go` (testify mocks).
- Dependencies identified: `gorm.io/gorm`, `github.com/go-chi/chi/v5`, `github.com/stretchr/testify`.

## Development Approach
- **Testing approach**: Regular (code first, then tests) for Go layers.
- Complete each task fully before moving to the next; make small, focused changes.
- **CRITICAL: every task with Go code MUST include new/updated Go tests** (success + error cases).
  - Frontend (petite-vue/JS) has **no unit-test harness** in this repo (npm scripts are
    build-only). Frontend tasks are verified by manual testing (see Post-Completion). Where a
    task is frontend-only, the "run tests" step means: build assets and manually verify.
- **CRITICAL: all tests must pass before starting the next task.**
- Run `go build ./...` and `go test ./...` after each Go change. Maintain backward compatibility
  (existing filters and endpoints must keep working).

## Testing Strategy
- **Unit tests (Go)**: required for the new handler (`routes/api/heartbeat_test.go`) using the
  existing testify mock service. Add a `SearchBranchesByUser` method to `mocks/heartbeat_service.go`.
- **Repository**: no unit-test harness exists (needs a live DB), consistent with the codebase —
  verified via manual/integration testing (Post-Completion).
- **Frontend/e2e**: no e2e harness in repo. `ComboboxFilter` verified via manual testing
  (Post-Completion): project type-to-filter, branch debounce + 3-char minimum, selection reload,
  mount-time pre-fill from URL.

## Progress Tracking
- Mark completed items with `[x]` immediately when done.
- Add newly discovered tasks with ➕ prefix; document blockers with ⚠️ prefix.
- Keep this plan in sync with actual work; update if scope changes.

## Solution Overview
- Backend: `GET /api/branches?project=<optional>&q=<query>` → JSON `[]string` of distinct
  branch names matching a case-insensitive substring, scoped to `project` when given, capped at 50.
  Three layers: repository query → service passthrough → chi handler with standard auth.
- Frontend: a reusable petite-vue `ComboboxFilter` (text input + dropdown). Project uses it in
  **local** mode (filters preloaded `availableProjectNames`); Branch uses it in **remote** mode
  (debounced fetch, min 3 chars) and is rendered only on project-detail pages. Selecting an
  option reloads the page with the corresponding query param, reusing existing filter plumbing.

## Technical Details
- Entity constant: `models.SummaryBranch` (index 6); `models.GetEntityColumn(SummaryBranch)` == `"branch"`.
- Repository query (parameterized, no interpolation):
  `Model(&models.Heartbeat{}).Distinct("branch").Where("user_id = ?", userId)`, add
  `Where("branch <> ''")`, `Where("LOWER(branch) LIKE ?", "%"+strings.ToLower(query)+"%")`,
  and `Where("project = ?", project)` only when `project != ""`; `Order("branch").Limit(limit)`.
- Endpoint constraints: `q` trimmed; if `< 3` chars return `200 []`; `limit` = 50.
- Frontend fetch URL is relative (`api/branches?...`) so it inherits the base path and session cookie
  (same convention as the badge `img src="api/badge/…"`).
- Debounce 1200 ms; min chars 3 — passed as component options, not hard-coded in the template.

## What Goes Where
- **Implementation Steps** (`[ ]`): repo/service/handler code + Go tests, and the frontend
  component + wiring (build-verified).
- **Post-Completion** (no checkboxes): manual UI verification and DB-backed integration checks.

## Implementation Steps

### Task 1: Repository branch search query

**Files:**
- Modify: `repositories/heartbeat.go`

- [x] add `SearchBranchesByUser(userId, project, query string, limit int) ([]string, error)` to `HeartbeatRepository`
- [x] build a GORM query: `Distinct("branch")` over `models.Heartbeat`, `user_id = userId`, `branch <> ''`, case-insensitive `LOWER(branch) LIKE ?` with `%query%` (lowercased), and `project = ?` only when `project != ""`
- [x] order by `branch`, apply `Limit(limit)`, scan into `[]string`, return
- [x] ensure all inputs are parameterized (no `fmt.Sprintf` into the query) to prevent SQL injection
- [x] run `go build ./...` — must succeed before next task

### Task 2: Service method + interface + mock

**Files:**
- Modify: `services/services.go`
- Modify: `services/heartbeat.go`
- Modify: `mocks/heartbeat_service.go`

- [x] add `SearchBranchesByUser(string, string, string, int) ([]string, error)` to the `IHeartbeatService` interface in `services/services.go`
- [x] implement `SearchBranchesByUser` on `HeartbeatService` in `services/heartbeat.go` as a passthrough to the repository, trimming empty/whitespace-only results (mirror the filtering in `GetEntitySetByUser`); do not cache
- [x] add the matching mock method to `mocks/heartbeat_service.go` (testify `m.Called(...)` returning `[]string, error`)
- [x] run `go build ./...` — must succeed before next task

### Task 3: `GET /api/branches` handler + tests

**Files:**
- Modify: `routes/api/heartbeat.go`
- Modify: `routes/api/heartbeat_test.go`

- [x] add a new route group in `HeartbeatApiHandler.RegisterRoutes` using `middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler` (cookie + API key), registering `r.Get("/branches", h.GetBranches)`
- [x] implement `GetBranches`: resolve principal via `middlewares.GetPrincipal(r)`; read `project` and `q` query params; if `len(strings.TrimSpace(q)) < 3` write `200` with JSON `[]`; else call `heartbeatSrvc.SearchBranchesByUser(user.ID, project, q, 50)`; on error log + `500`; on success JSON-encode the `[]string`
- [x] write handler test: `q` shorter than 3 chars returns `200` and empty array without calling the service
- [x] write handler test: valid `q` (+ `project`) forwards exact args to the mocked service and returns the mocked branches as a JSON string array
- [x] write handler test: unauthenticated request is rejected (mirror existing auth assertions in `heartbeat_test.go`)
- [x] run `go test ./routes/... ./services/... ./repositories/...` — must pass before next task

### Task 4: `ComboboxFilter` petite-vue component

**Files:**
- Create: `static/assets/js/components/combobox-filter.js`
- Modify: `views/entity-filter.tpl.html`

- [x] add a `#combobox-filter-template` block to `views/entity-filter.tpl.html`: a label, a text input bound to `query`, and a dropdown `<ul>`/list of `visibleOptions` shown when `open`; reuse existing `.entity-filter-control` / input styling
- [x] create `combobox-filter.js` exporting `ComboboxFilter({ type, options, selection, remote, minChars, debounceMs, project })` (petite-vue object with `$template`, `$delimiters: ['${','}']`)
- [x] implement local mode: `visibleOptions` = `options` filtered by case-insensitive substring of `query`
- [x] implement remote mode: on input clear any pending timer; if `query.length < minChars` clear options + show hint; else start a `debounceMs` timer that fetches `api/branches?project=<project>&q=<query>`, sets `visibleOptions`, toggles a `loading` flag, and handles fetch errors gracefully (clear options, `console` log, no crash)
- [x] implement `select(option)`: set `selection`, then update `window.location.search` (build `URLSearchParams`, set/delete `type` with the existing `unknown` → `-` mapping) and reload — matching `entity-filter.js`
- [x] implement `mounted`: pre-fill `query`/`selection` when the URL already has the `type` param; add outside-click / Escape handling to close the dropdown
- [x] frontend has no unit tests — run `npm run build` (or `build:tailwind`) to confirm assets compile; manual verification is tracked in Post-Completion

### Task 5: Wire Project + Branch filters into the summary page

**Files:**
- Modify: `views/summary.tpl.html`

- [ ] add `<script src="assets/js/components/combobox-filter.js?v={{ getCacheBuster }}"></script>` alongside the existing component scripts
- [ ] replace the Project `EntityFilter({...})` `v-scope` with `ComboboxFilter({ type: 'project', options: wakapiData.availableProjectNames.toSorted(), selection: null, remote: false })`
- [ ] add a Branch `ComboboxFilter` immediately after Project, wrapped in `{{ if .IsProjectDetails }} … {{ end }}`, configured `{ type: 'branch', options: [], selection: null, remote: true, minChars: 3, debounceMs: 1200, project: '{{ .GetProjectFilter }}' }`
- [ ] leave Language / Machine / Label / Category as existing `EntityFilter` native selects (no change)
- [ ] run `npm run build` to confirm templates/assets compile; manual verification tracked in Post-Completion

### Task 6: Verify acceptance criteria
- [ ] verify all requirements from Overview are implemented (searchable project filter; branch filter, project-page only, debounced 1.2s, min 3 chars, project-scoped)
- [ ] verify edge cases: empty/short branch query returns no request/`[]`; no project selected still returns results (global); selecting a branch reloads with `?branch=` and the backend applies it
- [ ] run full Go test suite: `go test ./...`
- [ ] run `go vet ./...` and `go build ./...`
- [ ] confirm no regression to the other four filters and existing API endpoints

### Task 7: Update documentation
- [ ] update `README.md` / API docs if the new `GET /api/branches` endpoint should be listed
- [ ] update `CLAUDE.md` if new patterns were introduced (e.g. the `ComboboxFilter` component)
- [ ] move this plan to `docs/plans/completed/` (create the directory if needed)

## Post-Completion
*Items requiring manual intervention or external systems — no checkboxes, informational only*

**Manual verification** (frontend has no automated test harness):
- Project filter: typing filters the dropdown client-side; selecting reloads with `?project=`;
  reopening a filtered page pre-fills the current selection.
- Branch filter: only visible on a project-detail page; typing < 3 chars shows the hint and
  fires no request; after 1.2 s pause with ≥ 3 chars a request is sent; results render;
  selecting reloads with `?branch=` and the branch chart/stats reflect the filter.
- Confirm the request URL respects a non-root base path deployment.

**Integration / DB verification** (repository layer has no unit tests):
- Against a populated database, confirm `GET /api/branches?project=X&q=abc` returns distinct,
  case-insensitive substring matches scoped to project X, capped at 50, excluding empty branches.
- Confirm behavior on both SQLite and Postgres/MySQL if targeted (LOWER/LIKE semantics).
