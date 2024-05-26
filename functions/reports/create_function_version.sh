#!/bin/bash
source env.sh

rm reports_fun.zip
zip reports_fun go.mod index.go reports.go
yc serverless function version create \
    --function-id="$FUNCTION_ID" \
    --runtime golang121 \
    --entrypoint index.Handler \
    --memory 128m \
    --execution-timeout 20s \
    --source-path reports_fun.zip \
    --service-account-id="$SA_ACCOUNT_ID" \
    --environment YDB_DSN="$YDB_DSN"