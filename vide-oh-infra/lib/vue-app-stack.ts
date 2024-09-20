// import { Stack, StackProps } from 'aws-cdk-lib';
// import { BlockPublicAccess, Bucket } from 'aws-cdk-lib/aws-s3';
// import { BucketDeployment, Source } from 'aws-cdk-lib/aws-s3-deployment';
// import { Distribution } from 'aws-cdk-lib/aws-cloudfront';
// import { S3Origin } from 'aws-cdk-lib/aws-cloudfront-origins';
// import { Construct } from 'constructs';
// import { CfnBucketPolicy } from 'aws-cdk-lib/aws-s3';
// import path = require('path');

// export class VueAppStack extends Stack {
//   constructor(scope: Construct, id: string, props?: StackProps) {
//     super(scope, id, props);

//     const websiteBucket = new Bucket(this, 'VueCdkDemoSiteBucket', {
//       websiteIndexDocument: 'index.html',
//       blockPublicAccess: BlockPublicAccess.BLOCK_ACLS, // Allow public access via policies
//     });

//     // Add a bucket policy to allow public read access
//     new CfnBucketPolicy(this, 'BucketPolicy', {
//       bucket: websiteBucket.bucketName,
//       policyDocument: {
//         Version: '2012-10-17',
//         Statement: [
//           {
//             Effect: 'Allow',
//             Principal: '*',
//             Action: 's3:GetObject',
//             Resource: `${websiteBucket.bucketArn}/*`,
//           },
//         ],
//       },
//     });

//     const webDistribution = new Distribution(this, 'VueCdkDistribution', {
//       defaultBehavior: {
//         origin: new S3Origin(websiteBucket),
//         // You might need to add other properties here based on your requirements
//       },
//       errorResponses: [
//         {
//           httpStatus: 403,
//           responsePagePath: '/index.html',
//           responseHttpStatus: 200,
//         },
//         {
//           httpStatus: 404,
//           responsePagePath: '/index.html',
//           responseHttpStatus: 200,
//         },
//       ],
//     });
//     const SRC_PATH = path.join(__dirname, "../../vide-oh-fe/dist");
//     new BucketDeployment(this, 'VueCdkDemoSiteDeployment', {
//       sources: [Source.asset(SRC_PATH)],
//       destinationBucket: websiteBucket,
//       distribution: webDistribution,
//       distributionPaths: ['/*'],
//       prune: true
//     });
//   }
// }


import * as cdk from "aws-cdk-lib";
import {aws_cloudfront as cloudfront, aws_s3 as s3, aws_s3_deployment as s3Deployment} from "aws-cdk-lib";
import {Construct} from "constructs";
import * as path from "path";
import * as fs from 'fs';
import {spawnSync} from "child_process";

const SRC_PATH = path.join(__dirname, "../../vide-oh-fe");
const BUILD_CONFIG = "development";

export class VueAppStack extends cdk.Stack {
   constructor(scope: Construct, id: string, props?: cdk.StackProps) {
       super(scope, id, props);

    //    const videohSecretKey = secretsmanager.Secret.fromSecretNameV2(this, 'VideohSecretKey', 'VideohSecretKey');
    //    const apiKey = videohSecretKey.secretValueFromJson('secretKey').toString();

       const vueBucket = new s3.Bucket(this, "VueBucket", {
           bucketName: "vide-oh-fe-bucket",
           blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
           autoDeleteObjects: true,
           removalPolicy: cdk.RemovalPolicy.DESTROY,
       });

       const webDistribution = this.createCloudFrontDistribution(vueBucket);
       new cdk.CfnOutput(this, "VueAppDomainName", {
           value: webDistribution.distributionDomainName,
       });

       this.createDeployment(vueBucket, webDistribution);
   }

   private createCloudFrontDistribution(webAppBucket: s3.IBucket) {
       return new cloudfront.CloudFrontWebDistribution(
           this,
           "VueAppWebDistribution",
           {
               originConfigs: [
                   {
                       s3OriginSource: {
                           s3BucketSource: webAppBucket,
                           originAccessIdentity: new cloudfront.OriginAccessIdentity(
                               this,
                               "originAccessIdentityForVueBucket"
                           ),
                       },
                       behaviors: [{ isDefaultBehavior: true }],
                   },
               ],
               errorConfigurations: [
                   {
                       errorCode: 403,
                       errorCachingMinTtl: 60,
                       responseCode: 200,
                       responsePagePath: "/index.html",
                   },
                   {
                       errorCode: 404,
                       errorCachingMinTtl: 60,
                       responseCode: 200,
                       responsePagePath: "/index.html",
                   },
               ],
           }
       );
   }

    private createDeployment(
        webAppBucket: s3.IBucket,
        webDistribution: cloudfront.CloudFrontWebDistribution
    ) {
        new s3Deployment.BucketDeployment(this, "VueAppDeployment", {
            destinationBucket: webAppBucket,
            sources: [
                s3Deployment.Source.asset(path.join(__dirname, "../../vide-oh-fe/dist"), {
                    bundling: {
                        image: cdk.DockerImage.fromRegistry("local"),
                        local: {
                            tryBundle(outputDir: string) {
                                console.log("Executing local bundling...");
                                console.log(outputDir);

                                const getStackOutput = (stackName: string, outputKey: string): string => {
                                    return spawnSync('aws', [
                                        'cloudformation', 'describe-stacks',
                                        '--stack-name', stackName,
                                        '--query', `Stacks[0].Outputs[?OutputKey=='${outputKey}'].OutputValue`,
                                        '--output', 'text'
                                    ]).stdout.toString().trim();
                                }
                                const restApiBaseUrl = getStackOutput('vide-oh-dev', 'ServiceEndpoint');
                                const websocketApiBaseUrl = getStackOutput('videoh-websocket', 'WebSocketApiEndpoint');

                                const apiKeysResult = spawnSync('aws', [
                                    'apigateway', 'get-api-keys',
                                    '--query', 'items[?name==`videohApiKey`]',
                                    '--include-values',
                                    '--output', 'json'
                                ]).stdout.toString().trim();
                                const apiKeys = JSON.parse(apiKeysResult);
                                if (apiKeys.length === 0) {
                                    throw new Error('API key not found.');
                                }
                                const apiKey = apiKeys[0].value;
                                console.log('API Key:', apiKey);

                                const envFilePath = path.join(SRC_PATH, '/.env');
                                const envContent = `VUE_APP_REST_API_BASE_URL=${restApiBaseUrl}\nVUE_APP_API_KEY=${apiKey}\nVUE_APP_WEBSOCKET_API_BASE_URL=${websocketApiBaseUrl}\n`;
                                fs.writeFileSync(envFilePath, envContent, 'utf8');

                                spawnSync(
                                    [
                                        `cd ${SRC_PATH}`,
                                        `npm ci`,
                                        `npm run build`,
                                        `cp -r dist/* '${outputDir}'`,
                                    ].join(" && "),
                                    {
                                        shell: true,
                                        stdio: "inherit",
                                    },
                                );
                                return true;
                            },
                        },
                    },
                }),
            ],
            distribution: webDistribution,
        });
    }
}