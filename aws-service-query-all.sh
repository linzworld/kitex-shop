#!/usr/bin/env bash

set -euo pipefail

AWS_REGION="us-west-1"
ECS_CLUSTER="test-rpc-ecs"
OUTPUT_FILE="aws-services.json"

SERVICE_NAMES="$(
  aws ecs list-services \
    --region "${AWS_REGION}" \
    --cluster "${ECS_CLUSTER}" \
    --output json |
  jq -r '.serviceArns[] | split("/")[-1]'
)"

if [[ -z "${SERVICE_NAMES}" ]]; then
  echo "集群 ${ECS_CLUSTER} 中没有 ECS Service"
  exit 0
fi

while IFS= read -r SERVICE_NAME; do
  aws ecs describe-services \
    --region "${AWS_REGION}" \
    --cluster "${ECS_CLUSTER}" \
    --services "${SERVICE_NAME}" \
    --include TAGS \
    --output json
done <<< "${SERVICE_NAMES}" |
jq -s '{
  services: [.[] | .services[]],
  failures: [.[] | .failures[]]
}' > "${OUTPUT_FILE}"

echo "查询到以下 Service："
echo "${SERVICE_NAMES}"
echo
echo "完整配置已保存到：${OUTPUT_FILE}"
