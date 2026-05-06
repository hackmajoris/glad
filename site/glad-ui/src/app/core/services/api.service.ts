import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable, from, switchMap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { AuthService } from './auth.service';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private readonly apiUrl = environment.api.endpoint;

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) {}

  /**
   * Create HTTP headers with Authorization token
   */
  private async createHeaders(): Promise<HttpHeaders> {
    const idToken = await this.authService.getIdToken();
    let headers = new HttpHeaders({
      'Content-Type': 'application/json'
    });

    if (idToken) {
      headers = headers.set('Authorization', `Bearer ${idToken}`);
    }

    return headers;
  }

  /**
   * GET request
   */
  get<T>(path: string, params?: HttpParams): Observable<T> {
    return from(this.createHeaders()).pipe(
      switchMap(headers =>
        this.http.get<T>(`${this.apiUrl}${path}`, { headers, params })
      )
    );
  }

  /**
   * POST request
   */
  post<T>(path: string, body: any): Observable<T> {
    return from(this.createHeaders()).pipe(
      switchMap(headers =>
        this.http.post<T>(`${this.apiUrl}${path}`, body, { headers })
      )
    );
  }

  /**
   * PUT request
   */
  put<T>(path: string, body: any): Observable<T> {
    return from(this.createHeaders()).pipe(
      switchMap(headers =>
        this.http.put<T>(`${this.apiUrl}${path}`, body, { headers })
      )
    );
  }

  /**
   * DELETE request
   */
  delete<T>(path: string): Observable<T> {
    return from(this.createHeaders()).pipe(
      switchMap(headers =>
        this.http.delete<T>(`${this.apiUrl}${path}`, { headers })
      )
    );
  }
}
