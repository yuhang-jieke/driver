# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

花小猪打车司机端 (Huaxiaozhu Driver App) - A ride-hailing driver application with a Go microservice backend and React frontend.

**Current date: 2026/05/03**

## Build Commands

### Go Backend

```bash
# Generate protobuf code (must run after modifying .proto files)
./taketaxi/scripts/gen_proto.sh

# Build srvDriver (gRPC service)
./taketaxi/scripts/build.sh
# Output: bin/srvDriver

# Run srvDriver
go run ./taketaxi/srvDriver/cmd/main.go -config=taketaxi/srvDriver/configs/config.yaml

# Run bffDriver (BFF layer - HTTP API)
go run ./taketaxi/bffDriver/cmd/main.go -config=taketaxi/bffDriver/configs/config.yaml

# Run tests
go test ./taketaxi/pkg/logger/...
go test ./taketaxi/pkg/upload/...
```

### React Frontend

```bash
cd driverfrontend

# Install dependencies
npm install

# Development server (runs on http://localhost:5173)
npm run dev

# Production build
npm run build

# Preview production build
npm run preview
```

## Architecture

### Backend: Three-Layer Microservice

```
┌─────────────────┐     HTTP      ┌─────────────────┐
│   React App     │ ───────────── │    bffDriver    │
│  (driverfrontend)│   :5173      │   HTTP :8080    │
└─────────────────┘               └────────┬────────┘
                                           │ gRPC
                                           ▼
                                  ┌─────────────────┐
                                  │    srvDriver    │
                                  │   gRPC :8001    │
                                  └────────┬────────┘
                                           │
                         ┌─────────────────┼─────────────────┐
                         ▼                 ▼                 ▼
                    ┌─────────┐      ┌─────────┐      ┌─────────┐
                    │  MySQL  │      │  Redis  │      │  MinIO  │
                    └─────────┘      └─────────┘      └─────────┘
```

**bffDriver** (BFF Layer):
- Entry: `taketaxi/bffDriver/cmd/main.go`
- HTTP server using Gin
- Routes defined in `taketaxi/bffDriver/internal/router/router.go`
- RPC client in `taketaxi/bffDriver/internal/rpcClient/`
- Handlers in `taketaxi/bffDriver/internal/handler/`

**srvDriver** (gRPC Service):
- Entry: `taketaxi/srvDriver/cmd/main.go`
- gRPC server implementing `DriverService`
- Layered architecture: `handler → service → repository`
- Models in `taketaxi/srvDriver/internal/model/`
- Proto definitions in `taketaxi/common/idl/driver.proto`

### Frontend: React SPA

```
driverfrontend/src/app/
├── api/           # API client modules (profile.ts, wallet.ts, etc.)
├── components/    # React components (DriverApp.tsx, AmapView.tsx)
├── hooks/         # Custom hooks (useGeolocation.ts)
├── store.tsx      # Global state management
└── utils/         # Utilities (amap.ts for AMap integration)
```

**Key Technologies**:
- React 18 + TypeScript
- Tailwind CSS 4
- Vite 6
- 高德地图 JS API 2.0 (AMap for maps/location)

## API Conventions

### Backend Proto → HTTP Mapping

The gRPC service is exposed via HTTP through bffDriver. Proto methods map to REST endpoints:

| gRPC Method | HTTP Endpoint | Method |
|-------------|---------------|--------|
| GetProfile | `/api/v1/driver/profile` | GET |
| UpdateProfile | `/api/v1/driver/profile` | PUT |
| GetWallet | `/api/v1/driver/wallet` | GET |
| BindBankCard | `/api/v1/driver/bankcard` | POST |
| ... | ... | ... |

### Frontend API Pattern

API modules in `driverfrontend/src/app/api/` follow this pattern:
- Export async functions that return `Promise<T | null>`
- Use `API_BASE = "http://localhost:8080"` for local development
- Handle errors with console.error and return null

## Configuration

Both services use YAML config files:
- `taketaxi/bffDriver/configs/config.yaml`
- `taketaxi/srvDriver/configs/config.yaml`

Key config sections:
- `server`: host, port, grpc_host, grpc_port
- `database`: MySQL connection
- `redis`: Redis connection
- `upload`: Storage config (MinIO/OSS/COS)
- `log`: Zap logger config

## Coordinate System

**Important**: The project uses `[lat, lng]` format internally, but AMap APIs require `[lng, lat]`. Convert when calling AMap:

```typescript
// Project format: [lat, lng]
const coords = [33.95, 118.3];

// AMap API call: [lng, lat]
map.setCenter([coords[1], coords[0]]);
```

## File Upload

Upload supports multiple storage backends (MinIO, Aliyun OSS, Tencent COS). Storage type is configured via `upload.storage_type` in config.

Upload endpoint: `POST /api/v1/upload/image`

## Adding New gRPC Methods

1. Add method definition to `taketaxi/common/idl/driver.proto`
2. Run `./taketaxi/scripts/gen_proto.sh` to generate Go code
3. Implement handler in `taketaxi/srvDriver/internal/handler/`
4. Add repository method if DB access needed
5. Add HTTP route in `taketaxi/bffDriver/internal/router/router.go`
6. Add RPC client call in `taketaxi/bffDriver/internal/rpcClient/driverClient.go`
7. Add frontend API function in `driverfrontend/src/app/api/`

## Error Code System

Error codes are defined in `taketaxi/pkg/errcode/errcode.go`:

| Range | Category | Example |
|-------|----------|---------|
| 0 | Success | `Success = 0` |
| 4xxxx | Parameter/Request errors | `ErrInvalidDriverID = 40002` |
| 5xxxx | Business rule errors | `ErrInsufficientBalance = 50009` |
| 6xxxx | Dependency service errors | `ErrRedisError = 60001` |
| 9xxxx | Internal system errors | `ErrInternal = 90001` |

Use `errcode.New(code)` for standard errors, `errcode.NewWithDetail(code, detail)` for additional context.

## Business Domains

The driver service is organized by domain:

| Domain | Service File | Repository File | Description |
|--------|-------------|-----------------|-------------|
| profile | `profileService.go` | `profileRepo.go` | Driver info, avatar, contact |
| verify | `verifyService.go` | `verifyRepo.go` | Realname, license, vehicle verification |
| bankcard | `bankCardService.go` | `bankCardRepo.go` | Bank card binding/management |
| wallet | `walletService.go` | `walletRepo.go` | Balance, withdraw, income |
| auth | `authService.go` | `authRepo.go` | Authentication, password |

## Key Database Tables

| Table | Model | Purpose |
|-------|-------|---------|
| `drivers` | `DriverS` | Driver account info |
| `driver_wallet` | `DriverWallet` | Balance, frozen amount, totals |
| `driver_bank_card` | `DriverBankCard` | Bound bank cards |
| `driver_realname` | `DriverRealname` | Real-name verification |
| `driver_license` | `DriverLicense` | Driver's license info |
| `driver_vehicle` | `DriverVehicle` | Vehicle registration |
| `withdraw_record` | `DriverWithdrawRecord` | Withdrawal records |
| `driver_income_log` | `DriverIncomeLog` | Income transaction log |
| `wallet_transaction_log` | `WalletTransactionLog` | Wallet balance changes |

## Key Business Rules

### Withdraw (提现)
- Min amount: 100分 (1元), Max per order: 500000分 (5000元)
- Max 3 withdrawals per day
- Requires: bound bank card (debit only), real-name verification passed
- Bank card holder name must match real-name

### Bank Card (银行卡)
- Card number max length: 20 digits
- Account holder name: 2-6 Chinese characters (`^[一-龥]{2,6}$`)
- Only debit cards allowed for withdrawal
- One card per driver

### Wallet Balance
- `Balance` = total balance (includes frozen)
- `FrozenAmount` = frozen subset (T-3 settlement)
- Available = Balance - FrozenAmount

## Architecture Rules

Follow the layered architecture strictly:

```
handler → service → repository → model
```

**Rules:**
- Handler: HTTP/gRPC entry, param validation, protocol conversion
- Service: business logic, state transitions, transactions
- Repository: database access only, no business rules
- Never skip layers or reverse dependencies
- Each layer uses distinct data structures (no reuse across request/entity/response)
