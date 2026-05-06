import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from './api.service';
import { User, UserListItem, UpdateUserRequest } from '../models';

@Injectable({
  providedIn: 'root'
})
export class UserService {
  constructor(private apiService: ApiService) {}

  /**
   * Get current authenticated user's profile
   */
  getCurrentUser(): Observable<User> {
    return this.apiService.get<User>('/users/me');
  }

  /**
   * Update current user's profile
   */
  updateCurrentUser(request: UpdateUserRequest): Observable<{ message: string }> {
    return this.apiService.put<{ message: string }>('/users/me', request);
  }

  /**
   * Get all users
   */
  listUsers(): Observable<UserListItem[]> {
    return this.apiService.get<UserListItem[]>('/users');
  }

  /**
   * Get user by username
   */
  getUser(username: string): Observable<User> {
    return this.apiService.get<User>(`/users/${username}`);
  }
}
