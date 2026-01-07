import { CognitoUser } from '../../core/models';

export interface AuthState {
  user: CognitoUser | null;
  initialized: boolean;
  loading: boolean;
  error: string | null;
}

export const initialAuthState: AuthState = {
  user: null,
  initialized: false,
  loading: false,
  error: null,
};
