import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule, ActivatedRoute } from '@angular/router';
import { AuthService } from '../../../core/services';

@Component({
  selector: 'app-confirm-signup',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './confirm-signup.component.html',
  styleUrls: ['./confirm-signup.component.css']
})
export class ConfirmSignupComponent implements OnInit {
  username = '';
  code = '';
  loading = false;
  errorMessage = '';
  successMessage = '';

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    // Get username from query params if available
    this.route.queryParams.subscribe(params => {
      if (params['username']) {
        this.username = params['username'];
      }
    });
  }

  onSubmit(): void {
    if (!this.username || !this.code) {
      this.errorMessage = 'Please enter username and confirmation code';
      return;
    }

    this.loading = true;
    this.errorMessage = '';
    this.successMessage = '';

    this.authService.confirmSignUp({
      username: this.username,
      code: this.code
    }).subscribe({
      next: () => {
        console.log('Email confirmed successfully');
        this.successMessage = 'Email confirmed successfully! Redirecting to login...';

        // Redirect to login after 2 seconds
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 2000);
      },
      error: (error) => {
        console.error('Confirmation error:', error);
        this.loading = false;

        // Handle specific error types
        if (error.name === 'CodeMismatchException') {
          this.errorMessage = 'Invalid confirmation code. Please check and try again.';
        } else if (error.name === 'ExpiredCodeException') {
          this.errorMessage = 'Confirmation code has expired. Please request a new one.';
        } else if (error.name === 'UserNotFoundException') {
          this.errorMessage = 'User not found. Please check the username.';
        } else {
          this.errorMessage = error.message || 'Confirmation failed. Please try again.';
        }
      },
      complete: () => {
        this.loading = false;
      }
    });
  }
}
