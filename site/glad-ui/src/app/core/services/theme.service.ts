import { Injectable, inject } from '@angular/core';
import { Store } from '@ngrx/store';
import { AppState, setTheme, selectCurrentTheme } from '../../store';

@Injectable({
  providedIn: 'root'
})
export class ThemeService {
  private store = inject(Store<AppState>);

  currentTheme$ = this.store.select(selectCurrentTheme);

  availableThemes = [
    { name: 'Aura Light', value: 'aura-light-blue' },
    { name: 'Aura Dark', value: 'aura-dark-blue' },
    { name: 'Lara Light', value: 'lara-light-blue' },
    { name: 'Lara Dark', value: 'lara-dark-blue' },
    { name: 'Material Light', value: 'md-light-indigo' },
    { name: 'Material Dark', value: 'md-dark-indigo' }
  ];

  setTheme(theme: string): void {
    this.store.dispatch(setTheme({ theme }));
  }
}
