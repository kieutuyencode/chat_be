# Chat Backend

A Go backend application for real-time chat, featuring email-based authentication, conversations, messaging, file uploads, and WebSocket (SignalR) support.

## Features

- **User Authentication**: Email sign-in with OTP verification sent via email; JWT access tokens for protected routes
- **User Profile**: Get and update profile (fullname, phone, avatar)
- **Conversations**: Create or load 1:1 conversations, list conversations with pagination and search
- **Messages**: Send text and media messages; list messages with pagination; real-time delivery via WebSocket
- **Real-time (WebSocket)**: SignalR hub for presence (online users), message broadcasting, and connection lifecycle
- **File Upload**: Multipart upload for attachments; serve files by path
- **Email Notifications**: SMTP mail with HTML templates (e.g. sign-in verification code)
- **Database**: Ent ORM with PostgreSQL; Atlas for schema migrations

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Iris v12
- **Database**: PostgreSQL with Ent ORM
- **Migrations**: Atlas
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Config**: Viper, env files
- **DI / App lifecycle**: Uber FX
- **Logging**: Uber Zap, Lumberjack (file rotate)
- **Validation**: go-playground/validator
- **Real-time**: SignalR (philippseith/signalr)
- **Mail**: go-mail (SMTP, HTML)

## Getting Started

### Prerequisites

- **Go** 1.24 or later
- **PostgreSQL** 15 (or compatible)
- **Atlas** (for migrations): [install](https://atlasgo.io/getting-started#installation)

### Installation

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd chat_be
   ```

2. **Install Go dependencies:**

   ```bash
   go mod download
   ```

3. **Set up environment variables:**

   Create a `.env` file in the project root:

   ```env
   PORT=3000

   DB_URL=postgres://root:secret@localhost:5434/chat_be?search_path=public&sslmode=disable

   JWT_ACCESS_TOKEN_SECRET_KEY=your_jwt_secret
   JWT_ACCESS_TOKEN_EXPIRES_IN=168h

   MAIL_HOST=smtp.example.com
   MAIL_PORT=587
   MAIL_USER=your_mail_user
   MAIL_PASSWORD=your_mail_password
   ```

4. **Start PostgreSQL (e.g. with Docker):**

   ```bash
   docker compose up -d postgres
   ```

5. **Run migrations:**

   Generate Ent code from schema (if needed):

   ```bash
   make schema_generate
   ```

   Apply Atlas migrations:

   ```bash
   make migrate_apply
   ```

   Optional: generate a new migration after schema changes:

   ```bash
   make migrate_generate MIGRATION_NAME=your_migration_name
   ```

6. **Run the server:**

   ```bash
   make server
   ```

7. **Access the API:**

   - HTTP API: [http://localhost:3000](http://localhost:3000)
   - Base path: `/api/v1` (e.g. `http://localhost:3000/api/v1/auth/sign-in`)
   - WebSocket (SignalR): `http://localhost:3000/websocket/v1`

## Usage

- **Development**: Use `make server` for local runs.
- **Production**: Build the binary and run it, or use the provided Dockerfile and `compose.yaml`.

### Makefile commands

| Command                                     | Description                                          |
| ------------------------------------------- | ---------------------------------------------------- |
| `make server`                               | Run the app (`go run cmd/server/server.go`)          |
| `make schema_generate`                      | Generate Ent code from `database/ent/schema`         |
| `make migrate_generate MIGRATION_NAME=name` | Generate a new Atlas migration from schema diff      |
| `make migrate_apply`                        | Apply pending migrations (uses `DB_URL` in Makefile) |
| `make migrate_status`                       | Show migration status                                |
| `make schema_inspect`                       | Inspect schema (Atlas)                               |

Override `DB_URL` in the Makefile or environment when your PostgreSQL connection differs.

## Project Structure

```
chat_be/
├── cmd/
│   └── server/
│       └── server.go          # Application entrypoint (FX modules)
├── apperror/                  # App errors and global handler
├── common/                    # Shared utilities (e.g. OTP, result type)
├── config/                    # Env loading (Viper) and constants
├── conversation/              # Conversation and message logic + router
├── database/
│   ├── ent/                   # Ent client, schema, codegen
│   │   ├── schema/            # User, Conversation, Message, etc.
│   │   └── migrate/           # Atlas migrations
│   ├── client.go
│   └── predicate/
├── file/                      # File upload and serving
├── http/
│   ├── handler/               # Error handler, request tracking
│   ├── pagination/            # Pagination helpers
│   ├── validation/            # Request validation
│   ├── http.go                # HTTP server wiring + SignalR mount
│   └── router.go              # API routes (/api/v1)
├── logger/                    # Zap logger setup
├── notification/
│   └── mail/                  # SMTP client and templates
├── security/
│   ├── auth/                  # JWT verification, RequireUser middleware
│   └── jwt/                   # JWT issue and claims
├── user/
│   ├── auth/                  # Sign-in, verify OTP, issue token
│   └── profile/               # Get/update profile
├── websocket/                 # SignalR hub (presence, messages)
├── .env                       # Local env (create from example above)
├── compose.yaml               # Docker Compose (Postgres + app)
├── Dockerfile
├── go.mod
├── Makefile
└── README.md
```

## Configuration

### Environment variables

| Variable                                               | Description                                    |
| ------------------------------------------------------ | ---------------------------------------------- |
| `PORT`                                                 | HTTP server port (default: 3000)               |
| `DB_URL`                                               | PostgreSQL connection string                   |
| `JWT_ACCESS_TOKEN_SECRET_KEY`                          | Secret for signing JWT access tokens           |
| `JWT_ACCESS_TOKEN_EXPIRES_IN`                          | Access token lifetime (e.g. `168h` for 7 days) |
| `MAIL_HOST`, `MAIL_PORT`, `MAIL_USER`, `MAIL_PASSWORD` | SMTP settings for verification emails          |

### API conventions

- Base path: `/api/v1`
- Protected routes require `Authorization: Bearer <accessToken>`
- Global error handler and CORS (e.g. allow all in dev)
- WebSocket endpoint: `/websocket/v1` (SignalR)

## License

This project is licensed under the MIT License.
