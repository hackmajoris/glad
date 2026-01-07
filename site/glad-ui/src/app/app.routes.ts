import { Routes } from '@angular/router';
import { authGuard, guestGuard } from './core/guards';
import { LayoutComponent } from './layout/layout.component';

export const routes: Routes = [
  // Default route - redirect to users (auth guard will redirect to login if needed)
  {
    path: '',
    redirectTo: '/users',
    pathMatch: 'full'
  },

  // Public routes - Authentication (no layout, redirects to /users if already logged in)
  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'signup',
    canActivate: [guestGuard],
    loadComponent: () => import('./features/auth/signup/signup.component').then(m => m.SignupComponent)
  },
  {
    path: 'confirm-signup',
    canActivate: [guestGuard],
    loadComponent: () => import('./features/auth/confirm-signup/confirm-signup.component').then(m => m.ConfirmSignupComponent)
  },

  // Protected routes - All wrapped in layout component
  {
    path: '',
    component: LayoutComponent,
    canActivate: [authGuard],
    children: [
      // Profile routes
      {
        path: 'profile',
        children: [
          {
            path: '',
            loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent) // TODO: Replace with ProfileViewComponent
          },
          {
            path: 'edit',
            loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent) // TODO: Replace with ProfileEditComponent
          }
        ]
      },

      // Skills routes
      {
        path: 'skills',
        children: [
          {
            path: '',
            loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent) // TODO: Replace with SkillsListComponent
          },
          {
            path: 'add',
            loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent) // TODO: Replace with SkillAddComponent
          }
        ]
      },

      // User Directory routes
      {
        path: 'users',
        children: [
          {
            path: '',
            loadComponent: () => import('./features/user-directory/user-list/user-list.component').then(m => m.UserListComponent)
          },
          {
            path: ':username',
            loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent) // TODO: Replace with UserDetailComponent
          }
        ]
      }
    ]
  },

  // Wildcard route - 404
  {
    path: '**',
    redirectTo: '/login'
  }
];
