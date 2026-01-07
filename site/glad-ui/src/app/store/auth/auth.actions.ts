import { createAction, props } from '@ngrx/store';
import { CognitoUser } from '../../core/models';

// Initialize Auth
export const initializeAuth = createAction('[Auth] Initialize Auth');
export const initializeAuthSuccess = createAction(
  '[Auth] Initialize Auth Success',
  props<{ user: CognitoUser | null }>()
);
export const initializeAuthFailure = createAction('[Auth] Initialize Auth Failure');

// Sign In
export const signIn = createAction(
  '[Auth] Sign In',
  props<{ username: string; password: string }>()
);
export const signInSuccess = createAction(
  '[Auth] Sign In Success',
  props<{ user: CognitoUser }>()
);
export const signInFailure = createAction(
  '[Auth] Sign In Failure',
  props<{ error: string }>()
);

// Sign Out
export const signOut = createAction('[Auth] Sign Out');
export const signOutSuccess = createAction('[Auth] Sign Out Success');
export const signOutFailure = createAction(
  '[Auth] Sign Out Failure',
  props<{ error: string }>()
);

// Clear Auth Error
export const clearAuthError = createAction('[Auth] Clear Error');
