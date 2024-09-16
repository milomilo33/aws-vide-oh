#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { VPCStack } from '../lib/vpc-stack';
import { RDSStack } from '../lib/rds-stack';
import { VueAppStack } from '../lib/vue-app-stack'

const app = new cdk.App();
const vpcStack = new VPCStack(app, `videoh-vpc`, {});
const rdsStack = new RDSStack(app, `videoh-db`, { vpc: vpcStack.vpc });
rdsStack.addDependency(vpcStack);
const vueAppStack = new VueAppStack(app, 'videoh-vue-appx', {
    env: {
        region: 'eu-central-1',
    },
});