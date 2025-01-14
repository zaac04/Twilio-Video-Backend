#!/bin/bash

set -eo pipefail #exit if any command fails

#initialize the variables for storing command line arguments
TEMP_REGION=""
TEMP_PROFILE=""
TEMP_IMAGE=""
TEMP_TASK=""
TEMP_SECRET=""
TEMP_ENV=""
TEMP_SECRET_NAME=""

#change the base directory to the location of running script
cd "$(dirname "$0")" || exit

while [[ "$#" -gt 0 ]]; do
    case $1 in
    --aws-region)
        TEMP_REGION="$2"
        shift 2
        ;;
    --aws-profile)
        TEMP_PROFILE="$2"
        shift 2
        ;;
    --image)
        TEMP_IMAGE="$2"
        shift 2
        ;;
    --task-name)
        TEMP_TASK="$2"
        shift 2
        ;;
    --env)
        TEMP_ENV="$2"
        shift 2
        ;;
    --secret)
        TEMP_SECRET="$2"
        shift 2
        ;;
    --secret-manager-name)
        TEMP_SECRET_NAME="$2"
        shift 2
        ;;
    *) echo "unknown parameter: $1" ;;
    esac
done

#checks if variables are empty if so exit
if [[ -z "$TEMP_REGION" || -z "$TEMP_PROFILE" || -z "$TEMP_IMAGE" || -z "$TEMP_TASK" || -z "$TEMP_ENV" || -z "$TEMP_SECRET" || -z "$TEMP_SECRET_NAME" ]]; then
    echo "Error: --aws-region, --aws-profile, --task-name, --image, --env,  --secret, --secret-manager-name   are required."
    exit 1
fi

convert_file_to_json_array() {
    filename="$1"
    json_array="["

    while IFS='=' read -r key value; do
        if [[ -z "$key" || "$key" =~ ^# ]]; then
            continue
        fi
        key=$(echo "$key" | xargs)
        value=$(echo "$value" | xargs)
        json_array+="{\"name\": \"$key\", \"value\": \"$value\"},"
    done <"$filename"
    json_array="${json_array%,}]"
    echo "$json_array"
}

convert_secret_file_to_json_array() {
    filename="$1"
    json_array="["
    secret_arn=$(aws secretsmanager list-secrets --query "SecretList[?Name=='${TEMP_SECRET_NAME}'].ARN" --output text --profile $TEMP_PROFILE --region $TEMP_REGION)

    while IFS='=' read -r key value; do
        if [[ -z "$key" || "$key" =~ ^# ]]; then
            continue
        fi
        key=$(echo "$key" | xargs)
        value=$(echo "$value" | xargs)
        json_array+="{\"name\": \"$key\", \"valueFrom\": \"$secret_arn:$value\"},"
    done <"$filename"

    json_array="${json_array%,}]"
    echo "$json_array"
}

NEW_ENV=$(convert_file_to_json_array "$TEMP_ENV")
NEW_ENV_SECRETS=$(convert_secret_file_to_json_array "$TEMP_SECRET")

echo $NEW_ENV

TASK_DEFINITION=$(aws ecs describe-task-definition --task-definition $TEMP_TASK --region $TEMP_REGION --profile "$TEMP_PROFILE")

NEW_CONTAINER_DEFINTIION=$(echo $TASK_DEFINITION | jq --arg IMAGE "$TEMP_IMAGE" \
    --argjson NEW_ENV "$NEW_ENV" \
    --argjson NEW_ENV_SECRET "$NEW_ENV_SECRETS" \
    '.taskDefinition | .containerDefinitions[0].image = $IMAGE |
    .containerDefinitions[0].environment = $NEW_ENV |
    .containerDefinitions[0].secrets = $NEW_ENV_SECRET |  
    del(.taskDefinitionArn) |
    del(.revision) |
    del(.status) |
    del(.requiresAttributes) |
    del(.compatibilities) |
    del(.registeredAt) |
    del(.registeredBy) |
    .containerDefinitions')

VOLUMES=$(echo $TASK_DEFINITION | jq '.taskDefinition.volumes')
taskRoleArn=$(echo $TASK_DEFINITION | jq -r '.taskDefinition.taskRoleArn')
executionRoleArn=$(echo $TASK_DEFINITION | jq -r '.taskDefinition.executionRoleArn')
network_mode=$(echo $TASK_DEFINITION | jq -r '.taskDefinition.networkMode')

aws ecs register-task-definition --region $TEMP_REGION --profile "$TEMP_PROFILE" --family $TEMP_TASK --container-definitions "${NEW_CONTAINER_DEFINTIION}" --volumes "${VOLUMES}" --task-role-arn $taskRoleArn --execution-role-arn $executionRoleArn --network-mode $network_mode
