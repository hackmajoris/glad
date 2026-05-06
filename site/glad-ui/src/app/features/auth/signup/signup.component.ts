import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../core/services';

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './signup.component.html',
  styleUrls: ['./signup.component.css']
})
export class SignupComponent {
  username = '';
  email = '';
  password = '';
  confirmPassword = '';
  loading = false;
  errorMessage = '';

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  onSubmit(): void {
    // Validation
    if (!this.username || !this.email || !this.password || !this.confirmPassword) {
      this.errorMessage = 'Please fill in all fields';
      return;
    }

    if (this.password !== this.confirmPassword) {
      this.errorMessage = 'Passwords do not match';
      return;
    }

    if (this.password.length < 8) {
      this.errorMessage = 'Password must be at least 8 characters long';
      return;
    }

    this.loading = true;
    this.errorMessage = '';

    this.authService.signUp({
      username: this.username,
      email: this.email,
      password: this.password
    }).subscribe({
      next: (result) => {
        console.log('Signup successful:', result);

        // Navigate to confirmation page
        this.router.navigate(['/confirm-signup'], {
          queryParams: { username: this.username }
        });
      },
      error: (error) => {
        console.error('Signup error:', error);
        this.loading = false;

        // Handle specific error types
        if (error.name === 'UsernameExistsException') {
          this.errorMessage = 'Username already exists. Please choose a different username.';
        } else if (error.name === 'InvalidPasswordException') {
          this.errorMessage = 'Password does not meet requirements. Must contain uppercase, lowercase, number, and special character.';
        } else if (error.name === 'InvalidParameterException') {
          this.errorMessage = 'Invalid email address or username format.';
        } else {
          this.errorMessage = error.message || 'Signup failed. Please try again.';
        }
      },
      complete: () => {
        this.loading = false;
      }
    });
  }
}
