.PHONY: build zip upload deploy all clean
.DEFAULT: build

build:
	GOOS=linux GOARCH=amd64 go build -o dist/main main.go

zip:
	-rm dist/main.zip
	zip -j dist/main.zip dist/main

upload:
	aws lambda update-function-code --function-name go-playground-aws-lambda --zip-file fileb://dist/main.zip | cat

deploy: build zip upload

all: build zip

clean:
	rm -r dist
