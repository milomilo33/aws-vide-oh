#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { VPCStack } from '../lib/vpc-stack';
import { RDSStack } from '../lib/rds-stack';
import { VueAppStack } from '../lib/vue-app-stack'
import { DynamoDBStack } from '../lib/dynamodb-stack';
import { WebSocketStack } from '../lib/websocket-stack';

const app = new cdk.App();
const vpcStack = new VPCStack(app, `videoh-vpc`, {});
const rdsStack = new RDSStack(app, `videoh-db`, { vpc: vpcStack.vpc });
rdsStack.addDependency(vpcStack);
const webSocketStack = new WebSocketStack(app, `videoh-websocket`, {});
const vueAppStack = new VueAppStack(app, 'videoh-vue-app', {
    env: {
        region: 'eu-central-1',
    },
});
const dynamoDBStack = new DynamoDBStack(app, `videoh-dynamodb`, {});
dynamoDBStack.addDependency(vpcStack);