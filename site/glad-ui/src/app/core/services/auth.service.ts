import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, from } from 'rxjs';
import { Amplify } from 'aws-amplify';
import {
  signUp,
  signIn,
  signOut,
  confirmSignUp,
  getCurrentUser,
  fetchAuthSession,
  SignUpOutput,
  SignInOutput
} from 'aws-amplify/auth';
import { CognitoUser, SignUpRequest, SignInRequest, ConfirmSignUpRequest } from '../models';
import amplifyConfig from '../../../../amplify_outputs.json';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private currentUserSubject = new BehaviorSubject<CognitoUser | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  private authInitializedSubject = new BehaviorSubject<boolean>(false);
  public authInitialized$ = this.authInitializedSubject.asObservable();

  constructor() {
    this.configureAmplify();
    this.initializeAuth();
  }

  private configureAmplify(): void {
    // Check if email is a username attribute or if email verification is enabled
    const emailLoginEnabled = amplifyConfig.auth.username_attributes.length === 0 ||
      (amplifyConfig.auth.username_attributes as string[]).includes('email') ||
      (amplifyConfig.auth.user_verification_types as string[]).includes('email');

    Amplify.configure({
      Auth: {
        Cognito: {
          userPoolId: amplifyConfig.auth.user_pool_id,
          userPoolClientId: amplifyConfig.auth.user_pool_client_id,
          loginWith: {
            email: emailLoginEnabled,
            username: true
          }
        }
      }
    });
  }

  private async initializeAuth(): Promise<void> {
    try {
      const user = await getCurrentUser();
      const session = await fetchAuthSession();

      if (user && session.tokens) {
        this.currentUserSubject.next({
          username: user.username,
          email: session.tokens.idToken?.payload['email'] as string || '',
          sub: user.userId
        });
      } else {
        this.currentUserSubject.next(null);
      }
    } catch (error) {
      // User not authenticated
      this.currentUserSubject.next(null);
    } finally {
      // Mark auth as initialized
      this.authInitializedSubject.next(true);
    }
  }

  /**
   * Sign up a new user with Cognito
   */
  signUp(request: SignUpRequest): Observable<SignUpOutput> {
    return from(
      signUp({
        username: request.username,
        password: request.password,
        options: {
          userAttributes: {
            email: request.email
          }
        }
      })
    );
  }

  /**
   * Confirm user sign up with verification code
   */
  confirmSignUp(request: ConfirmSignUpRequest): Observable<void> {
    return from(
      confirmSignUp({
        username: request.username,
        confirmationCode: request.code
      }).then(() => undefined)
    );
  }

  /**
   * Sign in a user
   */
  signIn(request: SignInRequest): Observable<SignInOutput> {
    return from(
      signIn({
        username: request.username,
        password: request.password
      }).then(async (result) => {
        // Update current user after successful sign in
        await this.initializeAuth();
        return result;
      })
    );
  }

  /**
   * Sign out the current user
   */
  signOut(): Observable<void> {
    return from(
      signOut().then(() => {
        this.currentUserSubject.next(null);
      })
    );
  }

  /**
   * Get the current authenticated user
   */
  getCurrentUser(): CognitoUser | null {
    return this.currentUserSubject.value;
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return this.currentUserSubject.value !== null;
  }

  /**
   * Get the ID token for making API requests
   */
  async getIdToken(): Promise<string | null> {
    try {
      const session = await fetchAuthSession();
      return session.tokens?.idToken?.toString() || null;
    } catch (error) {
      console.error('Error fetching ID token:', error);
      return null;
    }
  }

  /**
   * Get the access token
   */
  async getAccessToken(): Promise<string | null> {
    try {
      const session = await fetchAuthSession();
      return session.tokens?.accessToken?.toString() || null;
    } catch (error) {
      console.error('Error fetching access token:', error);
      return null;
    }
  }
}
