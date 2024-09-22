generate-secret:
	SECRET_KEY=$$(openssl rand -base64 32) && \
	echo "Generated SECRET_KEY: $$SECRET_KEY" && \
	# aws secretsmanager update-secret --secret-id VideohSecretKey --secret-string "{\"secretKey\":\"$$SECRET_KEY\"}" || \
	aws secretsmanager create-secret --name VideohSecretKey --secret-string "{\"secretKey\":\"$$SECRET_KEY\"}" || \
	echo "Skipping secret creation, it may already exist"

make-s3-bucket:
	@BUCKET_NAME=vide-oh-videos; \
	REGION=eu-central-1; \
	EXISTING_BUCKET=$$(aws s3api head-bucket --bucket $$BUCKET_NAME --region $$REGION 2>&1); \
	if echo $$EXISTING_BUCKET | grep -q 'Not Found'; then \
		echo "Bucket does not exist. Creating bucket..."; \
		aws s3api create-bucket --bucket $$BUCKET_NAME --region $$REGION --create-bucket-configuration LocationConstraint=$$REGION; \
	else \
		echo "Bucket already exists. Skipping creation."; \
	fi

build:
	$(MAKE) -C user-service build
	$(MAKE) -C video-service build
	$(MAKE) -C comment-service build

deploy: generate-secret make-s3-bucket build
	serverless deploy
	
clean:
	rm -rf ./bin ./vendor Gopkg.lock ./serverless