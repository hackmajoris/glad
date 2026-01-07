import { Component, OnInit, OnDestroy, signal, HostListener, ElementRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, RouterLinkActive } from '@angular/router';
import { Store } from '@ngrx/store';
import { Subject, takeUntil } from 'rxjs';
import { AppState, signOut, selectCurrentUser } from '../../../store';
import { LayoutService } from '../../../layout/service/layout.service';
import { ButtonModule } from 'primeng/button';
import { AvatarModule } from 'primeng/avatar';
import { MenuModule } from 'primeng/menu';
import { MenuItem } from 'primeng/api';
import { StyleClassModule } from 'primeng/styleclass';
import { CognitoUser } from '../../../core';
import { AppConfigurator } from '../../../layout/component/app.configurator';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ButtonModule,
    AvatarModule,
    MenuModule,
    StyleClassModule,
    AppConfigurator,
  ],
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit, OnDestroy {

  currentUser: CognitoUser | null = null;
  userMenuItems: MenuItem[] = [];
  isConfiguratorVisible = signal(false);
  private destroy$ = new Subject<void>();

  constructor(
    private store: Store<AppState>,
    public layoutService: LayoutService,
    private elementRef: ElementRef
  ) {}

  ngOnInit(): void {
    this.store.select(selectCurrentUser)
      .pipe(takeUntil(this.destroy$))
      .subscribe(user => {
        this.currentUser = user;
        this.initializeUserMenu();
      });
  }

  initializeUserMenu(): void {
    this.userMenuItems = [
      {
        label: this.currentUser?.username || 'User',
        disabled: true
      },
      {
        separator: true
      },
      {
        label: 'My Profile',
        icon: 'pi pi-user',
        routerLink: '/profile'
      },
      {
        label: 'Settings',
        icon: 'pi pi-cog',
        routerLink: '/profile/edit'
      },
      {
        separator: true
      },
      {
        label: 'Logout',
        icon: 'pi pi-sign-out',
        command: () => this.onLogout()
      }
    ];
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onMenuButtonClick(): void {
    this.layoutService.onMenuToggle();
  }

  onLogout(): void {
    this.store.dispatch(signOut());
  }

  getUserInitial(): string {
    return this.currentUser?.username?.charAt(0).toUpperCase() || 'U';
  }

  toggleConfigurator(): void {
    this.isConfiguratorVisible.set(!this.isConfiguratorVisible());
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    if (this.isConfiguratorVisible()) {
      const clickedInside = this.elementRef.nativeElement.contains(event.target);
      if (!clickedInside) {
        this.isConfiguratorVisible.set(false);
      }
    }
  }

  toggleDarkMode(): void {
    this.layoutService.layoutConfig.update((state) => ({ ...state, darkTheme: !state.darkTheme }));
  }

  isDarkMode(): boolean {
    return this.layoutService.layoutConfig().darkTheme || false;
  }
}
