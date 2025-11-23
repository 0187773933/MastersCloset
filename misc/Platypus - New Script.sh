#!/bin/bash

APP_PATH="$(dirname "$(dirname "$(dirname "$0")")")"
cd "$(dirname "$APP_PATH")/.files" || {
    echo "Could not find .files next to app."
    exit 1
}

open -a "Firefox" "http://localhost:5950/admin/user/checkin" >/dev/null 2>&1 &
./mct config.json