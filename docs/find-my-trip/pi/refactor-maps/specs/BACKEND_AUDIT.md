# Backend Service Audit: @services/find-my-trip

**Auditor Note**: This document outlines what we're expecting from the backend and identifies any gaps or fixes needed before the client can integrate properly.

**Status**: Pre-audit  
**Date**: 2026-07-08  
**Backend Location**: `/Users/prathm/services/find-my-trip/`

---

## 1. Current State

### Endpoints
- ✅ `POST /location/group` — Find group tours
- ✅ `POST /location/places` — Find places to see
- ✅ `POST /location/activities` — Find activities
- ✅ `GET /health` — Health check
- ✅ CORS middleware enabled (configured for `*`)

### Schemas
- ✅ `LocationRequest` with latitude/longitude validation
- ✅ `LocationResponse` with results array
- ✅ `SearchResult` with title, url, snippet

### Configuration
- ✅ FastAPI app
- ✅ CORS enabled
- ✅ Pydantic validation

---

## 2. Integration Points

### Request/Response Contract

**Expected Request**:
```json
{
  "latitude": 37.7749,
  "longitude": -122.4194
}
```

**Expected Response**:
```json
{
  "results": [
    {
      "title": "String",
      "url": "https://example.com",
      "snippet": "Short description..."
    }
  ]
}
```

**Validation Checks** (Client will verify):
- ✅ Content-Type: application/json
- ✅ Response follows LocationResponse schema
- ✅ All fields present and correctly typed
- ✅ Status code 200 on success
- ✅ Status code 422 on validation error

---

## 3. Required Fixes / Verification

Before client integration can proceed, verify:

### ✅ Health Endpoint
```bash
curl http://localhost:8000/health
# Expected: 200 OK { "status": "ok" }
```

**Status**: Ready

---

### ⚠️  CORS Headers

**What to Verify**:
```bash
curl -i -X OPTIONS http://localhost:8000/location/group \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST"
```

**Expected Headers**:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: POST, GET, OPTIONS
Access-Control-Allow-Headers: *
```

**Current Status in Code**: ✅ 
```python
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["POST", "GET", "OPTIONS"],
    allow_headers=["*"],
)
```

**Action**: Verify at runtime with curl above.

---

### ⚠️  Request Validation

**Test Case 1: Valid Request**
```bash
curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}'
```

**Expected**: 200 OK with valid response

**Test Case 2: Invalid Latitude**
```bash
curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 91, "longitude": 0}'
```

**Expected**: 422 with error message about invalid latitude

**Test Case 3: Missing Field**
```bash
curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749}'
```

**Expected**: 400 with error about missing longitude

**Current Status**: ✅ Pydantic validation in place

---

### ⚠️  Response Schema Consistency

**Verify**: All 3 endpoints return same schema format.

```bash
# All three should return same structure:
curl -X POST http://localhost:8000/location/group -d {...}
curl -X POST http://localhost:8000/location/places -d {...}
curl -X POST http://localhost:8000/location/activities -d {...}

# All should have: { results: [{ title, url, snippet }, ...] }
```

**Current Status**: ✅ In code

---

### ⚠️  Empty Results Handling

**Test**: Search location with no results

```bash
curl -X POST http://localhost:8000/location/activities \
  -H "Content-Type: application/json" \
  -d '{"latitude": 0, "longitude": 0}'  # International waters
```

**Expected**: 200 OK with `{ "results": [] }`  
**Not**: 404 or error

**Current Status**: ⚠️  **VERIFY** — Ensure backend returns 200, not error

**Fix if Needed**: In `internal/service/itinerary.py`, ensure:
```python
# Should return empty results, not error
return LocationResponse(results=[])
```

---

### ⚠️  Performance: Response Time

**Test**: Measure response time for each endpoint

```bash
time curl -X POST http://localhost:8000/location/group \
  -d '{"latitude": 37.7749, "longitude": -122.4194}'
```

**Expected**: < 5 seconds  
**Target**: < 2 seconds (ideal)

**Current Status**: ⚠️  **VERIFY** — Depends on SearXNG performance

**If Slow**:
- Check SearXNG at localhost:8888
- Check Ollama agent at localhost:11434
- May need timeout handling (client-side timeout: 5s)

---

### ⚠️  Error Messages Format

**Test**: Trigger validation error and check response body

```bash
curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 200, "longitude": 0}' \
  | jq .
```

**Expected**: Meaningful error message (not stack trace)

**Example Bad Response**:
```json
{
  "detail": "Internal server error"
}
```

**Example Good Response**:
```json
{
  "detail": [
    {
      "type": "value_error",
      "loc": ["body", "latitude"],
      "msg": "latitude must be between -90 and 90"
    }
  ]
}
```

**Current Status**: ✅ Pydantic provides this by default

---

### 🔧  Potential Fixes Needed

If any of the above fail, here are the likely fixes:

#### Fix 1: Empty Results Not Returning 200
**File**: `internal/service/itinerary.py`

**Current** (likely):
```python
def search_locations(...):
    results = await search_locations(...)
    if not results:
        raise HTTPException(status_code=404, detail="No results found")
    return LocationResponse(results=...)
```

**Change To**:
```python
def search_locations(...):
    results = await search_locations(...)
    # Always return 200, even if results are empty
    return LocationResponse(results=results or [])
```

---

#### Fix 2: CORS Not Working
**File**: `main.py`

**Verify**:
```python
from fastapi.middleware.cors import CORSMiddleware

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["POST", "GET", "OPTIONS"],
    allow_headers=["*"],
)
```

**If missing**, add above `app.include_router(router)`.

---

#### Fix 3: Response Schema Mismatch
**File**: `internal/schemas/itinerary.py`

**Current**:
```python
class SearchResult(BaseModel):
    title: str
    url: str
    snippet: str

class LocationResponse(BaseModel):
    results: list[SearchResult]
```

**Ensure**: All endpoints return exactly this structure.

If endpoints return extra fields (e.g., `category`, `rating`), client Zod validation will reject them. Either:
1. Strip extra fields in schema: `model_config = ConfigDict(extra="ignore")`
2. Add those fields to SearchResult schema

---

## 4. Audit Checklist

Run these before starting E02 (API Client development):

```bash
# 1. Health check
curl http://localhost:8000/health
# Expected: {"status": "ok"} + 200

# 2. CORS preflight
curl -i -X OPTIONS http://localhost:8000/location/group \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST"
# Expected: 200 with CORS headers

# 3. Valid request (group tours)
curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}' | jq .
# Expected: 200 with results array

# 4. Valid request (places)
curl -X POST http://localhost:8000/location/places \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}' | jq .
# Expected: 200 with results array

# 5. Valid request (activities)
curl -X POST http://localhost:8000/location/activities \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}' | jq .
# Expected: 200 with results array

# 6. Invalid latitude
curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 91, "longitude": 0}'
# Expected: 422 with validation error

# 7. Empty results (optional)
curl -X POST http://localhost:8000/location/activities \
  -H "Content-Type: application/json" \
  -d '{"latitude": 0, "longitude": 0}'
# Expected: 200 with {"results": []}

# 8. Response time
time curl -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}'
# Expected: < 5 seconds
```

---

## 5. Postman Collection

During E05 (Integration Audit), create a Postman collection with these tests. Template:

```json
{
  "info": {
    "name": "find-my-trip-service",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "POST /location/group (valid)",
      "request": {
        "method": "POST",
        "url": "http://localhost:8000/location/group",
        "body": {
          "raw": "{\"latitude\": 37.7749, \"longitude\": -122.4194}"
        }
      },
      "tests": [
        "pm.expect(pm.response.code).to.equal(200)",
        "pm.expect(pm.response.json().results).to.be.an('array')"
      ]
    },
    {
      "name": "POST /location/group (invalid lat)",
      "request": {
        "method": "POST",
        "url": "http://localhost:8000/location/group",
        "body": {
          "raw": "{\"latitude\": 91, \"longitude\": 0}"
        }
      },
      "tests": [
        "pm.expect(pm.response.code).to.equal(422)"
      ]
    }
  ]
}
```

---

## 6. Sign-Off

**Before Starting E02 (API Client)**:
- [ ] Health endpoint returns 200
- [ ] CORS headers present in response
- [ ] All 3 endpoints accessible with valid request
- [ ] Invalid coordinates return 422
- [ ] Empty results return 200 (not 404)
- [ ] Response time < 5 seconds
- [ ] Response schema matches LocationResponse

**If All Pass**: ✅ Backend ready for client integration

**If Any Fail**: 🔧 Apply fixes above, then re-test

---

## 7. Quick Commands

```bash
# Start backend
cd /Users/prathm/services/find-my-trip
python main.py
# Should see: Uvicorn running on http://0.0.0.0:8000

# Run all audit checks in one script
cat > /tmp/audit-backend.sh << 'EOF'
#!/bin/bash
set -e

echo "🔍 Backend Audit: find-my-trip"
echo ""

echo "1️⃣  Health check..."
curl -s http://localhost:8000/health | jq . || echo "FAIL: Health check"

echo ""
echo "2️⃣  CORS preflight..."
curl -s -i -X OPTIONS http://localhost:8000/location/group \
  -H "Origin: http://localhost:5173" | grep -i "access-control" || echo "FAIL: CORS headers"

echo ""
echo "3️⃣  Valid request (group)..."
curl -s -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}' | jq '.results | length' || echo "FAIL"

echo ""
echo "4️⃣  Valid request (places)..."
curl -s -X POST http://localhost:8000/location/places \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}' | jq '.results | length' || echo "FAIL"

echo ""
echo "5️⃣  Valid request (activities)..."
curl -s -X POST http://localhost:8000/location/activities \
  -H "Content-Type: application/json" \
  -d '{"latitude": 37.7749, "longitude": -122.4194}' | jq '.results | length' || echo "FAIL"

echo ""
echo "6️⃣  Invalid latitude (should be 422)..."
curl -s -w "\nStatus: %{http_code}\n" -X POST http://localhost:8000/location/group \
  -H "Content-Type: application/json" \
  -d '{"latitude": 91, "longitude": 0}' | grep "422" || echo "FAIL"

echo ""
echo "✅ Audit complete!"
EOF

chmod +x /tmp/audit-backend.sh
/tmp/audit-backend.sh
```

---

**Backend Audit Ready**. Run before E02 development begins.
