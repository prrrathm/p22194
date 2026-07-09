# Integration Audit Plan: find-my-trip-client ↔ @services/find-my-trip

**Status**: Pre-audit (E05 pending)  
**Created**: 2026-07-08  
**Audit Owner**: [to be assigned]  

---

## 1. Audit Scope

### APIs Under Test
1. **POST /location/group** — Find group tours at coordinates
2. **POST /location/places** — Find places to see at coordinates
3. **POST /location/activities** — Find activities at coordinates

### Client Integration Points
- `app/api/client.ts` — HTTP client
- `app/api/location.ts` — Location API wrapper
- `app/components/Map.tsx` — Map click handler → API calls
- `app/components/ResultsPanel.tsx` — Result display
- Error handling throughout

### Service Integration Points
- `services/find-my-trip/internal/router/router.py` — All 3 endpoints
- `services/find-my-trip/internal/schemas/location.py` — Request validation
- `services/find-my-trip/internal/schemas/itinerary.py` — Response structure

---

## 2. Contract Validation

### Request Contract

**Endpoint**: `POST /location/{group|places|activities}`

**Expected Request Shape**:
```typescript
{
  latitude: number,      // -90 to 90
  longitude: number      // -180 to 180
}
```

**Validation by Client** (before API call):
- [ ] Latitude in range [-90, 90]
- [ ] Longitude in range [-180, 180]
- [ ] Both fields present and non-null
- [ ] Both are numbers (not strings)

**Validation by Backend**:
- [✓] Already done in `LocationRequest` schema (Pydantic)

**Test Cases**:
1. Valid: `{ latitude: 37.7749, longitude: -122.4194 }` → 200
2. Invalid lat: `{ latitude: 91, longitude: 0 }` → 422
3. Invalid lng: `{ latitude: 0, longitude: 181 }` → 422
4. Missing lat: `{ longitude: 0 }` → 400
5. Missing lng: `{ latitude: 0 }` → 400
6. Invalid type: `{ latitude: "37.7749", longitude: 0 }` → 400 or coerce?

---

### Response Contract

**Expected Response Shape**:
```typescript
{
  results: [
    {
      title: string,
      url: string,
      snippet: string
    },
    ...
  ]
}
```

**Validation by Client** (Zod schema):
- [ ] Response is valid JSON
- [ ] `results` field is an array
- [ ] Each result has `title`, `url`, `snippet` (all strings)
- [ ] No additional required fields

**Validation by Backend**:
- [✓] Already done in `LocationResponse` schema (Pydantic)

**Edge Cases**:
1. Empty results: `{ results: [] }` → Client shows "No results found"
2. Malformed JSON: `{ results: ...invalid...` → Client throws validation error
3. Missing `snippet`: `{ results: [{ title: "...", url: "..." }] }` → Should this fail or default to ""?
4. Extra fields: `{ results: [...], extra_field: "..." }` → Zod should strip

---

## 3. HTTP Contract

### Headers

**Request**:
- `Content-Type: application/json`
- `Accept: application/json`

**Response**:
- `Content-Type: application/json`
- `Access-Control-Allow-Origin: *` (CORS)

**Verification**:
- [ ] Postman: Inspect response headers
- [ ] Browser DevTools: Network tab, look for CORS headers

### Status Codes

| Scenario | Expected | Actual |
|----------|----------|--------|
| Valid request, found results | 200 OK | ? |
| Valid request, no results | 200 OK | ? |
| Invalid coordinates | 422 Unprocessable Entity | ? |
| Missing fields | 400 Bad Request | ? |
| Malformed JSON | 400 Bad Request | ? |
| Server error | 500 Internal Server Error | ? |

---

## 4. Error Scenarios

### Network Errors
**Scenario**: Client cannot reach backend

**Trigger**: 
```bash
# Disconnect network or stop backend
curl http://localhost:8000/health
# Connection refused
```

**Expected Behavior**:
- [ ] Client catches error
- [ ] User sees: "Unable to connect to server. Check your connection and try again."
- [ ] Retry button appears

**Test**:
```typescript
// Mock fetch to throw NetworkError
global.fetch = jest.fn(() => 
  Promise.reject(new TypeError('Failed to fetch'))
)
// Click map, verify error message
```

---

### Timeout Errors
**Scenario**: Backend takes > 5 seconds to respond

**Trigger**:
```bash
# Add 10s delay to backend endpoint
# (temporarily modify main.py or use mock)
```

**Expected Behavior**:
- [ ] Request aborts after 5s
- [ ] User sees: "Request took too long. Please try again."
- [ ] Retry button appears

**Test**:
```typescript
// Mock fetch to delay >5s
global.fetch = jest.fn(() => 
  new Promise((_, reject) => 
    setTimeout(() => reject(new Error('timeout')), 6000)
  )
)
```

---

### Malformed Response
**Scenario**: Backend returns invalid JSON or wrong schema

**Trigger**:
```bash
# Temporarily return malformed response from backend
# e.g., { results: "not an array" }
```

**Expected Behavior**:
- [ ] Zod validation catches mismatch
- [ ] User sees: "Unexpected data format from server."
- [ ] Retry button appears

**Test**:
```typescript
// Mock fetch to return invalid schema
global.fetch = jest.fn(() => 
  Promise.resolve(new Response(
    JSON.stringify({ results: "invalid" })
  ))
)
```

---

### Empty Results
**Scenario**: Valid request but no results found

**Trigger**:
```bash
# Search remote ocean (no results)
# lat: 0, lng: 0 (international waters)
```

**Expected Behavior**:
- [ ] Request succeeds (200)
- [ ] `results: []` is returned
- [ ] User sees: "No results found for this location."
- [ ] No error state
- [ ] Map still interactive

---

### Server Errors (5xx)
**Scenario**: Backend returns 500, 502, 503

**Trigger**:
```bash
# Simulate backend crash or SearXNG unavailable
# Modify backend to return 503
```

**Expected Behavior**:
- [ ] Client catches error
- [ ] User sees: "Server error. Please try again in a moment."
- [ ] Retry button appears
- [ ] Retry logic retries with exponential backoff (1s, 2s, 4s)

---

## 5. Data Flow Validation

### Happy Path
```
User clicks map at (37.7749, -122.4194)
        ↓
Client captures lat/lng
        ↓
Client validates with LocationRequestSchema
        ↓
Client calls:
  - POST /location/group { lat, lng }
  - POST /location/places { lat, lng }
  - POST /location/activities { lat, lng }
        ↓
Backend validates with LocationRequest schema
        ↓
Backend calls SearXNG search
        ↓
Backend returns { results: [...] }
        ↓
Client validates response with LocationResponseSchema
        ↓
Client aggregates results by category
        ↓
ResultsPanel displays results
        ↓
User sees "Group Tours", "Places", "Activities" tabs/sections
```

**Verification**:
- [ ] All 3 API calls complete
- [ ] Results aggregated correctly
- [ ] No console errors
- [ ] All types correct (TypeScript strict mode)

---

## 6. Audit Checklist

### Pre-Audit Setup
- [ ] Backend running at localhost:8000
- [ ] Backend healthcheck: `curl http://localhost:8000/health` → 200
- [ ] Client at localhost:5173
- [ ] Browser console open (watch for errors)
- [ ] Network tab open (watch for requests)

### API Contract Tests
- [ ] Postman: All happy path tests pass
- [ ] Postman: All error scenario tests pass
- [ ] Schema: Client Zod schemas match backend Pydantic schemas
- [ ] Headers: CORS headers present
- [ ] Status codes: Match expected (200 for success, 4xx/5xx for errors)

### Error Handling Tests
- [ ] Network error: Handled gracefully, user message shown
- [ ] Timeout: Handled after 5s, user message shown
- [ ] Malformed response: Zod validation rejects, error shown
- [ ] Empty results: Handled with "No results found" message
- [ ] Invalid coordinates: Rejected before API call (client-side validation)

### Integration Tests
- [ ] Map click → API call → Results display (happy path)
- [ ] Click error → Retry button → Click retry → Success
- [ ] Verify no unhandled promise rejections
- [ ] Verify no memory leaks (multiple clicks)

### Code Quality
- [ ] TypeScript strict mode passes
- [ ] No eslint/prettier violations
- [ ] Unit tests for API client pass
- [ ] Unit tests for error handling pass
- [ ] Integration test for schema contract passes

### Performance
- [ ] API response < 5s (most cases)
- [ ] Map render time < 2s
- [ ] Results display < 1s after API response
- [ ] No excessive re-renders (React DevTools)

---

## 7. Audit Report Output

Create `specs/verifications/INTEGRATION_AUDIT_FINAL.md`:

```markdown
# Integration Audit Report: find-my-trip-client ↔ find-my-trip

**Date**: [date]
**Auditor**: [name]
**Status**: ✅ PASS / ❌ FAIL

## Summary
[Executive summary of findings]

## Test Results
- Happy Path: ✅ 3/3 pass
- Error Scenarios: ✅ 5/5 pass
- Schema Contract: ✅ All validations pass
- Integration: ✅ E2E flow works

## Issues Found
[List any failures, gaps, recommendations]

## Sign-off
- [ ] Ready for production
- [ ] Remediation items (if any)

```

---

## 8. Remediation Strategy

If audit finds issues:

1. **Minor** (doesn't block release):
   - Document in issue
   - Create follow-up story in v1.1 backlog

2. **Major** (blocks release):
   - Return to develop-tdd phase
   - Fix in epic
   - Re-run audit
   - Don't advance to release-branch until ✅

3. **API Mismatch** (backend + client disagree):
   - Contact backend team
   - Align on canonical schema
   - Update both sides
   - Re-audit

---

## Next Steps

1. **E02 (API Client)**: Implement fetch wrapper, error handling, Zod validation
2. **E01-E04**: Implement UI, integration, results display
3. **E05 (Audit)**: Run this audit plan against completed client + running backend
4. **Findings**: Document in INTEGRATION_AUDIT_FINAL.md
5. **Release**: If ✅ PASS, proceed to commit-message → release-branch

---

**Audit Plan Ready**. Proceed to build-epic Phase 4.
