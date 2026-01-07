import amplifyConfig from '../../amplify_outputs.json';

export const environment = {
  production: false,
  cognito: {
    region: amplifyConfig.auth.aws_region,
    userPoolId: amplifyConfig.auth.user_pool_id,
    userPoolClientId: amplifyConfig.auth.user_pool_client_id,
  },
  api: {
    // Use production API endpoint for local development
    // To use local backend, change this to 'http://localhost:3000'
    endpoint: amplifyConfig.custom?.api?.endpoint || 'http://localhost:3000',
  }
};
