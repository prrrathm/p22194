# find-my-trip-client: Quick Reference Card

**Save this for easy access during development!**

---

## 🚀 Start Development

```bash
cd /Users/prathm/Documents/p22194

# 1. Verify backend is running
curl http://localhost:8000/health

# 2. Install dependencies
npm install maplibre-gl zod

# 3. Set active epic
bash scripts/bp-yaml-set.sh active_flow build_epic
bash scripts/bp-yaml-set.sh active_epic e02

# 4. Start build-epic
build-epic
```

---

## 📍 Build Progression

```
E02 (API Client) 👈 START HERE
  ├─ s01: Fetch wrapper
  ├─ s02: Error boundary
  └─ s03: Zod validation
       ↓ (when E02 done)
E01 (Map UI)
  ├─ s01: MapLibre setup
  ├─ s02: Frosted glass panel
  └─ s03: Layout integration
       ↓
E03 (Map Click)
  ├─ s01: Click handler
  ├─ s02: Coordinate extraction
  └─ s03: API orchestration
       ↓
E04 (Results Display)
  ├─ s01: Result card
  ├─ s02: Results panel
  └─ s03: Loading states
       ↓
E05 (Integration Audit) 👈 FINAL GATE
  ├─ s01: Postman audit
  ├─ s02: Schema contract
  └─ s03: Error scenarios
```

---

## 📂 Key Files to Know

| File | Purpose |
|------|---------|
| `specs/IMPLEMENTATION_PLAN_SUMMARY.md` | Full plan (read if confused) |
| `specs/state.yaml` | Current execution state (build-epic updates this) |
| `specs/epics/e02-api-client-error-handling/epic.yaml` | Current epic details |
| `specs/execution-status.yaml` | Story status tracker |
| `specs/BACKEND_AUDIT.md` | Backend verification checklist |
| `specs/verifications/INTEGRATION_AUDIT_PLAN.md` | How to audit client-service integration |
| `.env` | API base URL (set to localhost:8000) |

---

## 🧪 TDD Cycle (During Step 4: develop-tdd)

```
1. Write failing test (red)
2. Write minimal code to pass (green)
3. Refactor without breaking test (refactor)
4. Repeat for each task
```

**Test Command**:
```bash
npm run test -- app/api/__tests__/client.test.ts
# or
npm run test  # Run all
```

---

## ✅ Step Checklist (build-epic)

Each epic runs 9 steps:

```
Step 0: security-review ........... [Input: threat model artifacts]
Step 1: survey-context ............ [Verify epic scope]
Step 2: plan-work ................. [Detail tasks + acceptance]
Step 3: kickoff-branch ............ [Create feature branch]
Step 4: develop-tdd ............... [Red-green-refactor loop]
Step 5: verify-work ............... [UAT + mechanical checks]
Step 6: audit-code ................ [Quality gate - MUST PASS]
Step 7: commit-message ............ [Conventional Commits]
Step 8: release-branch ............ [PR/land]
```

**At each step, you approve before advancing.**

---

## 🛠 Common Commands

```bash
# Check status
cat specs/state.yaml | grep epic_cycle -A10

# View current epic
cat specs/epics/e$(cat specs/state.yaml | grep active_epic | cut -d: -f2 | tr -d ' ')-*/epic.yaml | head -30

# Run tests
npm run test

# Type check
npm run typecheck

# Build
npm run build

# Start dev server
npm run dev
```

---

## 🔥 E02 (API Client) Key Tasks

This is the first epic you'll build. Focus on these files:

| File | What to Create |
|------|---|
| `app/api/client.ts` | HttpClient class with fetch wrapper |
| `app/api/errors.ts` | ApiError class + retry logic |
| `app/api/schemas.ts` | Zod validators (LocationRequest, LocationResponse) |
| `app/components/ErrorBoundary.tsx` | React error boundary |
| `app/api/__tests__/client.test.ts` | Unit tests |
| `app/api/__tests__/errors.test.ts` | Error handling tests |
| `app/api/__tests__/schemas.test.ts` | Schema validation tests |

---

## ⚠️ Error Handling Checklist (E02)

Must handle these 5 scenarios:

- [ ] Network error (no internet)
  → User message: "Unable to connect to server. Check your connection."
  → Retry button: YES

- [ ] Timeout (>5s)
  → User message: "Request took too long. Please try again."
  → Retry button: YES (with exponential backoff)

- [ ] Malformed response (invalid JSON)
  → User message: "Unexpected data format from server."
  → Retry button: YES

- [ ] Empty results (200 OK, results=[])
  → User message: "No results found for this location."
  → Retry button: NO (not an error)

- [ ] Invalid coordinates
  → Validation BEFORE API call (client-side)
  → Error message: "Latitude must be between -90 and 90."

---

## 🎨 UI/UX Guidelines (E01, E04)

**Frosted Glass Style**:
```css
/* TailwindCSS classes */
bg-white/70              /* 70% opaque white */
backdrop-blur-lg         /* Blur effect */
border border-white/50   /* Subtle white border */
shadow-lg                /* Subtle shadow */
rounded-2xl              /* Apple-like radius */
```

**Layout** (E01):
```
┌─────────────────────────────┐
│         Header              │
├──────────────┬──────────────┤
│   Map (60%)  │  Panel (40%) │
│              │              │
│              │ Results here │
│              │  (frosted)   │
│              │              │
└──────────────┴──────────────┘
```

---

## 🔗 API Contract (For Reference)

**Request**: POST /location/{group|places|activities}
```json
{
  "latitude": number (-90 to 90),
  "longitude": number (-180 to 180)
}
```

**Response**: 200 OK
```json
{
  "results": [
    {
      "title": "string",
      "url": "string",
      "snippet": "string"
    }
  ]
}
```

**Errors**: 
- 422: Invalid coordinates
- 400: Missing fields
- 500: Server error

---

## 📊 Progress Tracking

After each epic, update execution-status.yaml:
```yaml
e01_map_frosted_panel:
  status: done  # (was pending)
  s01_map_setup:
    status: done
  s02_frosted_glass_panel:
    status: done
  s03_layout_integration:
    status: done
```

Or run:
```bash
bash scripts/sync-status-from-epics.sh
```

---

## 🚨 If Stuck

1. **Can't run build-epic?**
   → `bash scripts/bp-yaml-set.sh active_epic e02` (set state)

2. **Tests failing?**
   → `npm run test -- --verbose` (see details)

3. **TypeScript errors?**
   → `npm run typecheck` (check types)

4. **Backend issues?**
   → `curl http://localhost:8000/health` (verify running)
   → Read `specs/BACKEND_AUDIT.md`

5. **Confused about task?**
   → `cat specs/epics/e0X-*/epic.yaml` (see detailed spec)

---

## 📞 Get Help

- **Full Plan**: `cat specs/IMPLEMENTATION_PLAN_SUMMARY.md`
- **Audit Plan**: `cat specs/BACKEND_AUDIT.md`
- **Current Epic**: `cat specs/epics/$(cat specs/state.yaml | grep active_epic | cut -d: -f2 | tr -d ' ')-*/epic.yaml`
- **Integration Audit**: `cat specs/verifications/INTEGRATION_AUDIT_PLAN.md`

---

## ✨ You're Ready!

```
E02 is ready to build. Run:
  build-epic
```

Good luck! 🚀
