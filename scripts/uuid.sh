#!/bin/bash
# uuid.sh - Generate a UUID (v4)

if command -v uuidgen >/dev/null 2>&1; then
  uuidgen
elif command -v cat >/dev/null 2>&1 && [ -r /proc/sys/kernel/random/uuid ]; then
  cat /proc/sys/kernel/random/uuid
else
  # Fallback using Python if available
  python3 -c 'import uuid; print(uuid.uuid4())'
fi
