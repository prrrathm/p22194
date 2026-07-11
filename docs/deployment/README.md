# Deployment Architecture

## System Overview

```
                         ┌──────────────────────┐
                         │   findmytrip.prrrathm │  (Vercel - React SPA)
                         └──────────┬───────────┘
                                    │ browser
                                    ▼
                    ┌───────────────────────────────┐
                    │  api.prrrathm.com              │
                    │  Cloudflare Worker             │
                    │  (CORS proxy + edge routing)   │
                    └──────────────┬────────────────┘
                                   │ HTTPS
                                   ▼
                    ┌──────────────────────────────┐
                    │  Cloud Run: api-gateway       │
                    │  [public, 0–5 instances]      │
                    │  Go reverse proxy, JWT auth,  │
                    │  rate limiting, /api/v1/*     │
                    └──┬──────┬──────┬──────┬──────┘
                       │      │      │      │
              ┌────────┘  ┌───┘  ┌───┘  ┌───┘
              ▼           ▼      ▼      ▼
          ┌────────┐ ┌───────┐ ┌─────────────┐ ┌─────────┐
          │ users  │ │god-svc│ │find-my-trip │ │ searxng │
          │ [int]  │ │ [int] │ │   [int]     │ │  [int]  │
          │ Go     │ │ Go    │ │ Python/LLM  │ │ Python  │
          └───┬────┘ └───┬───┘ └──────┬──────┘ └────┬────┘
              │          │            │              │
              ▼          ▼            ▼              ▼
         ┌──────────────────────────────────────────────┐
         │           MongoDB Atlas (p22194)              │
         │           asia-south2 (Delhi)                 │
         └──────────────────────────────────────────────┘
```

## Service Summary

| Service | Language | Ingress | Scale | CPU | Memory | Purpose |
|---------|----------|---------|-------|-----|--------|---------|
| api-gateway | Go | public | 0–5 | 1 | 512Mi | Reverse proxy, JWT auth, rate limiting, caching |
| users | Go | internal | 0–3 | 1 | 512Mi | User accounts, auth sessions |
| god-svc | Go | internal | 0–3 | 1 | 512Mi | God-mode admin service |
| find-my-trip | Python | internal | 0–3 | 1 | 1Gi | AI trip search (LLM agents + SearXNG) |
| searxng | Python | internal | 0–3 | 1 | 1Gi | Metasearch engine (self-hosted) |

All services run on **GCP Cloud Run** in `asia-south2` (Delhi), with `min-instances=0` (scale to zero) for cost savings.

## Request Flow

1. Browser sends request to `https://api.prrrathm.com/...`
2. DNS resolves via Cloudflare (CNAME to Cloudflare-proxied domain)
3. **Cloudflare Worker** (`p22194-api-proxy`) intercepts the request:
   - Handles CORS preflight (`OPTIONS` → 204 with CORS headers)
   - Proxies all other requests to the Cloud Run API gateway
   - Adds `Access-Control-Allow-Origin: *` to all responses
4. **API Gateway** (`api-gateway`) receives the request:
   - Routes by path prefix: `/api/v1/{service}/*`
   - Public services (e.g. find-my-trip) are proxied directly
   - Protected services require a valid JWT in the `Authorization` header
   - Strips the `/api/v1/{service}` prefix before forwarding
5. The upstream service processes the request and returns a response
6. The response flows back through the gateway → Cloudflare Worker → browser

## Secrets Management

Secrets are stored in **GCP Secret Manager** and injected as env vars at deploy time:

| Secret | Used By | Purpose |
|--------|---------|---------|
| `jwt-secret` | api-gateway, users, god-svc | JWT signing key |
| `mongo-uri` | users, god-svc, find-my-trip | MongoDB Atlas connection string |
| `searxng-secret` | searxng | SearXNG secret key |
| `agent-url` | find-my-trip | LLM API base URL (e.g. Groq) |
| `agent-api-key` | find-my-trip | LLM API key |

## Domain & DNS

| Domain | Points To | Purpose |
|--------|-----------|---------|
| `api.prrrathm.com` | Cloudflare Worker → Cloud Run api-gateway | API endpoint |
| `findmytrip.prrrathm.com` | Vercel | Frontend SPA |

**Important:** The `api.prrrathm.com` CNAME must point to the Cloudflare Worker (not directly to Cloud Run). This is configured in Cloudflare DNS as a proxied CNAME record.

## Cloud Run Networking (Critical)

Cloud Run services communicate via their public `.run.app` URLs. **The `Host` header matters.** When the api-gateway proxies a request to an upstream service, it must set both:
- `req.URL.Host` — for the TCP connection (which service to connect to)
- `req.Host` — for the HTTP Host header (which service Cloud Run routes to)

If only `req.URL.Host` is set (the default `httputil.ReverseProxy` behavior), Cloud Run routes the request back to the gateway based on the original `Host` header, causing 404 errors. The custom `Director` in `api-gateway/internal/proxy/proxy.go` handles this correctly.

## Budget

GCP billing alert set at **500 INR/month** with notifications at 50%, 80%, and 100%. All services scale to zero when idle.

---

# Deployment Commands Cheatsheet

## Full Initial Deployment

```bash
# 1. Set up environment
cp .env.local.example .env.local
# Edit .env.local with real values

# 2. Run the full deployment script
./deploy.sh
```

## Redeploy a Single Service (after code changes)

```bash
# Build only the changed service's image
# Option A: Rebuild everything (slow but safe)
gcloud builds submit --config=cloudbuild.yaml --project=p22194 .

# Option B: Rebuild only the gateway (fast)
gcloud builds submit --config=cloudbuild-gateway.yaml --project=p22194 .
```

```bash
# Deploy the specific service
gcloud run deploy <service-name> \
  --image gcr.io/p22194/<service-name> \
  --region=asia-south2 \
  --project=p22194 \
  --platform=managed \
  --quiet
```

Service names: `api-gateway`, `users`, `god-svc`, `find-my-trip`, `searxng`

### Full Service Redeploy Examples

```bash
# find-my-trip
gcloud builds submit --config=cloudbuild.yaml --project=p22194 .
gcloud run deploy find-my-trip \
  --image gcr.io/p22194/find-my-trip \
  --region=asia-south2 --project=p22194 --quiet \
  --set-secrets="AGENT_URL=agent-url:latest,AGENT_API_KEY=agent-api-key:latest,MONGO_URI=mongo-uri:latest,USERS_JWT_SECRET=jwt-secret:latest" \
  --set-env-vars="SEARCH_XNG_URL=https://searxng-3b3mud4qva-em.a.run.app,AGENT_MODEL=llama3.2" \
  --min-instances=0 --max-instances=3

# users
gcloud run deploy users \
  --image gcr.io/p22194/users \
  --region=asia-south2 --project=p22194 --quiet \
  --set-secrets="USERS_MONGO_URI=mongo-uri:latest,USERS_JWT_SECRET=jwt-secret:latest" \
  --min-instances=0 --max-instances=3

# god-svc
gcloud run deploy god-svc \
  --image gcr.io/p22194/god-svc \
  --region=asia-south2 --project=p22194 --quiet \
  --set-secrets="GOD_MONGO_URI=mongo-uri:latest" \
  --min-instances=0 --max-instances=3

# searxng
gcloud run deploy searxng \
  --image gcr.io/p22194/searxng \
  --region=asia-south2 --project=p22194 --quiet \
  --set-secrets="SEARXNG_SECRET=searxng-secret:latest" \
  --min-instances=0 --max-instances=3

# api-gateway
gcloud run deploy api-gateway \
  --image gcr.io/p22194/api-gateway \
  --region=asia-south2 --project=p22194 --quiet \
  --set-env-vars="GATEWAY_SERVER_ADDR=:8080,GATEWAY_LOG_FORMAT=pretty,GATEWAY_LOG_LEVEL=info,GATEWAY_UPSTREAM_USERS_URL=https://users-3b3mud4qva-em.a.run.app,GATEWAY_UPSTREAM_FIND_MY_TRIP_URL=https://find-my-trip-3b3mud4qva-em.a.run.app,GATEWAY_UPSTREAM_GOD_SVC_URL=https://god-svc-3b3mud4qva-em.a.run.app" \
  --set-secrets="GATEWAY_AUTH_JWT_SECRET=jwt-secret:latest" \
  --min-instances=0 --max-instances=10
```

**Note:** When updating env vars on a service, use `--update-env-vars` for non-secret changes (avoids re-deploying the container):

```bash
# Update only env vars (no image rebuild)
gcloud run services update api-gateway \
  --region=asia-south2 --project=p22194 \
  --update-env-vars="GATEWAY_LOG_FORMAT=pretty"
```

## Route Traffic to Latest Revision

```bash
gcloud run services update-traffic <service-name> \
  --region=asia-south2 --project=p22194 --to-latest
```

## View Logs

```bash
# Pretty logs for a specific service
gcloud run services logs read <service-name> \
  --region=asia-south2 --project=p22194 --limit=50

# Structured search via Cloud Logging
gcloud logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=<service-name> AND resource.labels.location=asia-south2" \
  --project=p22194 --limit=50

# Tail logs in real time
gcloud logging read \
  "resource.type=cloud_run_revision AND resource.labels.service_name=<service-name>" \
  --project=p22194 --limit=5 --follow
```

## Check Service Status

```bash
# List all services and their URLs
gcloud run services list --region=asia-south2 --project=p22194

# Describe a specific service (env vars, revisions, traffic)
gcloud run services describe <service-name> \
  --region=asia-south2 --project=p22194

# Check which revision is live
gcloud run services describe <service-name> \
  --region=asia-south2 --project=p22194 \
  --format="yaml(status.traffic)"
```

## Cloudflare Worker

```bash
# Deploy/update the CORS proxy worker
cd cloudflare-worker
wrangler deploy

# View worker logs
wrangler tail p22194-api-proxy
```

## Secret Management

```bash
# List all secrets
gcloud secrets list --project=p22194

# Add a new version of a secret
echo -n "new-value" | gcloud secrets versions add <secret-name> \
  --project=p22194 --data-file=-

# Read a secret value (for debugging)
gcloud secrets versions access latest --secret=<secret-name> --project=p22194

# Grant a service account access to secrets
gcloud secrets add-iam-policy-binding <secret-name> \
  --project=p22194 \
  --member="serviceAccount:766511188075-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## MongoDB Atlas

```bash
# Connect via mongosh (from local machine)
mongosh "mongodb+srv://p22194.9ir0eyh.mongodb.net/p22194" \
  --username dev --password <password>

# Reset the dev user password
atlas auth login --username prrrrrathm@gmail.com
atlas dbusers resetRootPassword --username dev --password <new-password>
```

## Vercel (Frontend)

```bash
# Set environment variable
cd .
vercel env add VITE_API_BASE_URL production
# Enter: https://api.prrrathm.com

# Deploy to production
vercel --prod --yes --archive=tgz

# Check env vars
vercel env ls
```

---

# Important Notes

## Cloud Run Region

The project uses `asia-south2` (Delhi). **Custom domain mappings are NOT supported in this region** — that's why we use the Cloudflare Worker proxy for `api.prrrathm.com` instead of Cloud Run's built-in domain mapping.

## `PORT` is Reserved

Cloud Run sets the `PORT` environment variable automatically. **Never** set it in YAML or env vars — it will be overwritten. All services listen on the port provided by Cloud Run (typically 8080).

## Go Workspace (`go.work`)

The `go.work` file references modules not present in each Docker build context. **All Dockerfiles must set `GOWORK=off`** to prevent Go from trying to resolve workspace references during builds.

## `httputil.ReverseProxy` Host Header Bug

Go's `httputil.NewSingleHostReverseProxy` only sets `req.URL.Host` (for the TCP connection), not `req.Host` (for the HTTP Host header). Cloud Run uses the Host header for routing. Without setting `req.Host`, requests get routed back to the sender.

The fix is in `api-gateway/internal/proxy/proxy.go`:
```go
rp.Director = func(req *http.Request) {
    req.URL.Scheme = target.Scheme
    req.URL.Host = target.Host
    req.Host = target.Host  // Critical for Cloud Run routing
}
```

## `.env.local` Security

`.env.local` contains production secrets (MongoDB URI, API keys, OIDC tokens). It is gitignored but should **never** be committed. Use `.env.local.example` as a template.

## Service Discovery

Upstream URLs for the api-gateway are passed as env vars (`GATEWAY_UPSTREAM_{SERVICE}_URL`). After deploying services, the gateway may need to be re-deployed to pick up new upstream URLs. The `config.go` loader converts env var underscores to hyphens: `GATEWAY_UPSTREAM_FIND_MY_TRIP_URL` → `find-my-trip`.

## Chi Router & Path Matching

Chi's wildcard routing (`/api/v1/*`) has issues matching deep nested paths. The gateway uses a `NotFound` handler to intercept all `/api/v1/*` requests and route them through the proxy, bypassing Chi's routing tree entirely.

## Cost Optimization

- All services use `min-instances=0` (scale to zero)
- GCP billing alert at 500 INR/month
- Scaled-down instances incur zero cost
- Cloud Build uses free tier (120 build-minutes/day)
