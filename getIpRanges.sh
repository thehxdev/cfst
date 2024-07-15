#!/bin/sh

if ! command -v curl >/dev/null; then
    echo '[ERROR] curl not found in your PATH. make sure that you installed curl.'
    exit 1
fi

BASE_URL='https://www.cloudflare.com'
IPV4_URL="$BASE_URL/ips-v4"
IPV6_URL="$BASE_URL/ips-v6"

echo "[INFO] downloading ipv4 ranges from $IPV4_URL"
curl -s -o 'ip.txt' "$IPV4_URL"

echo "[INFO] downloading ipv6 ranges from $IPV6_URL"
curl -s -o 'ipv6.txt' "$IPV6_URL"
