# Driver Withdraw Product Rules Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a product-grade driver withdraw flow that adds a dedicated withdraw-page rules API, upgrades withdraw validation, and supports a “2 hours arrival” experience without integrating a real payout channel.

**Architecture:** Extend `driver.proto` with a dedicated `GetWithdrawPage` RPC, then implement the rule assembly in `srvDriver` and expose it through `bffDriver`. Reuse one eligibility path for both page rendering and withdraw submission so the page state and submit-time validation stay consistent.

**Tech Stack:** Go, gRPC/protobuf, Gin, GORM, existing `errcode` package, existing wallet/bankcard/verify models.

---

### Task 1: Extend Withdraw RPC Contract

**Files:**
- Modify: `D:/software/GoWork/src/driver/taketaxi/common/idl/driver.proto`
- Modify: `D:/software/GoWork/src/driver/taketaxi/common/kitexGen/driver.pb.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/common/kitexGen/driver_grpc.pb.go`

- [ ] **Step 1: Add the new withdraw-page messages and RPC definition**

```proto
service DriverService {
  rpc GetWithdrawPage(GetWithdrawPageReq) returns (GetWithdrawPageResp);
}

message GetWithdrawPageReq {
  int64 driver_id = 1;
}

message WithdrawRuleInfo {
  double min_withdraw_amount = 1;
  double max_withdraw_amount = 2;
  int32 today_withdraw_count = 3;
  int32 today_withdraw_limit = 4;
  string estimated_arrival_text = 5;
  bool fee_free = 6;
  double fee_amount = 7;
  string fee_desc = 8;
}

message WithdrawPageBankCard {
  bool has_bank_card = 1;
  string bank_name = 2;
  string bank_card_no = 3;
}

message WithdrawPageActionState {
  bool can_withdraw = 1;
  string disable_reason_code = 2;
  string disable_reason_text = 3;
}

message GetWithdrawPageResp {
  double wallet_balance = 1;
  double frozen_amount = 2;
  double available_withdraw_amount = 3;
  WithdrawRuleInfo rule_info = 4;
  WithdrawPageBankCard bank_card = 5;
  WithdrawPageActionState action_state = 6;
  repeated double suggested_amounts = 7;
  repeated string withdraw_notice = 8;
}
```

- [ ] **Step 2: Generate protobuf and gRPC bindings**

Run: regenerate the protobuf outputs for `driver.proto`
Expected: `driver.pb.go` and `driver_grpc.pb.go` include `GetWithdrawPage*` types and the new `DriverService_GetWithdrawPage_FullMethodName`

- [ ] **Step 3: Verify the generated client/server surfaces exist**

Run: inspect generated files for `GetWithdrawPage`
Expected: both the client interface and server interface contain `GetWithdrawPage`

- [ ] **Step 4: Commit the protocol-only change**

```bash
git add taketaxi/common/idl/driver.proto taketaxi/common/kitexGen/driver.pb.go taketaxi/common/kitexGen/driver_grpc.pb.go
git commit -m "feat: add withdraw page rpc contract"
```

### Task 2: Add Withdraw Rule Modeling and Error Codes

**Files:**
- Modify: `D:/software/GoWork/src/driver/taketaxi/pkg/errcode/errcode.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/model/enum_wallet.go`
- Create: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/service/walletService_test.go`

- [ ] **Step 1: Write the failing service-level tests for withdraw eligibility**

```go
func TestEvaluateWithdrawEligibility(t *testing.T) {
    tests := []struct {
        name       string
        wallet     *model.DriverWallet
        hasCard    bool
        verifyOK   bool
        todayCount int64
        wantCode   string
        wantCan    bool
    }{
        {name: "no bank card", wallet: &model.DriverWallet{Balance: 120}, hasCard: false, verifyOK: true, todayCount: 0, wantCode: "NO_BANK_CARD", wantCan: false},
        {name: "verify pending", wallet: &model.DriverWallet{Balance: 120}, hasCard: true, verifyOK: false, todayCount: 0, wantCode: "VERIFY_PENDING", wantCan: false},
        {name: "today limit reached", wallet: &model.DriverWallet{Balance: 120}, hasCard: true, verifyOK: true, todayCount: 3, wantCode: "WITHDRAW_COUNT_LIMIT", wantCan: false},
        {name: "available amount zero", wallet: &model.DriverWallet{Balance: 0}, hasCard: true, verifyOK: true, todayCount: 0, wantCode: "AVAILABLE_AMOUNT_ZERO", wantCan: false},
        {name: "can withdraw", wallet: &model.DriverWallet{Balance: 120}, hasCard: true, verifyOK: true, todayCount: 0, wantCode: "", wantCan: true},
    }
    _ = tests
}
```

- [ ] **Step 2: Run the focused test and confirm it fails before implementation**

Run: `go test ./taketaxi/srvDriver/internal/service -run TestEvaluateWithdrawEligibility -v`
Expected: FAIL because the new eligibility helper and related rule constants do not exist yet

- [ ] **Step 3: Add the new error codes and withdraw-page rule constants**

```go
const (
    ErrWithdrawMinAmount       Code = 50020
    ErrWithdrawRealnameNeeded  Code = 50021
    ErrWithdrawPageUnavailable Code = 50022
)

const (
    WithdrawMinAmount       = 1
    WithdrawArrivalText     = "预计2小时到账"
    WithdrawDisableNoCard   = "NO_BANK_CARD"
    WithdrawDisableVerify   = "VERIFY_PENDING"
    WithdrawDisableCount    = "WITHDRAW_COUNT_LIMIT"
    WithdrawDisableZero     = "AVAILABLE_AMOUNT_ZERO"
)
```

- [ ] **Step 4: Add the minimal rule-evaluation helpers to make the test pass**

```go
type withdrawEligibility struct {
    CanWithdraw       bool
    DisableReasonCode string
    DisableReasonText string
}
```

- [ ] **Step 5: Run the focused test again and confirm it passes**

Run: `go test ./taketaxi/srvDriver/internal/service -run TestEvaluateWithdrawEligibility -v`
Expected: PASS

- [ ] **Step 6: Commit the rule-modeling change**

```bash
git add taketaxi/pkg/errcode/errcode.go taketaxi/srvDriver/internal/model/enum_wallet.go taketaxi/srvDriver/internal/service/walletService_test.go taketaxi/srvDriver/internal/service/walletService.go
git commit -m "feat: add withdraw rule modeling"
```

### Task 3: Implement srvDriver Withdraw Page and Transactional Apply

**Files:**
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/handler/walletHandler.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/service/walletService.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/repository/walletRepo.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/repository/withdrawRepo.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/repository/verifyRepo.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/repository/bankCardRepo.go`
- Test: `D:/software/GoWork/src/driver/taketaxi/srvDriver/internal/service/walletService_test.go`

- [ ] **Step 1: Write the failing tests for `GetWithdrawPage` and transactional apply behavior**

```go
func TestGetWithdrawPageReturnsRulePayload(t *testing.T) {
    // assert estimated arrival text, limits, bank card summary, and action state
}

func TestApplyWithdrawRejectsAmountBelowMinimum(t *testing.T) {
    // amount=0.5 should return ErrWithdrawMinAmount
}
```

- [ ] **Step 2: Run the focused wallet service tests and confirm they fail**

Run: `go test ./taketaxi/srvDriver/internal/service -run 'Test(GetWithdrawPageReturnsRulePayload|ApplyWithdrawRejectsAmountBelowMinimum)' -v`
Expected: FAIL because `GetWithdrawPage` and the new validation branch are not implemented

- [ ] **Step 3: Implement repository helpers required by the new rule flow**

```go
func (r *DriverRepo) GetDriverRealname(ctx context.Context, driverID int64) (*model.DriverRealname, error)
func (r *DriverRepo) RunInTx(ctx context.Context, fn func(txRepo *DriverRepo) error) error
```

- [ ] **Step 4: Implement the withdraw-page service assembly**

```go
func (s *DriverService) GetWithdrawPage(ctx context.Context, req *driver.GetWithdrawPageReq) (*driver.GetWithdrawPageResp, error) {
    // load wallet, bank card, verify state, today count
    // compute action state once
    // return page-ready payload
}
```

- [ ] **Step 5: Refactor `ApplyWithdraw` to reuse the same validation path and update data in one transaction**

```go
return s.repo.RunInTx(ctx, func(txRepo *repository.DriverRepo) error {
    // create withdraw record
    // update wallet by optimistic lock
    // create wallet transaction log
    return nil
})
```

- [ ] **Step 6: Add the new gRPC handler method**

```go
func (h *DriverHandler) GetWithdrawPage(ctx context.Context, req *driver.GetWithdrawPageReq) (*driver.GetWithdrawPageResp, error) {
    return h.svc.GetWithdrawPage(ctx, req)
}
```

- [ ] **Step 7: Run the wallet service tests and the srvDriver package tests**

Run: `go test ./taketaxi/srvDriver/internal/service ./taketaxi/srvDriver/internal/repository -v`
Expected: PASS

- [ ] **Step 8: Commit the srvDriver implementation**

```bash
git add taketaxi/srvDriver/internal/handler/walletHandler.go taketaxi/srvDriver/internal/service/walletService.go taketaxi/srvDriver/internal/service/walletService_test.go taketaxi/srvDriver/internal/repository/walletRepo.go taketaxi/srvDriver/internal/repository/withdrawRepo.go taketaxi/srvDriver/internal/repository/verifyRepo.go taketaxi/srvDriver/internal/repository/bankCardRepo.go
git commit -m "feat: implement withdraw page rules in srv driver"
```

### Task 4: Expose Withdraw Page Through BFF

**Files:**
- Modify: `D:/software/GoWork/src/driver/taketaxi/bffDriver/internal/rpcClient/driverClient.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/bffDriver/internal/handler/driverHandler.go`
- Modify: `D:/software/GoWork/src/driver/taketaxi/bffDriver/internal/router/router.go`

- [ ] **Step 1: Add the new RPC client wrapper**

```go
func (c *DriverClient) GetWithdrawPage(ctx context.Context, req *driver.GetWithdrawPageReq) (*driver.GetWithdrawPageResp, error) {
    return c.client.GetWithdrawPage(ctx, req)
}
```

- [ ] **Step 2: Add a BFF HTTP handler for the withdraw page**

```go
func (h *DriverHandler) GetWithdrawPage(c *gin.Context) {
    driverID, _ := strconv.ParseInt(c.Query("driver_id"), 10, 64)
    resp, err := h.client.GetWithdrawPage(c.Request.Context(), &pb.GetWithdrawPageReq{DriverId: driverID})
    _ = resp
    _ = err
}
```

- [ ] **Step 3: Register the route next to the existing wallet endpoints**

```go
wallet.GET("/withdraw/page", driverHandler.GetWithdrawPage)
```

- [ ] **Step 4: Run a build for BFF and srvDriver**

Run: `go test ./taketaxi/bffDriver/... ./taketaxi/srvDriver/...`
Expected: PASS

- [ ] **Step 5: Commit the BFF integration**

```bash
git add taketaxi/bffDriver/internal/rpcClient/driverClient.go taketaxi/bffDriver/internal/handler/driverHandler.go taketaxi/bffDriver/internal/router/router.go
git commit -m "feat: expose withdraw page api in bff"
```

### Task 5: Full Regression and Documentation Touch-Up

**Files:**
- Modify: `D:/software/GoWork/src/driver/docs/driver-profile-module-spec.md`
- Modify: `D:/software/GoWork/src/driver/docs/superpowers/specs/2026-04-29-driver-withdraw-product-rules-design.md`

- [ ] **Step 1: Update the existing docs to reference the new withdraw-page capability**

```markdown
- 新增提现页查询接口 `GetWithdrawPage`
- 提现页规则统一由 srvDriver 输出
```

- [ ] **Step 2: Run the targeted test suite and a full project test sweep**

Run: `go test ./...`
Expected: PASS, or a clear list of unrelated pre-existing failures if the repository already contains them

- [ ] **Step 3: Perform a manual grep for the new API names and disable-reason constants**

Run: search for `GetWithdrawPage` and `WITHDRAW_COUNT_LIMIT`
Expected: proto, client, handler, service, and tests all reference the same names

- [ ] **Step 4: Commit the final cleanup**

```bash
git add docs/driver-profile-module-spec.md docs/superpowers/specs/2026-04-29-driver-withdraw-product-rules-design.md
git commit -m "docs: sync withdraw product rules docs"
```
