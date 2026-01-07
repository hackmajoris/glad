# GLAD Stack: Cognito + Angular Integration Plan

**Goal**: Integrate Amazon Cognito authentication and Angular frontend into GLAD Stack

**Strategy**:
- Fresh start with Cognito (no user migration)
- API Gateway Cognito Authorizer (auth validation at gateway level)
- **Angular 21 (latest)** in `/site/` directory with AWS Amplify
- **Tailwind CSS + DaisyUI** for modern, responsive UI components
- All features: Authentication UI, User Profile, Skills Management, User Directory

**Tech Stack Updates**:
- Angular 21 with standalone components
- AWS Amplify v6 for Cognito integration
- Tailwind CSS v3.4 for utility-first styling
- DaisyUI v4.12 for pre-built components (buttons, cards, forms, tables, modals)
- Theme support: Light, Dark, Cupcake (easily extensible)

---

## Architecture Changes

### Current → Target

**Authentication:**
- Current: Custom JWT (HS256) with bcrypt in Lambda
- Target: Amazon Cognito User Pool with OAuth 2.0 + PKCE

**API Authorization:**
- Current: `AuthorizationType: NONE` (JWT middleware in Lambda)
- Target: `AuthorizationType: COGNITO` (API Gateway validates tokens)

**User Storage:**
- Current: DynamoDB stores username, passwordHash, email
- Target: Cognito stores credentials, DynamoDB stores profile/skills only

**Frontend:**
- Current: None (empty `/site/` directory)
- Target: Angular 17+ SPA with Amplify Auth

### Request Flow

```
Angular App → Cognito → API Gateway (Authorizer) → Lambda → DynamoDB
    ↓            ↓              ↓                      ↓
 Amplify    ID Token    Validates Token         Extracts Claims
```

---

## Implementation Phases

### Phase 1: Cognito User Pool (Infrastructure)

**Create**: `deployments/glad/auth_stack.go`

**Resources:**
- Cognito User Pool
  - Sign-in: Username OR Email
  - Password policy: Min 8 chars, complexity required
  - Email verification: Required
  - MFA: Optional (TOTP)

- User Pool Client
  - Auth flows: USER_PASSWORD_AUTH, USER_SRP_AUTH
  - OAuth 2.0: Authorization Code + PKCE
  - Scopes: openid, email, profile
  - Token validity: ID (1h), Access (1h), Refresh (30d)

- Outputs: UserPoolId, UserPoolClientId, UserPoolArn

**Update**: `deployments/glad/main.go`
- Add `NewAuthStack()` before `NewAppStack()`

---

### Phase 2: API Gateway Cognito Authorizer

**Modify**: `deployments/glad/app_stack.go`

**Changes:**
1. Import User Pool ARN from auth stack
2. Create CognitoUserPoolsAuthorizer
3. Update all protected routes:
   - Change `AuthorizationType: NONE` → `AuthorizationType: COGNITO`
   - Attach `cognitoAuthorizer` to protected routes
4. Keep OPTIONS methods without authorization (CORS)
5. Remove `/register` and `/login` endpoints (Cognito handles auth)

**Routes to protect:**
- All routes except OPTIONS (CORS preflight)
- API Gateway validates tokens before reaching Lambda

---

### Phase 3: Lambda Cognito Integration

#### A. Create Cognito Claims Extractor

**Create**: `pkg/auth/cognito.go`

```go
type CognitoClaims struct {
    Sub      string // Cognito UUID
    Username string // cognito:username
    Email    string
}

func ExtractCognitoClaimsFromRequest(request events.APIGatewayProxyRequest) (*CognitoClaims, error)
```

Extracts user info from `request.RequestContext.Authorizer` map.

#### B. Update Main Entry Point

**Modify**: `cmd/glad/main.go`

**Changes:**
- Remove JWT middleware (`auth.RequireAuth()`) from routes
- Remove `/register` and `/login` route handlers
- All routes now rely on API Gateway authorizer

**Before:**
```go
r.Use(auth.RequireAuth())
r.POST("/register", h.Register)
r.POST("/login", h.Login)
```

**After:**
```go
// No middleware needed - API Gateway validates
// No register/login - Cognito handles it
```

#### C. Update Handlers

**Modify**: `cmd/glad/internal/handler/user_handler.go`

**Changes:**
- Replace JWT claims with Cognito claims extraction
- Use `auth.ExtractCognitoClaimsFromRequest(request)`
- Extract username from claims
- Auto-create user profile if not exists in DynamoDB

**Example (GetCurrentUser):**
```go
claims, err := auth.ExtractCognitoClaimsFromRequest(request)
username := claims.Username

user, err := h.userService.GetUser(username)
if errors.Is(err, apperrors.ErrUserNotFound) {
    // Auto-create profile from Cognito data
    user, err = h.userService.CreateUserFromCognito(claims.Username, claims.Email)
}
```

#### D. Add Service Method

**Modify**: `cmd/glad/internal/service/user_service.go`

**Add:**
```go
func (s *UserService) CreateUserFromCognito(username, email string) (*models.User, error)
```

Creates minimal DynamoDB user profile from Cognito data (username, email, timestamps).

#### E. Deprecate Custom Auth

**Mark as deprecated** (keep for reference):
- `pkg/auth/jwt.go` - Custom JWT service
- `pkg/middleware/auth.go` - JWT middleware
- Remove usage from main.go but keep files with deprecation comments

---

### Phase 4: Angular Application

#### A. Initialize Project

**Location**: `/Users/hackmajoris/GolandProjects/glad-stack/site/`

**Command:**
```bash
cd site
npx @angular/cli@latest new glad-ui --routing --style=css --standalone
```

**Configuration:**
- Angular 21 (latest) with standalone components
- Routing enabled
- CSS (for Tailwind CSS + DaisyUI integration)

#### B. Project Structure

```
site/glad-ui/
├── src/app/
│   ├── core/                           # Singleton services
│   │   ├── services/
│   │   │   ├── auth.service.ts         # Amplify Auth wrapper
│   │   │   ├── api.service.ts          # HTTP client with tokens
│   │   │   └── user.service.ts         # User API calls
│   │   ├── guards/
│   │   │   └── auth.guard.ts           # Route protection
│   │   ├── interceptors/
│   │   │   └── auth.interceptor.ts     # Add Bearer token
│   │   └── models/                     # TypeScript interfaces
│   │
│   ├── features/                       # Feature components
│   │   ├── auth/
│   │   │   ├── login/                  # Login component
│   │   │   ├── signup/                 # Signup component
│   │   │   └── confirm-signup/         # Email confirmation
│   │   ├── profile/
│   │   │   ├── profile-view/           # View profile
│   │   │   └── profile-edit/           # Edit profile
│   │   ├── skills/
│   │   │   ├── skills-list/            # List user's skills
│   │   │   ├── skill-add/              # Add skill
│   │   │   └── skill-edit/             # Edit skill
│   │   └── user-directory/
│   │       ├── user-list/              # Browse all users
│   │       └── user-detail/            # User detail page
│   │
│   └── shared/                         # Shared components
│       └── components/
│           ├── header/
│           ├── footer/
│           └── loading-spinner/
│
├── src/environments/
│   ├── environment.ts                  # Dev config
│   └── environment.prod.ts             # Prod config (CDK outputs)
│
└── angular.json
```

#### C. Install Dependencies

**Add to package.json:**
```json
{
  "dependencies": {
    "@angular/core": "^21.0.0",
    "@angular/common": "^21.0.0",
    "@angular/forms": "^21.0.0",
    "@angular/router": "^21.0.0",
    "aws-amplify": "^6.0.0",
    "@aws-amplify/ui-angular": "^5.0.0",
    "rxjs": "^7.8.0"
  },
  "devDependencies": {
    "tailwindcss": "^3.4.0",
    "daisyui": "^4.12.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0"
  }
}
```

**Install commands:**
```bash
cd site/glad-ui
npm install
npm install -D tailwindcss daisyui autoprefixer postcss
npx tailwindcss init
```

#### D. Tailwind CSS + DaisyUI Configuration

**1. Create `tailwind.config.js`:**
```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["light", "dark", "cupcake"],
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
  },
}
```

**2. Create `postcss.config.js`:**
```javascript
module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

**3. Update `src/styles.css`:**
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

**4. Update `angular.json`:**
```json
{
  "projects": {
    "glad-ui": {
      "architect": {
        "build": {
          "options": {
            "styles": [
              "src/styles.css"
            ]
          }
        }
      }
    }
  }
}
```

**DaisyUI provides:**
- Pre-built components (buttons, cards, forms, modals, navbars)
- Multiple themes (light/dark mode support)
- Tailwind CSS utility classes
- Responsive design utilities

**Example usage in components:**
```html
<!-- Button -->
<button class="btn btn-primary">Login</button>

<!-- Card -->
<div class="card bg-base-100 shadow-xl">
  <div class="card-body">
    <h2 class="card-title">User Profile</h2>
    <p>Profile content...</p>
  </div>
</div>

<!-- Form -->
<input type="text" class="input input-bordered" placeholder="Username" />
```

#### F. Environment Configuration

**File**: `src/environments/environment.prod.ts`
```typescript
export const environment = {
  production: true,
  cognito: {
    region: 'us-east-1',
    userPoolId: '${USER_POOL_ID}',        // Replaced by script
    userPoolClientId: '${CLIENT_ID}',     // Replaced by script
  },
  api: {
    endpoint: '${API_GATEWAY_URL}',       // Replaced by script
  }
};
```

#### G. Core Services

**1. AuthService** (`core/services/auth.service.ts`):
- Wraps AWS Amplify Auth
- Methods: `login()`, `signup()`, `confirmSignup()`, `logout()`, `getIdToken()`
- Maintains `currentUser$` observable

**2. ApiService** (`core/services/api.service.ts`):
- HTTP client wrapper
- Automatically adds Bearer token to requests
- Methods: `get()`, `post()`, `put()`, `delete()`

**3. AuthGuard** (`core/guards/auth.guard.ts`):
- Protects routes requiring authentication
- Redirects to `/login` if not authenticated

**4. AuthInterceptor** (`core/interceptors/auth.interceptor.ts`):
- Adds `Authorization: Bearer <token>` header to all API requests

#### H. Routing

**File**: `src/app/app.routes.ts`

**Public routes:**
- `/login` - Login page
- `/signup` - Signup page
- `/confirm-signup` - Email confirmation

**Protected routes** (with `canActivate: [authGuard]`):
- `/profile` - User profile
- `/profile/edit` - Edit profile
- `/skills` - Skills list
- `/skills/add` - Add skill
- `/users` - User directory
- `/users/:username` - User detail

#### I. Component Examples with DaisyUI

**Login Component Template:**
```html
<div class="hero min-h-screen bg-base-200">
  <div class="hero-content flex-col">
    <div class="text-center lg:text-left">
      <h1 class="text-5xl font-bold">Login to GLAD Stack</h1>
      <p class="py-6">Welcome back! Please login to continue.</p>
    </div>
    <div class="card flex-shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
      <form class="card-body" (ngSubmit)="onSubmit()" #loginForm="ngForm">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Username or Email</span>
          </label>
          <input
            type="text"
            placeholder="username"
            class="input input-bordered"
            [(ngModel)]="username"
            name="username"
            required
          />
        </div>
        <div class="form-control">
          <label class="label">
            <span class="label-text">Password</span>
          </label>
          <input
            type="password"
            placeholder="password"
            class="input input-bordered"
            [(ngModel)]="password"
            name="password"
            required
          />
          <label class="label">
            <a href="#" class="label-text-alt link link-hover">Forgot password?</a>
          </label>
        </div>

        <div *ngIf="errorMessage" class="alert alert-error">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{{ errorMessage }}</span>
        </div>

        <div class="form-control mt-6">
          <button
            type="submit"
            class="btn btn-primary"
            [disabled]="!loginForm.form.valid || loading"
          >
            <span *ngIf="loading" class="loading loading-spinner"></span>
            {{ loading ? 'Logging in...' : 'Login' }}
          </button>
        </div>
      </form>
      <div class="divider">OR</div>
      <div class="card-body pt-0">
        <a routerLink="/signup" class="btn btn-outline">Create Account</a>
      </div>
    </div>
  </div>
</div>
```

**Skills List Component Template:**
```html
<div class="container mx-auto p-4">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-3xl font-bold">My Skills</h1>
    <a routerLink="/skills/add" class="btn btn-primary">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
      Add Skill
    </a>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    <div *ngFor="let skill of skills" class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <h2 class="card-title">
          {{ skill.skillName }}
          <div class="badge badge-secondary">{{ skill.proficiencyLevel }}</div>
        </h2>
        <p>{{ skill.yearsOfExperience }} years of experience</p>
        <div class="card-actions justify-end">
          <button class="btn btn-ghost btn-sm" (click)="editSkill(skill)">Edit</button>
          <button class="btn btn-error btn-sm" (click)="deleteSkill(skill)">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <div *ngIf="skills.length === 0" class="alert alert-info">
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
    </svg>
    <span>You haven't added any skills yet. Click "Add Skill" to get started!</span>
  </div>
</div>
```

**User Directory Component Template:**
```html
<div class="container mx-auto p-4">
  <h1 class="text-3xl font-bold mb-6">User Directory</h1>

  <!-- Search Bar -->
  <div class="form-control mb-6">
    <div class="input-group">
      <input
        type="text"
        placeholder="Search users by skill..."
        class="input input-bordered w-full"
        [(ngModel)]="searchTerm"
        (ngModelChange)="onSearch()"
      />
      <button class="btn btn-square">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </button>
    </div>
  </div>

  <!-- User Cards -->
  <div class="overflow-x-auto">
    <table class="table table-zebra w-full">
      <thead>
        <tr>
          <th>Username</th>
          <th>Email</th>
          <th>Skills</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr *ngFor="let user of users">
          <td>
            <div class="flex items-center space-x-3">
              <div class="avatar placeholder">
                <div class="bg-neutral-focus text-neutral-content rounded-full w-12">
                  <span class="text-xl">{{ user.username.charAt(0).toUpperCase() }}</span>
                </div>
              </div>
              <div>
                <div class="font-bold">{{ user.username }}</div>
              </div>
            </div>
          </td>
          <td>{{ user.email }}</td>
          <td>
            <div class="badge badge-primary badge-outline" *ngFor="let skill of user.skills?.slice(0, 3)">
              {{ skill }}
            </div>
            <div class="badge badge-ghost" *ngIf="user.skills?.length > 3">
              +{{ user.skills.length - 3 }} more
            </div>
          </td>
          <td>
            <button class="btn btn-ghost btn-xs" [routerLink]="['/users', user.username]">
              View Profile
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
```

**Navbar Component Template:**
```html
<div class="navbar bg-base-100 shadow-lg">
  <div class="flex-1">
    <a routerLink="/" class="btn btn-ghost normal-case text-xl">GLAD Stack</a>
  </div>
  <div class="flex-none">
    <ul class="menu menu-horizontal px-1">
      <li><a routerLink="/profile">Profile</a></li>
      <li><a routerLink="/skills">Skills</a></li>
      <li><a routerLink="/users">Users</a></li>
    </ul>
    <div class="dropdown dropdown-end">
      <label tabindex="0" class="btn btn-ghost btn-circle avatar placeholder">
        <div class="bg-neutral-focus text-neutral-content rounded-full w-10">
          <span>{{ username?.charAt(0).toUpperCase() }}</span>
        </div>
      </label>
      <ul tabindex="0" class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
        <li><a routerLink="/profile">My Profile</a></li>
        <li><a routerLink="/profile/edit">Settings</a></li>
        <li><a (click)="onLogout()">Logout</a></li>
      </ul>
    </div>
  </div>
</div>
```

**Theme Switcher (Optional):**
```html
<!-- Add to navbar or footer -->
<div class="dropdown dropdown-end">
  <label tabindex="0" class="btn btn-ghost">
    <svg class="fill-current w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
      <path d="M7.5 2.25A.75.75 0 0 1 8.25 3v1.5h7.5V3a.75.75 0 0 1 1.5 0v1.5h1.5A2.25 2.25 0 0 1 21 6.75v12A2.25 2.25 0 0 1 18.75 21H5.25A2.25 2.25 0 0 1 3 18.75v-12A2.25 2.25 0 0 1 5.25 4.5h1.5V3a.75.75 0 0 1 .75-.75Z"/>
    </svg>
  </label>
  <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-52">
    <li><a data-set-theme="light">Light</a></li>
    <li><a data-set-theme="dark">Dark</a></li>
    <li><a data-set-theme="cupcake">Cupcake</a></li>
  </ul>
</div>
```

**DaisyUI Components Used:**
- **hero** - Full-screen hero sections
- **card** - Content cards with shadow
- **form-control** - Form field containers
- **input** - Styled input fields with variants
- **btn** - Buttons with variants (primary, outline, ghost, error)
- **alert** - Alert messages with icons
- **badge** - Small status indicators
- **table** - Styled data tables
- **navbar** - Navigation bars
- **dropdown** - Dropdown menus
- **avatar** - User avatars with placeholder
- **loading** - Loading spinners
- **divider** - Section dividers
- **menu** - Menu lists

---

### Phase 5: Frontend Hosting Infrastructure

**Create**: `deployments/glad/frontend_stack.go`

**Resources:**
1. **S3 Bucket**
   - Private bucket (no public access)
   - Website configuration: index.html, error.html → index.html (SPA fallback)
   - Auto-delete objects on stack destroy

2. **CloudFront Distribution**
   - Origin: S3 bucket with OAI (Origin Access Identity)
   - Default behavior: Compress, cache static assets
   - Error responses: 404/403 → 200 /index.html (SPA routing)
   - Default root object: index.html

3. **Outputs**
   - CloudFront URL
   - S3 bucket name (for deployment)

**Update**: `deployments/glad/main.go`
- Add `NewFrontendStack()` after app stack

---

### Phase 6: Build & Deploy Automation

#### A. Update Root Taskfile

**Modify**: `Taskfile.yml`

**Add tasks:**
```yaml
frontend:install:
  desc: 'Install Angular dependencies'
  dir: site/glad-ui
  cmds: [npm install]

frontend:serve:
  desc: 'Serve Angular locally'
  dir: site/glad-ui
  cmds: [npm start]

frontend:build:
  desc: 'Build Angular for production'
  dir: site/glad-ui
  cmds: [npm run build -- --configuration production]

frontend:deploy:
  desc: 'Deploy Angular to S3'
  cmds:
    - task: frontend:build
    - aws s3 sync site/glad-ui/dist/glad-ui s3://$(aws cloudformation describe-stacks --stack-name glad-frontend-stack-production --query "Stacks[0].Outputs[?OutputKey=='WebsiteBucketName'].OutputValue" --output text) --delete
    - aws cloudfront create-invalidation --distribution-id $(DIST_ID) --paths "/*"

deploy:full:
  desc: 'Deploy entire stack'
  cmds:
    - task: glad:test
    - task: glad:cdk:deploy
    - task: frontend:deploy

sync:config:
  desc: 'Sync CDK outputs to Angular environment'
  cmds: [scripts/sync-cdk-outputs.sh]
```

#### B. Create CDK Output Sync Script

**Create**: `scripts/sync-cdk-outputs.sh`

**Purpose**: Fetch CloudFormation outputs and update Angular `environment.prod.ts`

**Fetches:**
- User Pool ID from auth stack
- User Pool Client ID from auth stack
- API Gateway URL from app stack
- AWS Region

**Writes to**: `site/glad-ui/src/environments/environment.prod.ts`

---

### Phase 7: Testing Updates

#### A. Backend Integration Tests

**Modify**: `cmd/glad/integration_test.go`

**Add:**
1. Test Cognito claims extraction
   - Mock API Gateway request with authorizer context
   - Verify claims parsing

2. Test user profile auto-creation
   - Simulate first request from Cognito user
   - Verify DynamoDB profile creation

3. Test protected endpoints
   - Mock Cognito authorizer context
   - Verify handler receives correct username

#### B. Frontend Testing

**Unit tests**: `ng test`
**E2E tests** (optional): Playwright or Cypress

**Manual test checklist:**
- [ ] Signup with email
- [ ] Email verification
- [ ] Login with username/email
- [ ] Protected routes redirect to login
- [ ] API calls include Bearer token
- [ ] Profile CRUD operations
- [ ] Skills management
- [ ] User directory

---

## Deployment Sequence

### Initial Deployment

**Step 1: Deploy Auth Stack**
```bash
cd deployments/glad
cdk deploy glad-auth-stack-production
```
Output: User Pool ID, Client ID

**Step 2: Deploy App Stack**
```bash
cdk deploy glad-app-stack-production
```
Output: API Gateway URL with Cognito authorizer

**Step 3: Deploy Frontend Stack**
```bash
cdk deploy glad-frontend-stack-production
```
Output: CloudFront URL, S3 bucket name

**Step 4: Sync Configuration**
```bash
cd ../..
task sync:config
```
Updates Angular environment with CDK outputs

**Step 5: Build & Deploy Frontend**
```bash
task frontend:deploy
```
Builds Angular and uploads to S3

**Step 6: Update Cognito Callback URLs**
- Add CloudFront URL to User Pool Client callback URLs
- Update CORS in API Gateway to include CloudFront URL

**Step 7: Test**
- Open CloudFront URL
- Sign up new user
- Confirm email
- Login and test all features

### Subsequent Deployments

**Backend only:**
```bash
task glad:cdk:deploy
```

**Frontend only:**
```bash
task frontend:deploy
```

**Full stack:**
```bash
task deploy:full
```

---

## Critical Files to Create/Modify

### Infrastructure (CDK)
1. ✅ Create `deployments/glad/auth_stack.go` - Cognito User Pool
2. ✅ Modify `deployments/glad/app_stack.go` - Add Cognito authorizer, update routes
3. ✅ Modify `deployments/glad/main.go` - Add auth stack
4. ✅ Create `deployments/glad/frontend_stack.go` - S3 + CloudFront

### Backend (Go)
5. ✅ Create `pkg/auth/cognito.go` - Claims extractor
6. ✅ Modify `cmd/glad/main.go` - Remove JWT middleware, update router
7. ✅ Modify `cmd/glad/internal/handler/user_handler.go` - Use Cognito claims
8. ✅ Modify `cmd/glad/internal/service/user_service.go` - Add `CreateUserFromCognito()`
9. ✅ Deprecate `pkg/auth/jwt.go` - Mark as deprecated
10. ✅ Deprecate `pkg/middleware/auth.go` - Mark as deprecated

### Frontend (Angular 21 + Tailwind CSS + DaisyUI)
11. ✅ Initialize `site/glad-ui/` - Angular 21 project with standalone components
12. ✅ Create `tailwind.config.js` - Tailwind CSS + DaisyUI configuration
13. ✅ Create `postcss.config.js` - PostCSS configuration
14. ✅ Update `src/styles.css` - Import Tailwind directives
15. ✅ Create `src/app/core/services/auth.service.ts` - Amplify Auth wrapper
16. ✅ Create `src/app/core/services/api.service.ts` - HTTP client with token injection
17. ✅ Create `src/app/core/guards/auth.guard.ts` - Route protection guard
18. ✅ Create `src/app/core/interceptors/auth.interceptor.ts` - Authorization interceptor
19. ✅ Create `src/app/app.routes.ts` - Routing configuration with lazy loading
20. ✅ Create `src/environments/environment.prod.ts` - Production environment config
21. ✅ Create feature components with DaisyUI styling:
    - auth/login - Login form with hero section
    - auth/signup - Signup form with validation
    - profile/profile-view - User profile card
    - profile/profile-edit - Editable profile form
    - skills/skills-list - Skills grid with cards
    - skills/skill-add - Add skill form
    - user-directory/user-list - User table with search
    - shared/navbar - Navigation bar with dropdown menu

### Build & Deploy
22. ✅ Modify `Taskfile.yml` - Add frontend tasks (install, serve, build, deploy)
23. ✅ Create `scripts/sync-cdk-outputs.sh` - Sync CDK outputs to Angular environment

### Testing
24. ✅ Modify `cmd/glad/integration_test.go` - Add Cognito integration tests

---

## Key Design Decisions

### 1. User Identity Mapping
**Decision**: Use Cognito username (not sub UUID)
**Rationale**: Simpler mapping to existing DynamoDB design, human-readable
**Trade-off**: Username changes not supported

### 2. Authentication Flow
**Decision**: API Gateway Cognito Authorizer
**Rationale**: AWS best practice, removes auth code from Lambda
**Trade-off**: Slightly higher latency (~100ms), harder to test locally

### 3. User Profile Storage
**Decision**: Auto-create DynamoDB profile on first API call
**Rationale**: Seamless user experience, no extra onboarding step
**Trade-off**: Less control over profile initialization

### 4. JWT Middleware
**Decision**: Deprecate but keep for reference
**Rationale**: Clean production code, but useful for understanding migration
**Trade-off**: Dead code in repository

### 5. /register and /login Endpoints
**Decision**: Remove completely
**Rationale**: Fresh start, Cognito handles all auth, cleaner architecture
**Trade-off**: No backward compatibility

---

## Security Considerations

**Cognito:**
- Password policy enforced (8+ chars, complexity)
- Email verification required
- MFA optional (can enable later)
- Advanced security features (compromised credentials check)

**API Gateway:**
- Cognito authorizer validates tokens
- Throttling: 100 RPS, 200 burst
- CORS properly configured
- CloudWatch logging enabled

**Frontend:**
- HTTPS only (CloudFront enforces)
- Tokens handled by Amplify (secure storage)
- CSP headers recommended
- Input sanitization

**Lambda:**
- Least privilege IAM (DynamoDB access only)
- Validates inputs despite API Gateway validation
- Structured logging (no sensitive data)

---

## Cost Estimate

**Monthly costs** (after free tier):
- Cognito: $0 (under 50K MAU)
- API Gateway: ~$0.35
- Lambda: ~$1-2
- DynamoDB: ~$1-3 (on-demand)
- S3: ~$0.10
- CloudFront: ~$1-2
**Total: ~$3-8/month** for moderate traffic

---

## Success Criteria

- [ ] Users can sign up with email verification
- [ ] Users can login with username or email
- [ ] API Gateway validates Cognito tokens
- [ ] Protected routes accessible only when authenticated
- [ ] User profile auto-created from Cognito data
- [ ] All CRUD operations work (profile, skills)
- [ ] User directory shows all users
- [ ] Frontend loads from CloudFront
- [ ] SPA routing works (no 404s)
- [ ] Logout clears session properly
- [ ] Tests pass (integration + unit)

---

## Next Steps After Approval

1. Start with Phase 1 (Cognito infrastructure)
2. Test each phase before moving to next
3. Use feature branch for development
4. Update documentation as we go
5. Create PR when complete
