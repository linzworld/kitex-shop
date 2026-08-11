#!/usr/bin/env bash

set -euo pipefail

AWS_REGION="${AWS_REGION:-us-west-1}"
NAMESPACE_NAME="${NAMESPACE_NAME:-dev1-kitex-shop}"
OUTPUT_FILE="${OUTPUT_FILE:-aws-namespace-dev1.json}"

find_namespace() {
  aws servicediscovery list-namespaces \
    --region "${AWS_REGION}" \
    --output json |
  jq -c \
    --arg name "${NAMESPACE_NAME}" \
    '[.Namespaces[] | select(.Name == $name and .Type == "HTTP")][0]'
}

NAMESPACE="$(find_namespace)"

if [[ "${NAMESPACE}" != "null" ]]; then
  jq . <<< "${NAMESPACE}" > "${OUTPUT_FILE}"
  echo "命名空间已存在：${NAMESPACE_NAME}"
  echo "配置已保存到：${OUTPUT_FILE}"
  exit 0
fi

echo "创建 HTTP 命名空间：${NAMESPACE_NAME}"

OPERATION_ID="$(
  aws servicediscovery create-http-namespace \
    --region "${AWS_REGION}" \
    --name "${NAMESPACE_NAME}" \
    --description "ECS Service Connect namespace for ${NAMESPACE_NAME}" \
    --query 'OperationId' \
    --output text
)"

while true; do
  OPERATION="$(
    aws servicediscovery get-operation \
      --region "${AWS_REGION}" \
      --operation-id "${OPERATION_ID}" \
      --query 'Operation' \
      --output json
  )"

  STATUS="$(jq -r '.Status' <<< "${OPERATION}")"

  case "${STATUS}" in
    SUCCESS)
      break
      ;;

    FAIL)
      echo "创建命名空间失败：" >&2
      jq '{ErrorCode, ErrorMessage}' <<< "${OPERATION}" >&2
      exit 1
      ;;

    SUBMITTED | PENDING)
      echo "等待创建完成，当前状态：${STATUS}"
      sleep 2
      ;;

    *)
      echo "未知操作状态：${STATUS}" >&2
      exit 1
      ;;
  esac
done

NAMESPACE="$(find_namespace)"

if [[ "${NAMESPACE}" == "null" ]]; then
  echo "操作已完成，但未查询到命名空间：${NAMESPACE_NAME}" >&2
  exit 1
fi

jq . <<< "${NAMESPACE}" > "${OUTPUT_FILE}"

echo "命名空间创建成功：${NAMESPACE_NAME}"
echo "配置已保存到：${OUTPUT_FILE}"
jq '{Id, Arn, Name, Type}' "${OUTPUT_FILE}"
