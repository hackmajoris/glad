import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { from, switchMap } from 'rxjs';
import { AuthService } from '../services/auth.service';

/**
 * Auth Interceptor - Automatically adds Authorization header to HTTP requests
 * Note: The ApiService already handles this, but this interceptor provides
 * a fallback for any direct HttpClient usage in the application
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);

  // Skip adding auth header if request already has Authorization
  if (req.headers.has('Authorization')) {
    return next(req);
  }

  // Get ID token and add to request
  return from(authService.getIdToken()).pipe(
    switchMap(token => {
      if (token) {
        const cloned = req.clone({
          headers: req.headers.set('Authorization', `Bearer ${token}`)
        });
        return next(cloned);
      }
      return next(req);
    })
  );
};
