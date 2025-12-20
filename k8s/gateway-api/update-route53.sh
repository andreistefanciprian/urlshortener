#!/bin/bash
# Update AWS Route53 A records for netl1.com and app.netl1.com
# Usage:
#   HOSTED_ZONE_ID=Z01644512M7PVAKK2E83U NEW_IP=34.40.158.236 ./update-route53.sh
#   Or edit the variables below and run: ./update-route53.sh

set -euo pipefail

# ----- Config -----
HOSTED_ZONE_ID=${HOSTED_ZONE_ID:-"Z01644512M7PVAKK2E83U"}
NEW_IP=${NEW_IP:-"1.2.3.4"}
ROOT_DOMAIN=${ROOT_DOMAIN:-"netl1.com"}
APP_DOMAIN=${APP_DOMAIN:-"app.netl1.com"}
TTL=${TTL:-60}

# ----- Checks -----
command -v aws >/dev/null 2>&1 || { echo "ERROR: aws CLI not found. Install and configure it first."; exit 1; }

if [[ -z "$HOSTED_ZONE_ID" || -z "$NEW_IP" ]]; then
  echo "ERROR: HOSTED_ZONE_ID and NEW_IP must be set."
  echo "Example: HOSTED_ZONE_ID=Z01644512M7PVAKK2E83U NEW_IP=34.40.158.236 $0"
  exit 1
fi

# ----- Functions -----
upsert_record() {
  local name="$1"
  echo "Updating A record for $name -> $NEW_IP (TTL=$TTL)"
  aws route53 change-resource-record-sets --hosted-zone-id "$HOSTED_ZONE_ID" --change-batch "$(cat <<JSON
{
  "Comment": "Update $name to $NEW_IP",
  "Changes": [{
    "Action": "UPSERT",
    "ResourceRecordSet": {
      "Name": "$name",
      "Type": "A",
      "TTL": $TTL,
      "ResourceRecords": [{"Value": "$NEW_IP"}]
    }
  }]
}
JSON
)"
}

# ----- Execute -----
upsert_record "$ROOT_DOMAIN"
upsert_record "$APP_DOMAIN"

echo "Done. Verify with:"
echo "  dig +short $ROOT_DOMAIN"
echo "  dig +short $APP_DOMAIN"
