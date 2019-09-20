#!/bin/sh

function init()
{
	# echo error message, when executable file doesn't exist.
	if CMD=$(command -v "$1" 2>/dev/null); then
		shift
		exec "$CMD" "${CONFIG_PATH:+--config-path=$CONFIG_PATH}" "${CONFIG_POSTFIX:+--config-postfix=$CONFIG_POSTFIX}" "$@"
	else
		echo "Command not found: $1"
		exit 1
	fi
}

function updateConfig()
{
	configPath="$(cd "${CONFIG_PATH:-${2##*=}}";pwd)"
        configPostfix="${CONFIG_POSTFIX:-${3##*=}}"
	configFile="${configPath}/config${configPostfix}.toml"
	jsonAwkFile="$(dirname "$0")/JSON.awk"
	if [ "$BALENA" = '1' ]; then
		ipAddress=$(wget -qO- --header "Content-Type:application/json" "$BALENA_SUPERVISOR_ADDRESS/v1/device?apikey=$BALENA_SUPERVISOR_API_KEY" | awk -f "$jsonAwkFile" - | awk '/\["ip_address"\]/ { print $2 }')
	else
		ipAddress='"127.0.0.1"'
	fi
	sed -i "/myAddress=/ c myAddress=$ipAddress" ${configPath}/*.toml
	echo ">>> $configFile..."
	cat "$configFile"
	echo "<<< $configFile"
	echo
}

DEFAULT_CONFIG_PATH="/data/zoobc-core"
DEFAULT_CONFIG_POSTFIX="${BALENA_SERVICE_NAME:+$(echo $BALENA_SERVICE_NAME | awk '{for(i=1;i<=NF;i++){ $i=toupper(substr($i,1,1)) substr($i,2) }}1')}"
echo "$1 ${1:+${2:---config-path=${CONFIG_PATH:=$DEFAULT_CONFIG_PATH}}} ${1:+${3:---config-postfix=${CONFIG_POSTFIX:=$DEFAULT_CONFIG_POSTFIX}}}"

updateConfig "$@"
init "$@"
