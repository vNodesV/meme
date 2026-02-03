#!/bin/bash

set -e

GIT_TAG=$(git describe --tags)

echo "> Running $GIT_TAG..."

docker run --rm -it -p 26657:26657 \
  -e CHAIN_ID=meme-local-1 \
  -e STAKE_TOKEN=umeme \
  -e FEE_TOKEN=umeme \
  --name meme-local-1 \
  MeMeCosmos/meme:$GIT_TAG /bin/sh -c "./setup_and_run.sh"
