import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Subject } from 'rxjs';
import { Observable } from 'rxjs';
import { UserListItem } from '../../../core/models';
import { UsersFacade } from '../../../store';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { AvatarModule } from 'primeng/avatar';
import { MessageModule } from 'primeng/message';
import { ProgressSpinnerModule } from 'primeng/progressspinner';

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [CommonModule, RouterModule, ButtonModule, CardModule, AvatarModule, MessageModule, ProgressSpinnerModule],
  templateUrl: './user-list.component.html',
  styleUrls: ['./user-list.component.css']
})
export class UserListComponent implements OnInit, OnDestroy {
  users$!: Observable<UserListItem[]>;
  loading$!: Observable<boolean>;
  error$!: Observable<string | null>;

  private destroy$ = new Subject<void>();

  constructor(private usersFacade: UsersFacade) {}

  ngOnInit(): void {
    console.log('[UserListComponent] ngOnInit called');
    this.users$ = this.usersFacade.users$;
    this.loading$ = this.usersFacade.loading$;
    this.error$ = this.usersFacade.error$;
    this.usersFacade.loadUsers();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  refreshUsers(): void {
    this.usersFacade.loadUsers();
  }
}
