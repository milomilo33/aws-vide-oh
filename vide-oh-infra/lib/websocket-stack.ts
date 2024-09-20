import * as cdk from 'aws-cdk-lib';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import { Construct } from 'constructs';
import * as ec2 from "aws-cdk-lib/aws-ec2";

export class WebSocketStack extends cdk.Stack {
  public readonly webSocketApi: apigatewayv2.WebSocketApi;

  constructor(scope: Construct, id: string, props: cdk.StackProps) {
    super(scope, id, props);

    const websocketApi = new apigatewayv2.WebSocketApi(this, 'VideohWebSocketApi', {
        apiName: 'VideohWebSocketApi',
        description: 'WebSocket API for vide-oh',
    });

    const websocketStage = new apigatewayv2.WebSocketStage(this, 'WebSocketStage', {
      webSocketApi: websocketApi,
      stageName: 'dev',
      autoDeploy: true,
    });

    new cdk.CfnOutput(this, 'WebSocketApiId', {
      value: websocketApi.apiId,
      exportName: 'WebSocketApiId',
    });

    new cdk.CfnOutput(this, 'WebSocketApiEndpoint', {
      value: `${websocketApi.apiEndpoint}/${websocketStage.stageName}`,
      exportName: 'WebSocketApiEndpoint',
    });

    this.webSocketApi = websocketApi;
  }
}