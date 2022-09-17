#!/bin/sh
set -e

if [ "$1" = "configure" ]; then
	# Add user and group
	if ! getent group gobana >/dev/null; then
		groupadd --system gobana
	fi
	if ! getent passwd gobana >/dev/null; then
		useradd --system \
			--gid gobana \
			--create-home \
			--home-dir /var/lib/gobana \
			--shell /usr/sbin/nologin \
			--comment "Gobana" \
			gobana
	fi

	# Add log directory with correct permissions
	if [ ! -d /var/log/gobana ]; then
		mkdir -p /var/log/gobana
		chown -R gobana:gobana /var/log/gobana
	fi
fi

if [ "$1" = "configure" ] || [ "$1" = "abort-upgrade" ] || [ "$1" = "abort-deconfigure" ] || [ "$1" = "abort-remove" ] ; then
	# This will only remove masks created by d-s-h on package removal.
	deb-systemd-helper unmask gobana_agent.service >/dev/null || true

	# was-enabled defaults to true, so new installations run enable.
	if deb-systemd-helper --quiet was-enabled gobana_agent.service; then
		# Enables the unit on first installation, creates new
		# symlinks on upgrades if the unit file has changed.
		deb-systemd-helper enable gobana_agent.service >/dev/null || true
		deb-systemd-invoke start gobana_agent.service >/dev/null || true
	else
		# Update the statefile to add new symlinks (if any), which need to be
		# cleaned up on purge. Also remove old symlinks.
		deb-systemd-helper update-state gobana_agent.service >/dev/null || true
	fi

	# Restart only if it was already started
	if [ -d /run/systemd/system ]; then
		systemctl --system daemon-reload >/dev/null || true
		if [ -n "$2" ]; then
			deb-systemd-invoke try-restart gobana_agent.service >/dev/null || true
		fi
	fi
fi
