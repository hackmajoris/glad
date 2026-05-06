import { AfterViewInit, Component, signal, PLATFORM_ID, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { Store } from '@ngrx/store';
import { isPlatformBrowser } from '@angular/common';
import { AppState, initializeAuth, initializeTheme } from './store';
import { LayoutService } from './layout/service/layout.service';
import { $t } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';
import Lara from '@primeuix/themes/lara';
import Material from '@primeuix/themes/material';

const presets = {
    Aura,
    Lara,
    Material
} as const;

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App implements AfterViewInit {
  protected readonly title = signal('glad-ui');
  private platformId = inject(PLATFORM_ID);
  private layoutService = inject(LayoutService);

  constructor(private store: Store<AppState>) {}

  ngAfterViewInit(): void {
    // Initialize authentication and theme after view init
    setTimeout(() => {
      this.store.dispatch(initializeAuth());
      this.store.dispatch(initializeTheme());

      // Initialize PrimeNG theme and layout config
      if (isPlatformBrowser(this.platformId)) {
        this.initializePrimeNGTheme();
      }
    }, 0);
  }

  private initializePrimeNGTheme(): void {
    // Load saved config from localStorage
    this.layoutService.loadFromLocalStorage();

    const config = this.layoutService.layoutConfig();
    const presetKey = (config.preset || 'Aura') as keyof typeof presets;
    const preset = presets[presetKey];

    $t()
      .preset(preset)
      .use({ useDefaultOptions: true });

    // Apply initial dark mode
    this.layoutService.toggleDarkMode(config);
  }
}
