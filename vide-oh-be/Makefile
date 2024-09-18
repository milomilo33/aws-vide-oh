generate-secret:
	SECRET_KEY=$$(openssl rand -base64 32) && \
	echo "Generated SECRET_KEY: $$SECRET_KEY" && \
	# aws secretsmanager update-secret --secret-id VideohSecretKey --secret-string "{\"secretKey\":\"$$SECRET_KEY\"}" || \
	aws secretsmanager create-secret --name VideohSecretKey --secret-string "{\"secretKey\":\"$$SECRET_KEY\"}" || \
	echo "Skipping secret creation, it may already exist"
	echo "Secret generated and stored in Secrets Manager"

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

# publish-ffmpeg-layer:
# 	@LAYER_NAME=ffmpeg; \
# 	ZIP_FILE=ffmpeg-layer.zip; \
# 	echo "Checking if layer exists..."; \
# 	layer_arn=$$(aws lambda list-layers --query "Layers[?LayerName=='$$LAYER_NAME'].LayerArn" --output text); \
# 	echo "Layer ARN (raw output): $$layer_arn"; \
# 	if [ -z "$$layer_arn" ]; then \
# 		echo "Layer does not exist. Publishing new layer..."; \
# 		(cd ffmpeg-layer && zip -r $$OLDPWD/$$ZIP_FILE bin/); \
# 		echo "Zipping completed."; \
# 		echo "Publishing layer..."; \
# 		aws lambda publish-layer-version --layer-name $$LAYER_NAME --zip-file fileb://$$ZIP_FILE --compatible-runtimes provided.al2023; \
# 	else \
# 		echo "Layer already exists. Skipping publish."; \
# 	fi

build:
	$(MAKE) -C user-service build
	$(MAKE) -C video-service build
	$(MAKE) -C comment-service build

deploy: generate-secret make-s3-bucket build
	serverless deploy
	
clean:
	rm -rf ./bin ./vendor Gopkg.lock ./serverless