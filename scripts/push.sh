#!/bin/bash

set -eo pipefail #exit if any command fails

#initialize the variables for storing command line arguments
TEMP_REGION=""
TEMP_PROFILE=""
TEMP_ACCOUNT_ID=""
TEMP_IMAGE=""

#change the base directory to the location of running script 
cd "$(dirname "$0")" || exit

#checks for arguments being passed
#matches patterns if ok store it in variable, 
#shift removes the current working argument from argument array
#shift 2 because remove key and value eg:(--repo my_repo)

while [[ "$#" -gt 0 ]]; do 
    case $1 in 
        --aws-region) TEMP_REGION="$2"; shift 2;;
        --aws-profile) TEMP_PROFILE="$2";  shift 2;;
        --aws-acc-id) TEMP_ACCOUNT_ID="$2"; shift 2;;
        --image) TEMP_IMAGE="$2";  shift 2;;
        *) echo "unknown parameter: $1";;
    esac
done

#checks if variables are empty if so exit
if [[ -z "$TEMP_REGION" || -z "$TEMP_PROFILE" || -z "$TEMP_ACCOUNT_ID" || -z "$TEMP_IMAGE" ]]; then
    echo "Error: --aws-region, --aws-profile, --aws-acc-id, --image are required."
    exit 1
fi


echo "Logging in to AWS account $TEMP_ACCOUNT_ID.dkr.ecr.$TEMP_REGION.amazonaws.com"
aws ecr get-login-password --region "$TEMP_REGION" --profile "$TEMP_PROFILE"|docker login --username AWS --password-stdin "$TEMP_ACCOUNT_ID".dkr.ecr."$TEMP_REGION".amazonaws.com

echo "Pushing image to ECR"
docker push "$TEMP_IMAGE"