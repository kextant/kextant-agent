#!/usr/bin/env bash
set -euo pipefail

if ! command -v npm >/dev/null 2>&1; then
	echo "npm is required before installing pi-coding-agent and Playwright."
	exit 1
fi

export PLAYWRIGHT_BROWSERS_PATH=/usr/local/share/playwright-browsers

mkdir -p "${PLAYWRIGHT_BROWSERS_PATH}"
npm install -g @mariozechner/pi-coding-agent playwright
playwright install --with-deps chromium

# Make the shared browser cache readable by the devcontainer user.
chmod -R a+rX "${PLAYWRIGHT_BROWSERS_PATH}"
