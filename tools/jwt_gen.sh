#! /bin/bash

passphrase=`cat /var/edge-orchestration/data/jwt/passPhraseJWT.txt`

device_id=`cat /var/edge-orchestration/device/orchestration_deviceID.txt`

payload="{
	\"exp\": $(($(date +%s)+10)),
	\"iat\": $(date +%s),
	\"deviceid\": \"${device_id}\"
}"

header="{
	\"typ\": \"JWT\",
	\"alg\": \"HS256\"
}"

base64_encode()
{
	declare input=${1:-$(</dev/stdin)}
	printf '%s' "${input}" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n'
}

json() {
	declare input=${1:-$(</dev/stdin)}
	printf '%s' "${input}" | jq -c .
}

hmacsha256_sign()
{
	declare input=${1:-$(</dev/stdin)}
	printf '%s' "${input}" | openssl dgst -binary -sha256 -hmac "${passphrase}"
}

header_base64=$(echo "${header}" | json | base64_encode)
payload_base64=$(echo "${payload}" | json | base64_encode)

header_payload=$(echo "${header_base64}.${payload_base64}")
signature=$(echo "${header_payload}" | hmacsha256_sign | base64_encode)

export EDGE_ORCHESTRATION_TOKEN=${header_payload}.${signature}

echo -e "\nheader = ${header}\n"
echo -e "payload = ${payload}\n"
echo -e "passphrase = $passphrase\n"
echo -e "Token = ${header_payload}.${signature}\n"

