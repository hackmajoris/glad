export interface CognitoUser {
  username: string;
  email: string;
  sub: string;  // Cognito user ID
}

export interface SignUpRequest {
  username: string;
  email: string;
  password: string;
}

export interface ConfirmSignUpRequest {
  username: string;
  code: string;
}

export interface SignInRequest {
  username: string;
  password: string;
}

export interface AuthTokens {
  accessToken: string;
  idToken: string;
  refreshToken: string;
}
