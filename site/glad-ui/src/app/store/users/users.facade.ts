import { Injectable, inject } from '@angular/core';
import { Store } from '@ngrx/store';
import { Observable } from 'rxjs';
import { AppState, loadUsers, selectAllUsers, selectUsersLoading, selectUsersError } from '../../store';
import { UserListItem } from '../../core/models';

@Injectable({
  providedIn: 'root'
})
export class UsersFacade {
  private store = inject(Store<AppState>);

  // Selectors as observables
  users$ = this.store.select(selectAllUsers);
  loading$ = this.store.select(selectUsersLoading);
  error$ = this.store.select(selectUsersError);

  // Actions
  loadUsers(): void {
    this.store.dispatch(loadUsers());
  }
}
