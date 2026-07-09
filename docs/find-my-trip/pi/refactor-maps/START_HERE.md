# 🚀 Quick Start: build-epic for E02 (API Client + Error Handling)

Your state is already configured! Here's what to do:

## Step 1: Start the Backend (New Terminal)

```bash
cd /Users/prathm/services/find-my-trip
python main.py
```

Expected output:
```
INFO:     Uvicorn running on http://0.0.0.0:8000 (Press CTRL+C to quit)
```

## Step 2: Verify Backend is Running

```bash
curl http://localhost:8000/health
```

Expected:
```json
{"status":"ok"}
```

## Step 3: In Another Terminal, Install Dependencies

```bash
cd /Users/prathm/Documents/p22194/clients/find-my-trip-client
npm install maplibre-gl zod
```

## Step 4: Start Build-Epic

```bash
cd /Users/prathm/Documents/p22194
build-epic
```

This will guide you through the 9-step cycle:
- Step 0: security-review (threat model)
- Step 1: survey-context (confirm scope)
- Step 2: plan-work (detail tasks)
- Step 3: kickoff-branch (create feature branch)
- **Step 4: develop-tdd** ⭐ (YOU BUILD HERE - red-green-refactor)
- Step 5: verify-work (UAT checks)
- Step 6: audit-code (quality gate - MUST PASS)
- Step 7: commit-message (Conventional Commits)
- Step 8: release-branch (PR/land)

## What You'll Build in Step 4 (develop-tdd)

For epic E02 (API Client + Error Handling), you'll create:

```
app/api/
├─ client.ts ................. HttpClient class (fetch wrapper)
├─ errors.ts ................. ApiError + retry logic  
├─ schemas.ts ................ Zod validators
└─ __tests__/
   ├─ client.test.ts ........ HttpClient tests
   ├─ errors.test.ts ........ Error tests
   └─ schemas.test.ts ....... Zod validation tests

app/components/
└─ ErrorBoundary.tsx ........ React error boundary
```

## The TDD Cycle During Step 4

For each task, develop-tdd will guide you:

1. **RED**: Write a test that describes what should happen
2. **GREEN**: Write minimal code to make the test pass
3. **REFACTOR**: Improve code without breaking tests
4. **REPEAT**: Move to next task

Example for first task:

```bash
# Write failing test (RED)
cat > app/api/__tests__/client.test.ts << 'TESTEOF'
import { HttpClient } from '../client'

describe('HttpClient', () => {
  it('should have base URL from environment', () => {
    const client = new HttpClient()
    expect(client.baseUrl).toBe('http://localhost:8000')
  })
})
TESTEOF

# Run test (will FAIL)
npm run test -- app/api/__tests__/client.test.ts

# Write code to make it pass (GREEN)
cat > app/api/client.ts << 'CODEEOF'
export class HttpClient {
  baseUrl: string

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000'
  }
}
CODEEOF

# Run test again (should PASS)
npm run test -- app/api/__tests__/client.test.ts

# Tell develop-tdd "test passes" and move to next task
```

## Quick Reference

**Check current status**:
```bash
cat specs/state.yaml | grep epic_cycle -A5
```

**View current epic tasks**:
```bash
cat specs/epics/e02-api-client-error-handling/epic.yaml | head -100
```

**Run tests**:
```bash
npm run test
```

**Type check**:
```bash
npm run typecheck
```

**Build**:
```bash
npm run build
```

## Terminal Layout (Recommended)

```
Terminal 1: Backend
  cd /Users/prathm/services/find-my-trip
  python main.py

Terminal 2: Development
  cd /Users/prathm/Documents/p22194
  build-epic
  # Follow prompts

Terminal 3: Commands (as needed)
  cd /Users/prathm/Documents/p22194/clients/find-my-trip-client
  npm run test
  npm run typecheck
```

## What Happens Next

1. **build-epic starts** → Answer brief questions for steps 0-3
2. **Step 4 starts** → develop-tdd guides you through TDD
3. **You write tests** → Then code → Then refactor
4. **Step 5 verifies** → Runs typecheck, tests, build
5. **Step 6 audits** → Code quality gate (must pass)
6. **Step 7-8** → Commits and PRs

## Stuck?

- **Don't know what to build?** → `cat specs/epics/e02-api-client-error-handling/epic.yaml`
- **Tests failing?** → `npm run test -- --verbose`
- **Backend issues?** → `cat specs/BACKEND_AUDIT.md`
- **Need quick ref?** → `cat QUICK_REFERENCE.md`

## State Already Configured

Your specs/state.yaml is already set to:
```yaml
active_flow: build_epic
active_epic: e02
```

So you just need to:
1. Start backend
2. Install deps
3. Run `build-epic`

Good luck! 🚀
