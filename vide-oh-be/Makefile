generate-secret:
	SECRET_KEY=$$(openssl rand -base64 32) && \
	echo "Generated SECRET_KEY: $$SECRET_KEY" && \
	# aws secretsmanager update-secret --secret-id VideohSecretKey --secret-string "{\"secretKey\":\"$$SECRET_KEY\"}" || \
	aws secretsmanager create-secret --name VideohSecretKey --secret-string "{\"secretKey\":\"$$SECRET_KEY\"}" || \
	echo "Skipping secret creation, it may already exist"
	echo "Secret generated and stored in Secrets Manager"

make-s3-bucket:
	aws s3api create-bucket --bucket vide-oh-videos --region eu-central-1 \
    	--create-bucket-configuration LocationConstraint=eu-central-1

build:
	$(MAKE) -C user-service build
	$(MAKE) -C video-service build

deploy: generate-secret make-s3-bucket build
	serverless deploy
	
clean:
	rm -rf ./bin ./vendor Gopkg.lock ./serverless