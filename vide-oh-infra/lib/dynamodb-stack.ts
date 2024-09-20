import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';

export class DynamoDBStack extends cdk.Stack {
  public readonly tableConnections: dynamodb.TableV2;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    this.tableConnections = new dynamodb.TableV2(this, 'Table1', {
      partitionKey: { name: 'id', type: dynamodb.AttributeType.STRING },
      tableName: 'ws-connection',
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    new cdk.CfnOutput(this, 'TableNameConnections', {
      value: this.tableConnections.tableName,
      exportName: 'TableNameConnections'
    });
  }
}
