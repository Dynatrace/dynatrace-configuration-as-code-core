#!/bin/sh

export GO_POST_PROCESS_FILE="gofmt -w"
openapi-generator generate -i https://api.dynatrace.com/spec-json -g go -t template -o account_management --remove-operation-id-prefix --enable-post-process-file --additional-properties=packageName=accountmanagement
openapi-generator generate -i schemas/buckets-spec.yaml -g go  -t template -o buckets --remove-operation-id-prefix --enable-post-process-file --additional-properties=packageName=buckets
openapi-generator generate -i schemas/automation-spec.yaml -g go  -t template -o automation --remove-operation-id-prefix --enable-post-process-file --additional-properties=packageName=automation
