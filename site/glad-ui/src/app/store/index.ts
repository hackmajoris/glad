import { ActionReducerMap } from '@ngrx/store';
import { AuthState, authReducer } from './auth';
import { UsersState, usersReducer } from './users';
import { ThemeState, themeReducer } from './theme';

export interface AppState {
  auth: AuthState;
  users: UsersState;
  theme: ThemeState;
}

export const reducers: ActionReducerMap<AppState> = {
  auth: authReducer,
  users: usersReducer,
  theme: themeReducer,
};

// Re-export everything for convenience
export * from './auth';
export * from './users';
export * from './theme';
