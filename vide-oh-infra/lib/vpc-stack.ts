import * as cdk from 'aws-cdk-lib';
import * as ec2 from "aws-cdk-lib/aws-ec2";
import { Construct } from 'constructs';
import { RemovalPolicy } from "aws-cdk-lib";

export class VPCStack extends cdk.Stack {
    public readonly vpc: ec2.Vpc;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const vpc = new ec2.Vpc(this, 'vpc-videoh', {
            maxAzs: 2,
            subnetConfiguration: [
                {
                    subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
                    name: 'PrivateSubnet',
                    cidrMask: 24
                },
                {
                    subnetType: ec2.SubnetType.PUBLIC,
                    name: 'PublicSubnet',
                    cidrMask: 24
                },
            ],
            // CHANGE
            natGateways: 0
        });

        vpc.applyRemovalPolicy(RemovalPolicy.DESTROY)

        this.vpc = vpc;

        const lambdaSG = new ec2.SecurityGroup(this, "lambda-sg", {
            vpc: vpc,
            description: "Allow all outbound traffic for Lambda",
            allowAllOutbound: true,
        });

        // new ec2.GatewayVpcEndpoint(this, 'S3VpcEndpoint', {
        //     vpc,
        //     service: ec2.GatewayVpcEndpointAwsService.S3,
        //     subnets: [
        //         {
        //             subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS
        //         }
        //     ]
        // });

        // vpc.addGatewayEndpoint('DynamoDbEndpoint', {
        //     service: ec2.GatewayVpcEndpointAwsService.DYNAMODB,
        //     subnets: [
        //         { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS }
        //     ]
        // });
        // new ec2.InterfaceVpcEndpoint(this, 'SecretsManagerEndpoint', {
        //     vpc,
        //     service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER,
        //     privateDnsEnabled: true,
        //     securityGroups: [lambdaSG],
        // });

        new cdk.CfnOutput(this, 'VpcPrivateSubnet1', {
            value: vpc.privateSubnets[0].subnetId,
            exportName: 'VpcPrivateSubnet1'
        });
        new cdk.CfnOutput(this, 'LambdaSecurityGroupId', {
            value: lambdaSG.securityGroupId,
            exportName: 'LambdaSecurityGroupId'
        });
    }
}
