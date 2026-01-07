import { UserListItem } from '../../core/models';

export interface UsersState {
  users: UserListItem[];
  loading: boolean;
  error: string | null;
  loaded: boolean;
}

export const initialUsersState: UsersState = {
  users: [],
  loading: false,
  error: null,
  loaded: false,
};
