import { createAction, props } from '@ngrx/store';

export const setTheme = createAction(
  '[Theme] Set Theme',
  props<{ theme: string }>()
);

export const initializeTheme = createAction('[Theme] Initialize Theme');