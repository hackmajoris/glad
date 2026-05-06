import amplifyConfig from '../../amplify_outputs.json';

export const environment = {
  production: true,
  cognito: {
    region: amplifyConfig.auth.aws_region,
    userPoolId: amplifyConfig.auth.user_pool_id,
    userPoolClientId: amplifyConfig.auth.user_pool_client_id,
  },
  api: {
    endpoint: amplifyConfig.custom?.api?.endpoint || '',
  }
};
