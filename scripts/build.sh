#!/bin/bash

set -eo pipefail #exit if any command fails

#initialize the variables for storing command line arguments
REPO=""
TAG=""

#change the base directory to the location of running script 
cd "$(dirname "$0")" || exit

#checks for arguments being passed
#matches patterns if ok store it in variable, 
#shift removes the current working argument from argument array
#shift 2 because remove key and value eg:(--repo my_repo)

while [[ "$#" -gt 0 ]]; do 
    case $1 in 
        --repo) REPO="$2"; shift 2;;
        --tag) TAG="$2";  shift 2;;
        *) echo "unknown parameter: $1";;
    esac
done


#checks if variables are empty if so exit
if [[ -z "$REPO" || -z "$TAG" ]]; then
    echo "Error: Both --repo and --tag are required."
    exit 1
fi

#builds image
echo "Building image $REPO:$TAG"
docker build --no-cache -t "$REPO:$TAG" -f ../deploy/Dockerfile ../
echo "Build Success $REPO:$TAG"
