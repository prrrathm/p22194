# find-my-trip-client: Implementation Plan Summary

**Project**: Find My Trip Client  
**Scope**: MapLibre map client for @services/find-my-trip backend service  
**Status**: Ready for Build (Phase 4)  
**Created**: 2026-07-08  
**Target Completion**: 2026-07-22  

---

## 📋 What You Asked For

✅ **UI Layout**: Map on left (60%), frosted glass panel on right (40%)  
✅ **Design**: Apple-style frosted glass (glassmorphism)  
✅ **API Client**: Simple fetch (no axios/tRPC complexity)  
✅ **Error Handling**: Critical — API validation + error scenarios  
✅ **Backend**: localhost:8000  

---

## 📊 What's Been Planned

### 5 Epics, 15 Stories, 79 Business Complexity Points

| Epic | Title | Stories | BCP | WSJF | Blocking? |
|------|-------|---------|-----|------|-----------|
| **E02** | API Client + Error Handling | 3 | 21 | 42 | YES — build first |
| **E01** | Map UI + Frosted Glass Panel | 3 | 13 | 33 | No — depends on E02 |
| **E03** | Map Click + Data Fetch Integration | 3 | 16 | 38 | No — depends on E01, E02 |
| **E04** | Results Display + Loading States | 3 | 11 | 28 | No — depends on E03 |
| **E05** | API Integration Audit + Contracts | 3 | 18 | 42 | No — quality gate (last) |

**Execution Order**: E02 → E01 → E03 → E04 → E05

---

## 🗂 Artifacts Created

All artifacts in `/Users/prathm/Documents/p22194/specs/`:

```
specs/
├── product/
│   └── SCOPE_LATEST.yaml ...................... In/out of scope, success criteria
├── release-plan.yaml ........................... Epic ordering, gates, DOD
├── state.yaml .................................. Build-epic execution state
├── execution-status.yaml ....................... Story status tracker
├── epics/
│   ├── e01-map-frosted-panel/
│   │   └── epic.yaml ........................... 3 stories, 13 BCP
│   ├── e02-api-client-error-handling/
│   │   └── epic.yaml ........................... 3 stories, 21 BCP (DETAILED)
│   ├── e03-map-click-data-fetch/
│   │   └── epic.yaml ........................... 3 stories, 16 BCP
│   ├── e04-results-display-loading/
│   │   └── epic.yaml ........................... 3 stories, 11 BCP
│   └── e05-api-audit-contracts/
│       └── epic.yaml ........................... 3 stories, 18 BCP
└── verifications/
    ├── INTEGRATION_AUDIT_PLAN.md .............. Detailed audit strategy
    ├── postman-collection.json ............... (will be created during E05)
    └── INTEGRATION_AUDIT_FINAL.md ............ (will be created during E05)
```

**All files are ready to be consumed by the bigpowers planning spine.**

---

## 🚀 How to Execute

### Option 1: Step-by-Step (Recommended for First Epic)

```bash
cd /Users/prathm/Documents/p22194

# 1. Set active epic
bash scripts/bp-yaml-set.sh active_flow build_epic
bash scripts/bp-yaml-set.sh active_epic e02

# 2. Start build-epic (interactive, one step per invocation)
build-epic

# You'll be guided through:
# Step 0: security-review (threat model)
# Step 1: survey-context (confirm epic scope)
# Step 2: plan-work (detail tasks)
# Step 3: kickoff-branch (feature branch)
# Step 4: develop-tdd (red-green-refactor)
# Step 5: verify-work (UAT + gates)
# Step 6: audit-code (quality gate)
# Step 7: commit-message (Conventional Commits)
# Step 8: release-branch (PR/land)

# After E02 completes:
# repeat for E01, E03, E04, E05
```

### Option 2: Auto-Run (Faster, Less Interaction)

```bash
# Run entire epic without checkpoints (if you trust the plan)
build-epic --auto

# Or use --fast mode (coalesces read-and-report steps)
build-epic --fast
```

### Option 3: Resume Mid-Epic

```bash
# If interrupted, resume at current step
build-epic
# build-epic reads epic_cycle.current_step and runs only that step
```

---

## 💡 Key Architecture Decisions

### API Client Pattern
**Decision**: Simple fetch wrapper (no dependencies)

**Rationale**: Keep client lean, no axios/tRPC/GraphQL complexity

**Files**:
- `app/api/client.ts` — HttpClient class
- `app/api/errors.ts` — ApiError + retry logic
- `app/api/schemas.ts` — Zod validators
- `app/api/location.ts` — Location API wrapper

---

### Error Handling Strategy
**Decision**: Comprehensive error boundary + Zod validation

**Coverage**:
1. **Network errors** (no internet) → User message + retry
2. **Timeout** (>5s) → AbortController + user message
3. **Malformed response** → Zod validation catches + error
4. **Empty results** → "No results found" (not an error)
5. **Invalid coordinates** → Client-side validation before API call

**Files**:
- `app/components/ErrorBoundary.tsx` — React error boundary
- `app/api/errors.ts` — ApiError class with retryable flag
- `app/api/__tests__/*` — Unit + integration tests

---

### UI/UX Design
**Decision**: Apple-style frosted glass (glassmorphism)

**Styling**: TailwindCSS v4 + custom components
- Translucent white (rgba(255,255,255,0.7))
- Backdrop blur (blur-lg)
- Subtle borders (white/50 opacity)
- Shadow effects

**Layout**: Flexbox
- Left: 60% (map)
- Right: 40% (panel)
- Full viewport height

**Files**:
- `app/components/FrostedGlassPanel.tsx` — Reusable panel
- `app/components/Map.tsx` — MapLibre integration
- `app/components/ResultCard.tsx` — Individual result card
- `app/components/ResultsPanel.tsx` — Results container
- `app/components/LoadingSpinner.tsx` — Loading state

---

### Type Safety
**Decision**: Full TypeScript + Zod runtime validation

**Benefits**:
- Compile-time type checking (TypeScript strict mode)
- Runtime validation (Zod) ensures schema contract between client & service
- Catch API response mismatches early

**Files**:
- `app/api/schemas.ts` — Zod definitions (single source of truth)

---

## 🧪 Testing Strategy (Built Into Each Epic)

### Unit Tests
**Coverage**: API client, error handling, schema validation

**Files**:
- `app/api/__tests__/client.test.ts` — HttpClient happy path + errors
- `app/api/__tests__/errors.test.ts` — ApiError + retry logic
- `app/api/__tests__/schemas.test.ts` — Zod validation

---

### Integration Tests
**Coverage**: Client ↔ service contract, full data flow

**Files**:
- `app/api/__tests__/contract.test.ts` — Schema contract validation
- `app/__tests__/integration.error-scenarios.test.ts` — Error handling E2E
- `app/__tests__/integration.map-click.test.ts` — Click → API → display flow

---

### Audit & Verification (Epic E05)
**Coverage**: Postman collection, manual error scenarios

**Files**:
- `specs/verifications/postman-collection.json` — HTTP tests
- `specs/verifications/POSTMAN_RESULTS.json` — Test run results
- `specs/verifications/SCHEMA_CONTRACT.md` — Contract findings
- `specs/verifications/ERROR_SCENARIOS.md` — Manual error testing
- `specs/verifications/INTEGRATION_AUDIT_FINAL.md` — Final report

---

## 🎯 Success Criteria (Per Epic)

### E02 (API Client) ✅
- [ ] Fetch wrapper handles timeouts, retries, validation
- [ ] Zod schemas validate all responses
- [ ] Error boundary displays user-friendly messages
- [ ] Unit tests pass (100% API client coverage)
- [ ] TypeScript strict mode passes

### E01 (UI) ✅
- [ ] Map renders without errors
- [ ] Frosted glass panel styled correctly
- [ ] Layout: 60/40 split, full height
- [ ] No horizontal scroll
- [ ] Visual tests pass (3 browsers)

### E03 (Click + Fetch) ✅
- [ ] Click handler captures lat/lng
- [ ] All 3 endpoints called on click
- [ ] Results aggregated and passed to panel
- [ ] Error in one endpoint doesn't block others
- [ ] Integration tests pass

### E04 (Results Display) ✅
- [ ] Result cards render correctly
- [ ] Results organized by category
- [ ] Loading spinner shows during fetch
- [ ] Error state shows with retry button
- [ ] Empty state shows "No results found"

### E05 (Audit) ✅
- [ ] Postman collection: All tests pass
- [ ] Schema contract: Client & backend aligned
- [ ] Error scenarios: All 5 handled gracefully
- [ ] Integration audit: ✅ PASS
- [ ] Ready for production

---

## 📦 Dependencies

### New Packages to Install
- **maplibre-gl** (~300KB gzipped) — Map library
- **zod** (~15KB gzipped) — Schema validation

### Existing Dependencies
- React 19+
- React Router 8
- TailwindCSS 4
- TypeScript 5.9+

---

## 🔧 Environment Configuration

### .env File
```bash
VITE_API_BASE_URL=http://localhost:8000
```

### Backend Requirements
- Service running at localhost:8000
- CORS enabled (already configured in FastAPI)
- Endpoints:
  - POST /location/group
  - POST /location/places
  - POST /location/activities
  - GET /health (healthcheck)

---

## 📈 Timeline Estimate

| Epic | Duration | Complexity |
|------|----------|-----------|
| E02 | 2-3 days | HIGH (foundation) |
| E01 | 1-2 days | MEDIUM |
| E03 | 2 days | MEDIUM |
| E04 | 1-2 days | LOW |
| E05 | 1 day | MEDIUM |
| **Total** | **7-10 days** | |

*Estimates assume 4-6 hours/day focused work*

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Backend API downtime | 🔴 HIGH | Implement error messages + health check |
| MapLibre rendering issues | 🟡 MEDIUM | Test on Chrome, Firefox, Safari early |
| CORS failures | 🔴 HIGH | Verify backend CORS config (already done) |
| Schema mismatch | 🔴 HIGH | Validate with Zod + Postman (E05 audit) |
| Performance (slow API) | 🟡 MEDIUM | Set 5s timeout + show loading state |

---

## ✅ Next Steps (Immediate)

1. **Review this plan** — Make sure scope aligns with your vision
2. **Set active epic** — `bash scripts/bp-yaml-set.sh active_epic e02`
3. **Start build-epic** — `build-epic` (runs Step 0: security-review)
4. **Follow prompts** — TDD: develop-tdd will guide red-green-refactor

---

## 📞 Command Reference

```bash
# Check current state
cat specs/state.yaml
cat specs/execution-status.yaml

# Start first epic
bash scripts/bp-yaml-set.sh active_flow build_epic
bash scripts/bp-yaml-set.sh active_epic e02
build-epic

# Resume current epic
build-epic

# Run specific step
build-epic --step 4  # Develop-TDD

# View epic details
cat specs/epics/e02-api-client-error-handling/epic.yaml

# Track progress
cat specs/execution-status.yaml | grep status
```

---

## 📚 Documentation

All specs files are in `/Users/prathm/Documents/p22194/specs/`:
- **SCOPE**: What's in/out
- **RELEASE PLAN**: Epics, ordering, gates
- **EPICS**: Detailed stories + tasks
- **STATE**: Execution progress
- **EXECUTION STATUS**: Story status tracker
- **VERIFICATION**: Integration audit plan

---

## 🎬 You're Ready!

This plan is:
- ✅ Scoped (clear in/out boundaries)
- ✅ Sliced (vertical stories, testable)
- ✅ Planned (detailed tasks per story)
- ✅ Audited (E05 integration audit strategy)
- ✅ Ready for build-epic Phase 4

**Next action**: Run `build-epic` with E02 active to start the first epic (API Client + Error Handling).

---

**Plan Author**: AI Coding Agent  
**Date**: 2026-07-08  
**Status**: READY FOR BUILD-EPIC EXECUTION
