import { inject } from '@angular/core';
import { Router, CanActivateFn } from '@angular/router';
import { Store } from '@ngrx/store';
import { map, filter, take } from 'rxjs/operators';
import { combineLatest } from 'rxjs';
import { AppState, selectAuthInitialized, selectIsAuthenticated } from '../../store';

/**
 * Auth Guard - Protects routes that require authentication
 * Redirects to login page if user is not authenticated
 *
 * Uses NgRx store to check authentication status
 */
export const authGuard: CanActivateFn = (route, state) => {
  const store = inject(Store<AppState>);
  const router = inject(Router);

  // Wait for auth initialization to complete, then check authentication
  return combineLatest([
    store.select(selectAuthInitialized),
    store.select(selectIsAuthenticated)
  ]).pipe(
    filter(([initialized]) => initialized === true), // Wait until initialized
    take(1), // Take only the first emission after initialization
    map(([_, isAuthenticated]) => {
      if (isAuthenticated) {
        console.log('Auth guard: User is authenticated');
        return true;
      } else {
        console.log('Auth guard: User not authenticated, redirecting to login');
        // Redirect to login page with return URL
        router.navigate(['/login'], { queryParams: { returnUrl: state.url } });
        return false;
      }
    })
  );
};
