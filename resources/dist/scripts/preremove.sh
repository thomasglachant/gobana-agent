#!/bin/sh
set -e

if [ -d /run/systemd/system ] && [ "$1" = remove ]; then
	deb-systemd-invoke stop spooter_agent.service >/dev/null || true
fi
