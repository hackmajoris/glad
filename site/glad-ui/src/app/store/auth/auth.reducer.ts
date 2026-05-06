import { createReducer, on } from '@ngrx/store';
import { AuthState, initialAuthState } from './auth.state';
import * as AuthActions from './auth.actions';

export const authReducer = createReducer(
  initialAuthState,

  // Initialize Auth
  on(AuthActions.initializeAuth, (state) => ({
    ...state,
    loading: true,
  })),
  on(AuthActions.initializeAuthSuccess, (state, { user }) => {
    console.log('[AuthReducer] initializeAuthSuccess called with user:', user);
    return {
      ...state,
      user,
      initialized: true,
      loading: false,
      error: null,
    };
  }),
  on(AuthActions.initializeAuthFailure, (state) => ({
    ...state,
    user: null,
    initialized: true,
    loading: false,
    error: null,
  })),

  // Sign In
  on(AuthActions.signIn, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(AuthActions.signInSuccess, (state, { user }) => {
    console.log('[AuthReducer] signInSuccess called with user:', user);
    return {
      ...state,
      user,
      loading: false,
      error: null,
    };
  }),
  on(AuthActions.signInFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Sign Out
  on(AuthActions.signOut, (state) => ({
    ...state,
    loading: true,
  })),
  on(AuthActions.signOutSuccess, (state) => ({
    ...state,
    user: null,
    loading: false,
    error: null,
  })),
  on(AuthActions.signOutFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Clear Error
  on(AuthActions.clearAuthError, (state) => ({
    ...state,
    error: null,
  }))
);
