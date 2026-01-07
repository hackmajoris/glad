import { inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, catchError, switchMap } from 'rxjs/operators';
import * as UsersActions from './users.actions';
import { UserService } from '../../core/services';

export const loadUsersEffect = createEffect(
  () => {
    const actions$ = inject(Actions);
    const userService = inject(UserService);
    return actions$.pipe(
      ofType(UsersActions.loadUsers),
      switchMap(() => {
        console.log('[UsersEffects] Loading users...');
        return userService.listUsers().pipe(
          map((users) => {
            console.log('[UsersEffects] Users loaded successfully:', users);
            return UsersActions.loadUsersSuccess({ users });
          }),
          catchError((error) => {
            console.error('[UsersEffects] Error loading users:', error);
            console.error('[UsersEffects] Error details:', {
              message: error.message,
              status: error.status,
              statusText: error.statusText,
              url: error.url
            });
            return of(UsersActions.loadUsersFailure({
              error: error.message || 'Failed to load users'
            }));
          })
        );
      })
    );
  },
  { functional: true }
);
