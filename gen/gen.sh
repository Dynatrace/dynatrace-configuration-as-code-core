#!/bin/sh

export GO_POST_PROCESS_FILE="gofmt -w"
openapi-generator generate -i ./specs/account_management/spec_formatted_fixed.json -g go -t template -o account_management --remove-operation-id-prefix --enable-post-process-file --additional-properties=packageName=accountmanagement,disallowAdditionalPropertiesIfNotPresent=false
