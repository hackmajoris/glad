export interface User {
  username: string;
  name: string;
  email: string;
  createdAt: string;
  updatedAt: string;
}

export interface UserListItem {
  username: string;
  name: string;
}

export interface UpdateUserRequest {
  name?: string;
  password?: string;
}
