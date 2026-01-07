import { createReducer, on } from '@ngrx/store';
import { UsersState, initialUsersState } from './users.state';
import * as UsersActions from './users.actions';

export const usersReducer = createReducer(
  initialUsersState,

  // Load Users
  on(UsersActions.loadUsers, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(UsersActions.loadUsersSuccess, (state, { users }) => {
    console.log('[UsersReducer] loadUsersSuccess called with users:', users);
    return {
      ...state,
      users,
      loading: false,
      loaded: true,
      error: null,
    };
  }),
  on(UsersActions.loadUsersFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Clear Error
  on(UsersActions.clearUsersError, (state) => ({
    ...state,
    error: null,
  }))
);
