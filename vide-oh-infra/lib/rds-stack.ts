import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as rds from 'aws-cdk-lib/aws-rds';
import * as secretsmanager from 'aws-cdk-lib/aws-secretsmanager';
import { Construct } from 'constructs';

const DB_PORT = 5432
const POSTGRES_VERSION = rds.PostgresEngineVersion.VER_16; 

interface RDSStackProps extends cdk.StackProps {
    vpc: ec2.Vpc
}

export class RDSStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: RDSStackProps) {
    super(scope, id, props);
    const vpc = props.vpc;

    // Create a security group for the RDS instance
    const rdsSecurityGroup = new ec2.SecurityGroup(this, 'RDSSecurityGroup', {
        vpc,
        description: 'Allow TCP access using specified port to RDS/Postgres instance',
    });
    rdsSecurityGroup.addIngressRule(
        ec2.Peer.ipv4(vpc.vpcCidrBlock),
        ec2.Port.tcp(DB_PORT),
        'Allow TCP from VPC'
    );

    // Create a Secrets Manager secret to store database credentials
    const rdsCredentialsSecret = new secretsmanager.Secret(this, 'RDSCredentialsSecret', {
        generateSecretString: {
            secretStringTemplate: JSON.stringify({ username: 'postgres' }),
            generateStringKey: 'password',
            excludeCharacters: '/@"',
        },
    });

    // Parameter group for not enforcing SSL
    const parameterGroup = new rds.ParameterGroup(this, 'RDSParameterGroup', {
        engine: rds.DatabaseInstanceEngine.postgres({ version: POSTGRES_VERSION }),
        parameters: {
            'rds.force_ssl': '0',
        },
    });

    // Create the RDS PostgreSQL instance
    const rdsInstance = new rds.DatabaseInstance(this, 'RDSPostgresInstance', {
        engine: rds.DatabaseInstanceEngine.postgres({ version: POSTGRES_VERSION }),
        vpc,
        securityGroups: [rdsSecurityGroup],
        credentials: rds.Credentials.fromSecret(rdsCredentialsSecret),
        instanceType: ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.MICRO), // Free tier eligible
        allocatedStorage: 20, // Minimum storage to remain within free tier
        maxAllocatedStorage: 20,
        publiclyAccessible: false, // Set to true if Lambda is in a different VPC
        databaseName: 'videoh',
        multiAz: false,
        deletionProtection: false, // Ensure this is false for easy cleanup
        removalPolicy: cdk.RemovalPolicy.DESTROY, // Destroy the database when the stack is deleted
        parameterGroup: parameterGroup
    });

    // Output the RDS endpoint and credentials
    new cdk.CfnOutput(this, 'RdsSecretName', {
        value: rdsCredentialsSecret.secretName,
        exportName: 'RdsSecretName',
    });
  }
}
