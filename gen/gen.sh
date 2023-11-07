#!/bin/sh

export GO_POST_PROCESS_FILE="gofmt -w"
openapi-generator generate -i https://api.dynatrace.com/spec-json -g go -t template -o account_management --remove-operation-id-prefix --enable-post-process-file --additional-properties=packageName=accountmanagement
