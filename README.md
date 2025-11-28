# 🎫 High Concurrency Ticket System

A high-traffic ticket sales system that solves the Race Condition problem. Built with Go, Gin, GORM, and Supabase (PostgreSQL).

## 🚀 Features

- **Race Condition Protection**: Data consistency in concurrent requests using Worker pattern and transactions
- **Buffered Channel Queue**: Queue system with 1000 request capacity for backpressure management
- **Graceful Shutdown**: Completion of active requests when server is shutting down
- **PostgreSQL + GORM**: Reliable database operations and ORM support
- **RESTful API**: Fast HTTP endpoints with Gin framework

## 🏗️ Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Client    │────▶│   Gin API    │────▶│   Channel   │
│  (100 req)  │     │   /buy       │     │  (Buffer)   │
└─────────────┘     └──────────────┘     └──────┬──────┘
                                                │
                                                ▼
                    ┌──────────────┐     ┌─────────────┐
                    │  PostgreSQL  │◀────│   Worker    │
                    │  (Supabase)  │     │ (Sequential)│
                    └──────────────┘     └─────────────┘
```

## 📁 Project Structure

```
ticket-system/
├── cmd/
│   └── api/
│       └── main.go          # Main application entry point
├── internal/
│   ├── models/
│   │   └── event.go         # Event and Booking models
│   └── worker/
│       └── processor.go     # Queue processor worker
├── attack.go                # Load testing script
├── .env                     # Environment variables (not in git)
├── go.mod
└── go.sum
```

## 🛠️ Installation

### Requirements

- Go 1.21+
- PostgreSQL (or Supabase account)

### Steps

1. **Clone the repository:**
```bash
git clone https://github.com/altugikiz/Ticket-Tracking-RaceCondition.git
cd Ticket-Tracking-RaceCondition/ticket-system
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Create environment file:**
```bash
cp .env.example .env
# Edit .env file and set DATABASE_URL
```

4. **Database tables (Auto Migration):**
```sql
-- Run in Supabase SQL Editor
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO events (id, name, total_quota, available_quota, version)
VALUES (uuid_generate_v4(), 'Concert Event', 100, 100, 1);
```

5. **Start the server:**
```bash
cd cmd/api
go run main.go
```

## 📡 API Usage

### Buy Ticket

```bash
POST /buy
Content-Type: application/json

{
  "event_id": "155ff34d-51ec-4053-841e-a6cc24253256",
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
}
```

**Success Response:**
```json
{
  "message": "Request queued",
  "status": "pending"
}
```

**System Busy Response:**
```json
{
  "error": "System is too busy"
}
```

## 🧪 Load Testing

To send 100 concurrent requests:

```bash
cd ticket-system
go run attack.go
```

**Expected Result:**
- 100 requests sent to an event with 100 quota
- Processed sequentially without race condition
- Quota drops to 0, no overselling

## 🔧 Race Condition Solution

### The Problem
```
User A: Read quota → 100
User B: Read quota → 100
User A: Update quota → 99
User B: Update quota → 99  ❌ (Dropped from 99 instead of 100)
```

### Solution: Worker Pattern + Buffered Channel

```go
// All requests enter a single channel
var TicketQueue = make(chan TicketRequest, 1000)

// Single worker processes sequentially
func StartWorker(db *gorm.DB) {
    go func() {
        for req := range TicketQueue {
            processTicket(db, req)  // Sequential processing
        }
    }()
}
```

## 🗄️ Database Schema

### Events Table
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary Key |
| name | VARCHAR | Event name |
| total_quota | INT | Total capacity |
| available_quota | INT | Remaining capacity |
| version | INT | For optimistic locking |
| created_at | TIMESTAMP | Creation date |

### Bookings Table
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary Key |
| event_id | UUID | Foreign Key → Events |
| user_id | UUID | User ID |
| status | VARCHAR | SUCCESS / FAILED |
| created_at | TIMESTAMP | Transaction date |

## 🔒 Security

- `.env` file in `.gitignore`
- Passwords are URL encoded
- Atomic operations with transactions

## 📈 Performance

- **Throughput**: ~20 requests/second (with 50ms artificial delay)
- **Queue Capacity**: 1000 requests
- **Backpressure**: Returns 503 when queue is full

## 🛡️ Tech Stack

| Technology | Usage |
|------------|-------|
| **Go** | Backend language |
| **Gin** | HTTP framework |
| **GORM** | ORM |
| **PostgreSQL** | Database |
| **Supabase** | DBaaS |

## 📝 License

MIT License

## 👤 Developer

**Altuğ İkiz**

- GitHub: [@altugikiz](https://github.com/altugikiz)

---

⭐ If you liked this project, don't forget to give it a star!