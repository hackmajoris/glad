import { createReducer, on } from '@ngrx/store';
import { ThemeState, initialThemeState } from './theme.state';
import * as ThemeActions from './theme.actions';

export const themeReducer = createReducer(
  initialThemeState,
  
  on(ThemeActions.setTheme, (state, { theme }) => {
    console.log('[ThemeReducer] Setting theme to:', theme);
    return {
      ...state,
      currentTheme: theme
    };
  })
);