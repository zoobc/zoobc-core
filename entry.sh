#!/bin/sh

function init()
{
	# echo error message, when executable file doesn't exist.
	if CMD=$(command -v "$1" 2>/dev/null); then
		shift
		exec "$CMD" "$@"
	else
		echo "Command not found: $1"
		exit 1
	fi
}

function updateConfig()
{
	configPath="$(cd "${2##*=}";pwd)"
	configFile="$configPath/config${3##*=}.toml"
	jsonAwkFile="$(dirname "$0")/JSON.awk"
	ipAddress=$(wget -qO- --header "Content-Type:application/json" "$RESIN_SUPERVISOR_ADDRESS/v1/device?apikey=$RESIN_SUPERVISOR_API_KEY" | awk -f "$jsonAwkFile" - | awk '/\["ip_address"\]/ { print $2 }')
	sed -i "/myAddress=/ c myAddress=$ipAddress" $configPath/*.toml
	echo ">>> $configFile..."
	cat "$configFile"
	echo "<<< $configFile"
	echo
}

updateConfig "$@"
init "$@"
