#!/usr/bin/env bash

set -euo pipefail

AWS_REGION="us-west-1"
AWS_ACCOUNT_ID="236763663116"

LOCAL_IMAGE="kitex-shop-stock:latest"
REPOSITORY_NAME="kitex-shop-stock"
IMAGE_TAG="latest"
DOCKERFILE="Dockerfile.stock"
TARGET_PLATFORM="linux/amd64"

REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

echo "Building ${LOCAL_IMAGE} for ${TARGET_PLATFORM}..."
docker buildx build \
  --platform "${TARGET_PLATFORM}" \
  --file "${DOCKERFILE}" \
  --tag "${LOCAL_IMAGE}" \
  --load \
  .

IMAGE_PLATFORM="$(docker image inspect "${LOCAL_IMAGE}" --format '{{.Os}}/{{.Architecture}}')"
if [[ "${IMAGE_PLATFORM}" != "${TARGET_PLATFORM}" ]]; then
  echo "Unexpected image platform: ${IMAGE_PLATFORM}; expected ${TARGET_PLATFORM}" >&2
  exit 1
fi

echo "Checking AWS identity..."
aws sts get-caller-identity

echo "Ensuring ECR repository exists: ${REPOSITORY_NAME}"
if ! aws ecr describe-repositories \
  --region "${AWS_REGION}" \
  --repository-names "${REPOSITORY_NAME}" >/dev/null 2>&1; then
  aws ecr create-repository \
    --region "${AWS_REGION}" \
    --repository-name "${REPOSITORY_NAME}" \
    --image-scanning-configuration scanOnPush=true
fi

echo "Logging Docker in to ECR..."
aws ecr get-login-password \
  --region "${AWS_REGION}" |
  docker login \
    --username AWS \
    --password-stdin "${REGISTRY}"

REPOSITORY_URI="$(
  aws ecr describe-repositories \
    --region "${AWS_REGION}" \
    --repository-names "${REPOSITORY_NAME}" \
    --query 'repositories[0].repositoryUri' \
    --output text
)"

IMAGE_URI="${REPOSITORY_URI}:${IMAGE_TAG}"

echo "Tagging image: ${LOCAL_IMAGE} -> ${IMAGE_URI}"
docker tag "${LOCAL_IMAGE}" "${IMAGE_URI}"

echo "Pushing image..."
docker push "${IMAGE_URI}"

DIGEST="$(
  aws ecr describe-images \
    --region "${AWS_REGION}" \
    --repository-name "${REPOSITORY_NAME}" \
    --image-ids imageTag="${IMAGE_TAG}" \
    --query 'imageDetails[0].imageDigest' \
    --output text
)"

echo
echo "Image pushed successfully."
echo "Tagged URI: ${IMAGE_URI}"
echo "Digest URI: ${REPOSITORY_URI}@${DIGEST}"
