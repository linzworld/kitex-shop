#!/usr/bin/env bash

set -euo pipefail

AWS_REGION="us-west-1"
AWS_ACCOUNT_ID="236763663116"

LOCAL_IMAGE="kitex-shop-api:latest"
REPOSITORY_NAME="kitex-shop-api"
IMAGE_TAG="latest"

REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

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
