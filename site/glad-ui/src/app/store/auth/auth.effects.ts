import { inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Router } from '@angular/router';
import { switchMap, tap } from 'rxjs/operators';
import {
  getCurrentUser,
  fetchAuthSession,
  signIn as amplifySignIn,
  signOut as amplifySignOut
} from 'aws-amplify/auth';
import * as AuthActions from './auth.actions';
import { CognitoUser } from '../../core/models';

export const initializeAuthEffect = createEffect(
  () => {
    const actions$ = inject(Actions);
    return actions$.pipe(
      ofType(AuthActions.initializeAuth),
      switchMap(async () => {
        console.log('[AuthEffects] Initializing auth...');
        try {
          const user = await getCurrentUser();
          const session = await fetchAuthSession();

          if (user && session.tokens) {
            const cognitoUser: CognitoUser = {
              username: user.username,
              email: session.tokens.idToken?.payload['email'] as string || '',
              sub: user.userId
            };
            console.log('[AuthEffects] Auth initialized successfully, user:', cognitoUser);
            return AuthActions.initializeAuthSuccess({ user: cognitoUser });
          }
          console.log('[AuthEffects] No authenticated user found');
          return AuthActions.initializeAuthSuccess({ user: null });
        } catch (error) {
          console.log('[AuthEffects] User not authenticated:', error);
          return AuthActions.initializeAuthFailure();
        }
      })
    );
  },
  { functional: true }
);

export const signInEffect = createEffect(
  (actions$ = inject(Actions)) =>
    actions$.pipe(
      ofType(AuthActions.signIn),
      switchMap(({ username, password }) =>
        amplifySignIn({ username, password }).then(async (result) => {
          const user = await getCurrentUser();
          const session = await fetchAuthSession();

          const cognitoUser: CognitoUser = {
            username: user.username,
            email: session.tokens?.idToken?.payload['email'] as string || '',
            sub: user.userId
          };

          return AuthActions.signInSuccess({ user: cognitoUser });
        }).catch((error) => {
          let errorMessage = 'Login failed. Please try again.';

          if (error.name === 'UserNotConfirmedException') {
            errorMessage = 'Please confirm your email before logging in.';
          } else if (error.name === 'NotAuthorizedException' || error.name === 'UserNotFoundException') {
            errorMessage = 'Invalid username or password.';
          } else if (error.message) {
            errorMessage = error.message;
          }

          return AuthActions.signInFailure({ error: errorMessage });
        })
      )
    ),
  { functional: true }
);

export const signInSuccessEffect = createEffect(
  (actions$ = inject(Actions), router = inject(Router)) =>
    actions$.pipe(
      ofType(AuthActions.signInSuccess),
      tap(() => {
        const returnUrl = router.routerState.snapshot.root.queryParams['returnUrl'] || '/users';
        router.navigate([returnUrl]);
      })
    ),
  { functional: true, dispatch: false }
);

export const signOutEffect = createEffect(
  (actions$ = inject(Actions)) =>
    actions$.pipe(
      ofType(AuthActions.signOut),
      switchMap(() =>
        amplifySignOut().then(() => {
          return AuthActions.signOutSuccess();
        }).catch((error) => {
          return AuthActions.signOutFailure({ error: error.message || 'Sign out failed' });
        })
      )
    ),
  { functional: true }
);

export const signOutSuccessEffect = createEffect(
  (actions$ = inject(Actions), router = inject(Router)) =>
    actions$.pipe(
      ofType(AuthActions.signOutSuccess),
      tap(() => {
        router.navigate(['/login']);
      })
    ),
  { functional: true, dispatch: false }
);
