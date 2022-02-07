#! /bin/bash

FILEPHRASE=/var/edge-orchestration/data/jwt/passPhraseJWT.txt
FILEPUBKEY=/var/edge-orchestration/data/jwt/app_rsa.key

device_id=`cat /var/edge-orchestration/device/orchestration_deviceID.txt`

# Token life time 24h = 86400s
payload="{
	\"exp\": $(($(date +%s)+86400)),
	\"iat\": $(date +%s),
	\"deviceid\": \"${device_id}\",
	\"aud\": \"$2\"
}"

header="{
	\"typ\": \"JWT\",
	\"alg\": \"$1\"
}"

base64_encode()
{
	openssl enc -base64 -A | tr '+/' '-_' | tr -d '='
}

json() {
	jq -c . | LC_CTYPE=C tr -d '\n'
}

hmacsha256_sign()
{
	openssl dgst -binary -sha256 -hmac "${passphrase}"
}


rsa256sha256_sign() {
	openssl dgst -binary -sha256 -sign <(printf '%s\n' "$1")
}

header_base64=$(printf %s "${header}" | json | base64_encode)
payload_base64=$(printf %s "${payload}" | json | base64_encode)
header_payload=$(printf %s "${header_base64}.${payload_base64}")

case $1 in
    HS256)
		[ ! -e ${FILEPHRASE} ] && echo "The ${FILEPHRASE} does not exist" && return 1
		passphrase=`cat ${FILEPHRASE}`
		signature=$(printf %s "${header_payload}" | hmacsha256_sign | base64_encode)
		;;
    RS256)
		[ ! -e ${FILEPUBKEY} ] && echo "The ${FILEPUBKEY} does not exist" && return 1
		rsa_pubkey=`cat ${FILEPUBKEY}`
		signature=$(printf %s "${header_payload}" | rsa256sha256_sign "$rsa_pubkey" | base64_encode)
		;;
    *)
        echo "Usage:"
        echo "-------------------------------------------------------------------------------------------------"
        echo "  $ . jwt_gen.sh [Algo] [User]  : Genereate JWT based on Algo:{HS256, RS256} User:{Admin, Member}"
        echo "  $ . jwt_gen.sh RS256 Admin    : Genereate JWT based on RS256 and User - Admin                  "
        echo "-------------------------------------------------------------------------------------------------"
		return 1
        ;;
esac

echo -e "\nheader = ${header}\n"
echo -e "payload = ${payload}\n"
echo -e "token = ${header_payload}.${signature}\n"

export EDGE_ORCHESTRATION_TOKEN=${header_payload}.${signature}
