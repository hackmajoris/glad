import { inject } from '@angular/core';
import { Router, CanActivateFn } from '@angular/router';
import { Store } from '@ngrx/store';
import { map, filter, take } from 'rxjs/operators';
import { combineLatest } from 'rxjs';
import { AppState, selectAuthInitialized, selectIsAuthenticated } from '../../store';

/**
 * Guest Guard - Prevents authenticated users from accessing public auth pages
 * Redirects to /users if user is already authenticated
 *
 * Uses NgRx store to check authentication status
 */
export const guestGuard: CanActivateFn = (route, state) => {
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
        console.log('Guest guard: User is authenticated, redirecting to /users');
        // User is already logged in, redirect to users page
        router.navigate(['/users']);
        return false;
      } else {
        console.log('Guest guard: User not authenticated, allowing access');
        // User is not logged in, allow access to login/signup page
        return true;
      }
    })
  );
};
