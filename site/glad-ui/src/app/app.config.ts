import { ApplicationConfig, provideBrowserGlobalErrorListeners, isDevMode } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideStore } from '@ngrx/store';
import { provideEffects, Actions } from '@ngrx/effects';
import { provideStoreDevtools } from '@ngrx/store-devtools';
import { providePrimeNG } from 'primeng/config';
import Aura from '@primeuix/themes/aura';
import Lara from '@primeuix/themes/lara';
import Material from '@primeuix/themes/material';

import { routes } from './app.routes';
import { authInterceptor } from './core/interceptors';
import { reducers } from './store';

import { initializeAuthEffect, signInEffect, signInSuccessEffect, signOutEffect, signOutSuccessEffect } from './store/auth/auth.effects';

import { loadUsersEffect } from './store/users/users.effects';

import { setThemeEffect, initializeThemeEffect } from './store/theme/theme.effects';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideAnimationsAsync(),
    provideHttpClient(
      withInterceptors([authInterceptor])
    ),
    provideStore(reducers),
    provideEffects({ initializeAuthEffect, signInEffect, signInSuccessEffect, signOutEffect, signOutSuccessEffect, loadUsersEffect, setThemeEffect, initializeThemeEffect }),
    provideStoreDevtools({
      maxAge: 25,
      logOnly: !isDevMode(),
      autoPause: true,
      trace: false,
      traceLimit: 75,
    }),
    providePrimeNG({
      theme: {
        preset: Aura,
        options: {
          darkModeSelector: '.app-dark'
        }
      }
    })
  ]
};
