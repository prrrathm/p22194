#!/bin/bash
# bp-yaml-set.sh - Set YAML values in bigpowers state files
# Usage: bash scripts/bp-yaml-set.sh <key> <value>
# Example: bash scripts/bp-yaml-set.sh active_epic e02

if [ $# -lt 2 ]; then
  echo "Usage: $0 <key> <value>"
  echo "Example: $0 active_epic e02"
  exit 1
fi

KEY="$1"
VALUE="$2"
STATE_FILE="specs/state.yaml"

if [ ! -f "$STATE_FILE" ]; then
  echo "❌ Error: $STATE_FILE not found"
  exit 1
fi

# Use Python to update YAML
cat > /tmp/yaml_update.py << 'PYEOF'
import yaml
import sys

key = sys.argv[1]
value = sys.argv[2]
state_file = sys.argv[3]

with open(state_file, 'r') as f:
    data = yaml.safe_load(f)

# Handle nested keys like "epic_cycle.step"
if '.' in key:
    keys = key.split('.')
    current = data
    for k in keys[:-1]:
        if k not in current:
            current[k] = {}
        current = current[k]
    current[keys[-1]] = value
else:
    data[key] = value

with open(state_file, 'w') as f:
    yaml.dump(data, f, default_flow_style=False, sort_keys=False)

print(f"✓ Set {key} = {value}")
PYEOF

python3 /tmp/yaml_update.py "$KEY" "$VALUE" "$STATE_FILE"
rm /tmp/yaml_update.py
