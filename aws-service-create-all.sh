#!/usr/bin/env bash

set -euo pipefail

AWS_REGION="us-west-1"
SOURCE_FILE="aws-services.json"
TARGET_CLUSTER="${TARGET_CLUSTER:?请设置 TARGET_CLUSTER}"

jq -c '.services[]' "${SOURCE_FILE}" |
while IFS= read -r SERVICE; do
  SERVICE_NAME="$(jq -r '.serviceName' <<< "${SERVICE}")"
  TASK_DEFINITION="$(jq -r '.taskDefinition' <<< "${SERVICE}")"
  DESIRED_COUNT="$(jq -r '.desiredCount' <<< "${SERVICE}")"
  LAUNCH_TYPE="$(jq -r '.launchType' <<< "${SERVICE}")"
  SCHEDULING_STRATEGY="$(jq -r '.schedulingStrategy' <<< "${SERVICE}")"

  NETWORK_CONFIGURATION="$(
    jq -c '.networkConfiguration' <<< "${SERVICE}"
  )"

  SERVICE_CONNECT_CONFIGURATION="$(
    jq -c '
      [
        .deployments[]
        | select(.status == "PRIMARY")
        | .serviceConnectConfiguration
      ][0]
    ' <<< "${SERVICE}"
  )"

  DEPLOYMENT_CONFIGURATION="$(
    jq -c '{
      maximumPercent: .deploymentConfiguration.maximumPercent,
      minimumHealthyPercent:
        .deploymentConfiguration.minimumHealthyPercent,
      deploymentCircuitBreaker: {
        enable:
          .deploymentConfiguration.deploymentCircuitBreaker.enable,
        rollback:
          .deploymentConfiguration.deploymentCircuitBreaker.rollback
      }
    }' <<< "${SERVICE}"
  )"

  COMMAND=(
    aws ecs create-service
    --region "${AWS_REGION}"
    --cluster "${TARGET_CLUSTER}"
    --service-name "${SERVICE_NAME}"
    --task-definition "${TASK_DEFINITION}"
    --desired-count "${DESIRED_COUNT}"
    --launch-type "${LAUNCH_TYPE}"
    --scheduling-strategy "${SCHEDULING_STRATEGY}"
    --network-configuration "${NETWORK_CONFIGURATION}"
    --deployment-configuration "${DEPLOYMENT_CONFIGURATION}"
  )

  if [[ "${SERVICE_CONNECT_CONFIGURATION}" != "null" ]]; then
    COMMAND+=(
      --service-connect-configuration
      "${SERVICE_CONNECT_CONFIGURATION}"
    )
  fi

  echo "创建 Service：${SERVICE_NAME}"
  "${COMMAND[@]}"
done
