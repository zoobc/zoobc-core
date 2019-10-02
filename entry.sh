#!/bin/sh

function init()
{
	# echo error message, when executable file doesn't exist.
	if CMD=$(command -v "$1" 2>/dev/null); then
		shift
		if [ "$CMD" = "/bin/zoobc-core" ]; then
			exec "$CMD" "${CONFIG_PATH:+--config-path=$CONFIG_PATH}" "${CONFIG_POSTFIX:+--config-postfix=_$CONFIG_POSTFIX}" "$@"
		else
			exec "$CMD" "$@"
		fi
	else
		echo "Command not found: $1"
		exit 1
	fi
}

function updateConfig()
{
	configPath="$(cd "${CONFIG_PATH:-${2##*=}}";pwd)"
	configPostfix="${CONFIG_POSTFIX:-${3##*=}}"
	configFile="${configPath}/config${configPostfix:+_$configPostfix}.toml"
	resourcePath="${configPath}/${configPostfix}"
	nodeKeysFile="${resourcePath}/node_keys.json"

	if [ "$BALENA" = '1' ]; then
		jsonAwkFile="$(dirname "$0")/JSON.awk"
		ipAddress=$(wget -qO- --header "Content-Type:application/json" "$BALENA_SUPERVISOR_ADDRESS/v1/device?apikey=$BALENA_SUPERVISOR_API_KEY" | awk -f "$jsonAwkFile" - | awk '/\["ip_address"\]/ { print $2 }')
	else
		ipAddress='"127.0.0.1"'
	fi

	[ -f $configFile ] || fail "${configFile} missing or cannot be opened"
	sed -i "/configPath\s*=/ c configPath=\"${resourcePath}\"" ${configFile}
	sed -i "/dbPath\s*=/ c dbPath=\"${resourcePath}\"" ${configFile}
	sed -i "/badgerDbPath\s*=/ c badgerDbPath=\"${resourcePath}\"" ${configFile}
	sed -i "/myAddress\s*=/ c myAddress=${ipAddress}" ${configFile}
	[ -n "${PEER_PORT}" ] && sed -i "/peerPort\s*=/ c peerPort=${PEER_PORT}" ${configFile}
	[ -n "${WELLKNOWN_PEERS}" ] && sed -i "/wellknownPeers\s*=/ c wellknownPeers=${WELLKNOWN_PEERS}" ${configFile}
	[ -n "${OWNER_ACCOUNT_ADDRESS}" ] && sed -i "/ownerAccountAddress\s*=/ c ownerAccountAddress=\"$OWNER_ACCOUNT_ADDRESS\"" ${configFile}
	[ -n "${SMITHING}" ] && sed -i "/smithing\s*=/ c smithing=$SMITHING" ${configFile}
	echo ">>> $configFile..."
	cat "$configFile"
	echo "<<< $configFile"
	echo

	mkdir -p "${resourcePath}"
	[ -n "${NODE_PUBLIC_KEY}" -a -n "${NODE_SEED}" -a ! -f $nodeKeysFile ] && cat <<-EOF > ${nodeKeysFile}
	[
	  {
	    "PublicKey": "${NODE_PUBLIC_KEY}",
	    "Seed": "${NODE_SEED}"
	  }
	]
	EOF
}

function fail {
    printf '%s\n' "$1" >&2  ## Send message to stderr. Exclude >&2 if you don't want it that way.
    exit "${2-1}"  ## Return a code specified by $2 or 1 by default.
}

DEFAULT_CONFIG_PATH="/data/zoobc-core"
DEFAULT_CONFIG_POSTFIX="${BALENA_SERVICE_NAME}"
#DEFAULT_CONFIG_POSTFIX="${BALENA_SERVICE_NAME:+$(echo $BALENA_SERVICE_NAME | awk '{for(i=1;i<=NF;i++){ $i=toupper(substr($i,1,1)) substr($i,2) }}1')}"
echo "$1 ${1:+${2:---config-path=${CONFIG_PATH:=$DEFAULT_CONFIG_PATH}}} ${1:+${3:---config-postfix=${CONFIG_POSTFIX:=$DEFAULT_CONFIG_POSTFIX}}}"

updateConfig "$@"
init "$@"
