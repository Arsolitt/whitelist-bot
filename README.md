# Whitelist Bot

Telegram bot for managing whitelist requests with admin approval workflow. Users submit requests with their nicknames, and administrators can approve or decline them through inline keyboard controls.

## Features

- **User requests**: Submit whitelist requests with custom nickname
- **Admin panel**: View pending requests with inline approve/decline buttons
- **State machine**: FSM-based conversation flow for handling multi-step interactions
- **Audit trail**: Track who approved/declined requests with timestamps
- **Locking mechanism**: Prevent concurrent request processing
- **Structured logging**: Context-aware logging with request tracking

## Tech Stack

- **Go 1.25.3**
- **SQLite** - Persistent data storage
- **[go-telegram/bot](https://github.com/go-telegram/bot)** - Telegram Bot API client
- **[sqlc](https://sqlc.dev)** - Type-safe SQL code generation
- **[goose](https://github.com/pressly/goose)** - Database migrations
- **[cleanenv](https://github.com/ilyakaznacheev/cleanenv)** - Environment configuration
- **[validator](https://github.com/go-playground/validator)** - Config validation

## Architecture

The project follows **Clean Architecture** principles with clear separation of concerns:

```
internal/
├── core/           # Configuration, commands, shared utilities
├── domain/         # Business entities (User, WLRequest) with builders
├── repository/     # Data access layer with SQLite implementation
├── handlers/       # Telegram message/callback handlers
├── router/         # Custom routing with matcher patterns
├── fsm/            # Finite State Machine for conversation flows
└── locker/         # Concurrency control
```

### Key Design Patterns

- **Repository Pattern**: Abstract data access from business logic
- **Builder Pattern**: Immutable domain entities with controlled mutations
- **FSM Pattern**: State-driven conversation management
- **Middleware Pattern**: Logging and error handling wrappers
- **Matcher Pattern**: Flexible routing based on multiple conditions

### Data Flow

1. User sends command → Router matches handler
2. Handler validates state (FSM) and permissions
3. Domain logic processes request with builder pattern
4. Repository persists changes to SQLite
5. Response sent back through Telegram API

## Project Structure

```
.
├── cmd/bot/              # Application entry point
├── internal/             # Private application code
│   ├── core/            # Core utilities, config, logging
│   ├── domain/          # Business entities
│   ├── handlers/        # Telegram handlers
│   ├── repository/      # Data layer
│   ├── router/          # Custom routing logic
│   ├── fsm/            # State machine
│   └── locker/         # Concurrency control
├── migrations/          # Database migrations (goose)
├── queries/            # SQL queries (sqlc source)
└── data/              # SQLite database file
```

## Setup

### Prerequisites

- Go 1.25.3 or higher
- SQLite3
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- Your Telegram ID (from [@userinfobot](https://t.me/userinfobot))

### Installation

1. **Clone the repository**

```bash
git clone <repository-url>
cd whitelist
```

2. **Create environment configuration**

```bash
cp env.example .env
```

Edit `.env` and fill in required values:

```env
# Telegram Configuration
TELEGRAM_TOKEN=your_bot_token_here
TELEGRAM_ADMIN_IDS=123456789,987654321  # Comma-separated admin IDs
TELEGRAM_DEBUG=false

# Database Configuration
DATABASE_PATH=data/whitelist.db
DATABASE_MAX_OPEN_CONNS=10
DATABASE_MAX_IDLE_CONNS=5

# Logging Configuration
LOGS_LEVEL=info  # debug, info, warn, error
LOGS_IS_PRETTY=true
LOGS_WITH_CONTEXT=true
LOGS_WITH_SOURCES=false

# Server Configuration
SERVER_MAX_REQUESTS_PER_USER=3
```

3. **Install dependencies**

```bash
go mod download
```

4. **Run database migrations**

```bash
# Install goose if not already installed
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir migrations/sqlite sqlite3 data/whitelist.db up
```

5. **Run the bot**

```bash
go run cmd/bot/main.go
```

## Usage

### User Commands

- `/start` - Register and get welcome message
- `/info` - Display bot information and available commands
- `/new_request` - Submit a new whitelist request
  1. Bot asks for nickname
  2. User enters nickname
  3. Request submitted for admin review

### Admin Commands

- `/view_pending` - View all pending whitelist requests
  - Shows up to 5 requests at a time
  - Each with ✅ Approve / ❌ Decline buttons
  - Displays requester info and timestamp

## Development

### Generate SQL code (sqlc)

After modifying `queries/*.sql`:

```bash
sqlc generate
```

### Create new migration

```bash
goose -dir migrations/sqlite create <migration_name> sql
```

## TODO

- [ ] JSON-based callback queries instead of string parsing
- [ ] FSM metadata storage as JSON
- [ ] Scheduled notifications for pending requests
- [ ] User notifications on request approval/decline
- [ ] Nickname validation (length, special characters)
- [ ] Permission middleware
- [ ] Panic recovery middleware
- [ ] Rate limiting per user

## License

This project is private and not licensed for public use.

