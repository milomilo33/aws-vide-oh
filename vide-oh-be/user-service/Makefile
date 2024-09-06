build:
	env GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -ldflags="-s -w" -o ./bin/bootstrap
	(cd bin && zip lambda-handler.zip bootstrap)