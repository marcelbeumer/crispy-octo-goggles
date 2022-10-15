#!/bin/sh
aws lambda update-function-code --function-name go-playground-aws-lambda --zip-file fileb://dist/main.zip | cat
