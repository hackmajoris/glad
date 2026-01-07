import { createAction, props } from '@ngrx/store';
import { UserListItem } from '../../core/models';

// Load Users
export const loadUsers = createAction('[Users] Load Users');
export const loadUsersSuccess = createAction(
  '[Users] Load Users Success',
  props<{ users: UserListItem[] }>()
);
export const loadUsersFailure = createAction(
  '[Users] Load Users Failure',
  props<{ error: string }>()
);

// Clear Users Error
export const clearUsersError = createAction('[Users] Clear Error');
