import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { Store } from '@ngrx/store';
import { Subject, takeUntil } from 'rxjs';
import { AppState, signIn, selectAuthLoading, selectAuthError } from '../../../store';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit, OnDestroy {
  username = '';
  password = '';
  loading = false;
  errorMessage = '';

  private destroy$ = new Subject<void>();

  constructor(
    private store: Store<AppState>
  ) {}

  ngOnInit(): void {
    // Subscribe to loading state from store
    this.store.select(selectAuthLoading)
      .pipe(takeUntil(this.destroy$))
      .subscribe(loading => {
        this.loading = loading;
      });

    // Subscribe to error state from store
    this.store.select(selectAuthError)
      .pipe(takeUntil(this.destroy$))
      .subscribe(error => {
        this.errorMessage = error || '';
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onSubmit(): void {
    if (!this.username || !this.password) {
      this.errorMessage = 'Please enter username and password';
      return;
    }

    // Clear any previous error
    this.errorMessage = '';

    // Dispatch sign in action - effects will handle the rest
    this.store.dispatch(signIn({
      username: this.username,
      password: this.password
    }));
  }
}
