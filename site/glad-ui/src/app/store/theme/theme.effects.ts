import { inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { tap } from 'rxjs/operators';
import { PrimeNG } from 'primeng/config';
import Aura from '@primeuix/themes/aura';
import Lara from '@primeuix/themes/lara';
import Material from '@primeuix/themes/material';
import * as ThemeActions from './theme.actions';

// Theme preset mapping
const THEME_PRESETS: Record<string, any> = {
  'aura-light-blue': Aura,
  'aura-dark-blue': Aura,
  'lara-light-blue': Lara,
  'lara-dark-blue': Lara,
  'md-light-indigo': Material,
  'md-dark-indigo': Material
};

export const setThemeEffect = createEffect(
  () => {
    const actions$ = inject(Actions);
    const primeng = inject(PrimeNG);

    return actions$.pipe(
      ofType(ThemeActions.setTheme),
      tap(({ theme }) => {
        console.log('[ThemeEffect] Setting theme to:', theme);

        // Determine if dark mode based on theme name
        const isDark = theme.includes('dark');
        const preset = THEME_PRESETS[theme] || Aura;

        // Update the document class for dark mode
        if (isDark) {
          document.documentElement.classList.add('dark-mode');
        } else {
          document.documentElement.classList.remove('dark-mode');
        }

        // Update PrimeNG theme preset
        primeng.theme.set({
          preset: preset,
          options: {
            darkModeSelector: '.dark-mode'
          }
        });

        localStorage.setItem('theme', theme);
      })
    );
  },
  { functional: true, dispatch: false }
);

export const initializeThemeEffect = createEffect(
  () => {
    const actions$ = inject(Actions);
    const primeng = inject(PrimeNG);

    return actions$.pipe(
      ofType(ThemeActions.initializeTheme),
      tap(() => {
        const savedTheme = localStorage.getItem('theme') || 'aura-light-blue';
        const isDark = savedTheme.includes('dark');
        const preset = THEME_PRESETS[savedTheme] || Aura;

        // Update the document class for dark mode
        if (isDark) {
          document.documentElement.classList.add('dark-mode');
        } else {
          document.documentElement.classList.remove('dark-mode');
        }

        // Initialize PrimeNG theme preset
        primeng.theme.set({
          preset: preset,
          options: {
            darkModeSelector: '.dark-mode'
          }
        });
      })
    );
  },
  { functional: true, dispatch: false }
);