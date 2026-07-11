#!/usr/bin/env bash
set -euo pipefail

# Auto-source .env.local if PROJECT_ID isn't set
if [ -z "${GCP_PROJECT_ID:-}" ] && [ -f .env.local ]; then
  echo ">>> Sourcing .env.local..."
  set +euo pipefail
  source .env.local
  set -euo pipefail
fi

# ─── Configuration ───────────────────────────────────────────────────────────
PROJECT_ID="${GCP_PROJECT_ID:?Set GCP_PROJECT_ID (source .env.local first)}"
REGION="${GCP_REGION:-us-central1}"
SERVICES=("searxng" "users" "god-svc" "find-my-trip" "api-gateway")

echo "=== p22194 Cloud Run Deployment ==="
echo "Project: $PROJECT_ID"
echo "Region:  $REGION"
echo ""

# ─── Step 1: Enable APIs ────────────────────────────────────────────────────
echo ">>> Enabling required GCP APIs..."
gcloud services enable \
  run.googleapis.com \
  containerregistry.googleapis.com \
  secretmanager.googleapis.com \
  cloudbuild.googleapis.com \
  --project="$PROJECT_ID"

# ─── Step 2: Create Secret Manager secrets ──────────────────────────────────
echo ">>> Creating Secret Manager secrets (skipping if they exist)..."

create_secret() {
  local name="$1"
  local value="$2"
  if gcloud secrets describe "$name" --project="$PROJECT_ID" &>/dev/null; then
    echo "    Secret '$name' already exists, adding new version..."
    echo -n "$value" | gcloud secrets versions add "$name" --project="$PROJECT_ID" --data-file=-
  else
    echo "    Creating secret '$name'..."
    echo -n "$value" | gcloud secrets create "$name" --project="$PROJECT_ID" --data-file=- --replication-policy="automatic"
  fi
}

# Generate a JWT secret if not provided
JWT_SECRET="${JWT_SECRET:-$(openssl rand -hex 32)}"
SEARXNG_SECRET="${SEARXNG_SECRET:-$(openssl rand -hex 24)}"

create_secret "jwt-secret" "$JWT_SECRET"
create_secret "searxng-secret" "$SEARXNG_SECRET"
create_secret "mongo-uri" "${MONGO_URI:?Set MONGO_URI (MongoDB Atlas connection string)}"
create_secret "agent-url" "${AGENT_URL:-}"
create_secret "agent-api-key" "${AGENT_API_KEY:-}"

# ─── Step 3: Build all container images ─────────────────────────────────────
echo ">>> Building container images with Cloud Build..."
gcloud builds submit --config cloudbuild.yaml --project="$PROJECT_ID" .

# ─── Step 4: Deploy services (in dependency order) ──────────────────────────
deploy_service() {
  local name="$1"
  local yaml="cloudrun/${name}.yaml"

  echo ">>> Deploying $name..."

  # Replace PROJECT_ID in the yaml
  local tmpyaml="/tmp/cloudrun-${name}.yaml"
  sed "s/PROJECT_ID/$PROJECT_ID/g" "$yaml" > "$tmpyaml"

  gcloud run services replace "$tmpyaml" \
    --region="$REGION" \
    --project="$PROJECT_ID"

  rm -f "$tmpyaml"
}

for svc in "${SERVICES[@]}"; do
  deploy_service "$svc"
done

# ─── Step 5: Create ConfigMap with service URLs ─────────────────────────────
echo ">>> Fetching service URLs..."

get_url() {
  gcloud run services describe "$1" --region="$REGION" --project="$PROJECT_ID" --format='value(status.url)'
}

SEARXNG_URL=$(get_url "searxng")
USERS_URL=$(get_url "users")
GODSVC_URL=$(get_url "god-svc")
FMT_URL=$(get_url "find-my-trip")
GW_URL=$(get_url "api-gateway")

echo ""
echo "=== Service URLs ==="
echo "searxng:      $SEARXNG_URL"
echo "users:        $USERS_URL"
echo "god-svc:      $GODSVC_URL"
echo "find-my-trip: $FMT_URL"
echo "api-gateway:  $GW_URL"
echo ""

# Update configmap with actual URLs
cat > /tmp/p22194-config.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: p22194-config
data:
  searxng-url: "$SEARXNG_URL"
  searxng-base-url: "$SEARXNG_URL"
  upstream-users-url: "$USERS_URL"
  upstream-find-my-trip-url: "$FMT_URL"
EOF

kubectl apply -f /tmp/p22194-config.yaml 2>/dev/null || \
  gcloud run services replace /tmp/p22194-config.yaml --region="$REGION" --project="$PROJECT_ID" 2>/dev/null || \
  echo ">>> NOTE: ConfigMap creation via kubectl failed. You may need to set config values manually."

# ─── Step 6: Re-deploy gateway with correct upstream URLs ───────────────────
echo ">>> Re-deploying api-gateway with upstream URLs..."
deploy_service "api-gateway"

GW_URL=$(get_url "api-gateway")

# ─── Step 7: Domain Mapping ────────────────────────────────────────────────
DOMAIN="${CUSTOM_DOMAIN:-api.prrrathm.com}"
echo ""
echo ">>> Setting up domain mapping: $DOMAIN → api-gateway"

# Verify domain ownership (one-time, harmless if already verified)
echo "    Verifying domain ownership..."
gcloud domains verify "${DOMAIN#*.}" --project="$PROJECT_ID" 2>/dev/null || \
  echo "    Domain already verified or needs manual verification."

# Create domain mapping
if gcloud run domain-mappings describe --domain "$DOMAIN" --region="$REGION" --project="$PROJECT_ID" &>/dev/null; then
  echo "    Domain mapping for $DOMAIN already exists."
else
  echo "    Creating domain mapping..."
  gcloud run domain-mappings create \
    --service api-gateway \
    --domain "$DOMAIN" \
    --region="$REGION" \
    --project="$PROJECT_ID"
fi

# Get DNS records to add
echo ""
echo ">>> DNS records to add in Cloudflare:"
echo "    ┌──────────────────────────────────────────────────────────┐"
echo "    │ Type  │ Name   │ Content                                │"
echo "    ├───────┼────────┼────────────────────────────────────────┤"

# Fetch the mapping details for DNS instructions
MAPPING_IP=$(gcloud run domain-mappings describe --domain "$DOMAIN" --region="$REGION" --project="$PROJECT_ID" --format='value(status.resourceRecords)' 2>/dev/null || echo "")
if echo "$MAPPING_IP" | grep -q "A "; then
  IP_ADDR=$(echo "$MAPPING_IP" | grep "A " | awk '{print $2}')
  echo "    │ A     │ api    │ $IP_ADDR"
else
  echo "    │ CNAME │ api    │ ghs.googlehosted.com                  │"
fi
echo "    └──────────────────────────────────────────────────────────┘"
echo ""
echo "    Add these in Cloudflare → DNS → Records for prrrathm.com"

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Public URL (after DNS setup): https://$DOMAIN"
echo "Temporary URL:                $GW_URL"
echo ""
echo "Service URLs (internal only):"
echo "  users:        $USERS_URL"
echo "  god-svc:      $GODSVC_URL"
echo "  find-my-trip: $FMT_URL"
echo "  searxng:      $SEARXNG_URL"
echo ""
echo "Next steps:"
echo "1. Add DNS records in Cloudflare (see above)"
echo "2. Wait 5-10 min for DNS propagation"
echo "3. Test: curl https://$DOMAIN/health/live"
echo "4. View logs: gcloud logging read 'resource.type=cloud_run_revision' --limit 50"
