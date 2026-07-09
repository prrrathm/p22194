#!/bin/bash
# start-epic.sh - Quick start for build-epic
# Usage: bash scripts/start-epic.sh [epic-id]
# Example: bash scripts/start-epic.sh e02

EPIC_ID="${1:-e02}"

echo "🚀 Starting build-epic for $EPIC_ID"
echo ""

# Verify epic exists
EPIC_DIR="specs/epics/${EPIC_ID}-"*
FOUND=0
for dir in $EPIC_DIR; do
  if [ -f "$dir/epic.yaml" ]; then
    FOUND=1
    break
  fi
done

if [ $FOUND -eq 0 ]; then
  echo "❌ Epic $EPIC_ID not found in specs/epics/"
  echo ""
  echo "Available epics:"
  ls specs/epics/ | grep "^e[0-9]" | sed 's/-.*$//'
  exit 1
fi

echo "✓ Epic $EPIC_ID found"

# Update state using Python
cat > /tmp/set_epic.py << 'PYEOF'
import yaml
import sys

epic_id = sys.argv[1]
state_file = 'specs/state.yaml'

with open(state_file, 'r') as f:
    data = yaml.safe_load(f)

data['active_flow'] = 'build_epic'
data['active_epic'] = epic_id

with open(state_file, 'w') as f:
    yaml.dump(data, f, default_flow_style=False, sort_keys=False)

print(f"✓ Set active_epic = {epic_id}")
PYEOF

python3 /tmp/set_epic.py "$EPIC_ID"
rm /tmp/set_epic.py

echo "✓ State configured"
echo ""
echo "Ready to build! Run:"
echo ""
echo "  build-epic"
echo ""
