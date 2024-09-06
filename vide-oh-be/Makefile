build:
	$(MAKE) -C user-service build

deploy: build
	serverless deploy
	
clean:
	rm -rf ./bin ./vendor Gopkg.lock ./serverless