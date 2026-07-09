# During build-epic Execution

**You're now running build-epic. Here's what to expect and how to respond.**

---

## Step 0: Security Review

**What build-epic will ask:**
```
Threat model for epic E02: API Client + Error Handling

Identify security concerns:
- API client making HTTP requests
- Error handling with user messages
- Environment variables for configuration
- Schema validation on responses

Any critical threats? (yes/no):
```

**Your response:** `no`

Why: Simple fetch wrapper with validation handles standard HTTP security.

---

## Step 1: Survey Context

**What build-epic will ask:**
```
Epic: e02-api-client-error-handling

Acceptance Criteria:
✓ Fetch wrapper handles timeouts, retries, validation
✓ Zod schemas validate responses
✓ Error boundary displays messages
✓ Unit tests pass
✓ TypeScript strict mode passes

Scope confirmed? (yes/no):
```

**Your response:** `yes`

---

## Step 2: Plan Work

**What build-epic will ask:**
```
Stories in E02:
  s01: Fetch Wrapper + Environment Configuration (8 BCP)
  s02: Error Boundary + Retry Logic (7 BCP)
  s03: Schema Validation with Zod (6 BCP)

Tasks detailed and ready? (yes/no):
```

**Your response:** `yes`

---

## Step 3: Kickoff Branch

**What build-epic will do:**
```
Creating feature branch: feature/e02-api-client-error-handling
Verifying clean test baseline...
✓ Branch created
✓ All tests passing
✓ Ready to develop
```

**Your action:** Just watch. It creates the branch automatically.

---

## Step 4: Develop-TDD (YOU BUILD HERE)

**This is the longest step (4-6 hours).**

### Task 1: Create app/api/client.ts

**develop-tdd shows:**
```
═══════════════════════════════════════════════════════════
TASK: s01t01 - Create app/api/client.ts

Acceptance Criteria:
  ✓ Fetch wrapper exported
  ✓ Base URL from VITE_API_BASE_URL
  ✓ 5-second timeout (AbortController)
  ✓ TypeScript strict mode passes

Next: Write a failing test describing HttpClient
═══════════════════════════════════════════════════════════
```

### Your Workflow:

```bash
# 1. Create test file (FAILS - RED)
cat > app/api/__tests__/client.test.ts << 'EOF'
import { HttpClient } from '../client'

describe('HttpClient', () => {
  it('should export HttpClient class', () => {
    expect(HttpClient).toBeDefined()
  })

  it('should have baseUrl from environment', () => {
    const client = new HttpClient()
    expect(client.baseUrl).toBe('http://localhost:8000')
  })

  it('should have request method', () => {
    const client = new HttpClient()
    expect(typeof client.request).toBe('function')
  })
})
EOF

# 2. Run test (RED - fails)
npm run test -- app/api/__tests__/client.test.ts

# 3. Write minimal code (GREEN)
cat > app/api/client.ts << 'EOF'
export class HttpClient {
  baseUrl: string

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000'
  }

  async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000)

    try {
      const response = await fetch(`${this.baseUrl}${path}`, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      return await response.json()
    } finally {
      clearTimeout(timeoutId)
    }
  }
}
EOF

# 4. Run test (GREEN - passes)
npm run test -- app/api/__tests__/client.test.ts

# 5. Tell develop-tdd: test passes
# develop-tdd will ask: "Is test passing?" → YES
# Then move to next task
```

### Task 2: Create app/api/errors.ts

Same process:
1. Write failing test
2. Run test (RED)
3. Write code
4. Run test (GREEN)
5. Tell develop-tdd "test passes"

**Example test:**
```typescript
import { ApiError } from '../errors'

describe('ApiError', () => {
  it('should create error with message', () => {
    const error = new ApiError('Network failed', 0, true)
    expect(error.message).toBe('Network failed')
    expect(error.retryable).toBe(true)
  })

  it('should mark 5xx as retryable', () => {
    const error = new ApiError('Server error', 500, true)
    expect(error.retryable).toBe(true)
  })

  it('should mark 4xx as non-retryable', () => {
    const error = new ApiError('Bad request', 400, false)
    expect(error.retryable).toBe(false)
  })
})
```

### Task 3: Create app/api/schemas.ts

Same TDD cycle for Zod validators:

```typescript
import { z } from 'zod'

export const LocationRequestSchema = z.object({
  latitude: z.number().min(-90).max(90),
  longitude: z.number().min(-180).max(180),
})

export const SearchResultSchema = z.object({
  title: z.string(),
  url: z.string().url(),
  snippet: z.string(),
})

export const LocationResponseSchema = z.object({
  results: z.array(SearchResultSchema),
})

export type LocationRequest = z.infer<typeof LocationRequestSchema>
export type SearchResult = z.infer<typeof SearchResultSchema>
export type LocationResponse = z.infer<typeof LocationResponseSchema>
```

### Task 4: Create app/components/ErrorBoundary.tsx

React component for error handling:

```typescript
import { ReactNode } from 'react'

interface Props {
  children: ReactNode
  onRetry?: () => void
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error) {
    console.error('Error caught:', error)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="bg-white/70 backdrop-blur-lg p-6 rounded-2xl border border-white/50">
          <p className="text-red-600">{this.state.error?.message}</p>
          {this.props.onRetry && (
            <button onClick={this.props.onRetry} className="mt-4 px-4 py-2 bg-blue-500 text-white rounded">
              Retry
            </button>
          )}
        </div>
      )
    }

    return this.props.children
  }
}
```

### Tasks 5-8: Unit Tests

Write tests for:
- HttpClient timeout handling
- HttpClient retry logic
- Zod validation
- Error boundary rendering

Same TDD cycle for each.

---

## Step 5: Verify Work

**What build-epic does:**
```
Running verification checklist:

  npm run typecheck ........... ✓ PASS
  npm run test ................ ✓ PASS (all tests)
  npm run build ............... ✓ PASS

Result: ✓ All checks passed
```

If anything fails:
```
FAIL: npm run typecheck

Error: Property 'request' has no initializer...
```

**Fix:** Go back to your code, fix the error, run tests again.

---

## Step 6: Audit Code (MUST PASS)

**What build-epic does:**
```
Running audit-code checklist:

  ✓ CONVENTIONS.md compliance
  ✓ Boy Scout Rule (leave code cleaner)
  ✓ Test coverage >80%
  ✓ TypeScript strict mode
  ✓ No console.error in production code

Result: ✓ PASS - All checks passed
```

If it fails, fix and go back to step 4 (develop-tdd).

---

## Step 7: Commit Message

**What build-epic does:**
```
Drafting Conventional Commits message:

Title:
  feat(api): implement HttpClient with error handling

Body:
  - Add HttpClient fetch wrapper with 5s timeout
  - Implement ApiError with retry logic
  - Add Zod schema validation
  - Create React ErrorBoundary component
  - Add unit tests covering all error scenarios

Semantic Release Version Bump: minor (feat = minor)

Approve? (yes/no):
```

**Your response:** `yes`

---

## Step 8: Release Branch

**What build-epic does:**
```
Creating PR with gh (GitHub CLI)

Branch: feature/e02-api-client-error-handling
Title: feat(api): implement HttpClient with error handling

✓ PR created: github.com/yourrepo/pull/123
```

If you don't have GitHub CLI, it will save the details for you to create manually.

---

## Quick Reference During Development

### When Tests Fail

```bash
npm run test -- --verbose
# Shows detailed error output
```

### When TypeScript Complains

```bash
npm run typecheck
# Shows all type errors with line numbers
```

### When You're Confused

```bash
cat specs/epics/e02-api-client-error-handling/epic.yaml | head -50
# See exact task details
```

### When You Forget Commands

```bash
cat QUICK_REFERENCE.md
# All commands in one place
```

---

## Common Issues & Fixes

### "Test won't run"
- Make sure `npm install maplibre-gl zod` completed
- Check that `package.json` has testing setup
- Run: `npm run test -- --version`

### "TypeScript errors"
- Run: `npm run typecheck` to see all errors
- Read the error message carefully
- Fix the code, then re-run

### "Fetch not working"
- Verify backend is running: `curl http://localhost:8000/health`
- Check `.env` has `VITE_API_BASE_URL=http://localhost:8000`
- Ensure Zod is installed: `npm list zod`

### "build-epic won't continue"
- Read the error message carefully
- Fix the issue (usually a test failure)
- Tell develop-tdd to continue or re-run the step

---

## You've Got This!

Just follow the TDD cycle:
1. **RED**: Write test
2. **GREEN**: Write code
3. **REFACTOR**: Clean up
4. **REPEAT**: Next task

develop-tdd guides you every step. Let's go! 🚀
