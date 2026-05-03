# 银行卡持卡人姓名验证实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标：** 在银行卡绑定和更换流程中添加持卡人姓名格式验证（2-6个汉字），前后端双重验证

**架构：** 后端在 `bankCardService.go` 的 `BindBankCard` 和 `UpdateBankCard` 方法中添加正则验证，前端在 `wallet.ts` API 调用前进行客户端验证。验证规则：`^[一-龥]{2,6}$`

**技术栈：** Go (regexp), TypeScript (正则表达式), 错误码系统

---

## 文件结构

**后端修改：**
- `taketaxi/pkg/errcode/errcode.go` - 添加新错误码 `ErrInvalidAccountName`
- `taketaxi/srvDriver/internal/service/bankCardService.go` - 添加验证逻辑和辅助函数

**前端修改：**
- `driverfrontend/src/app/api/wallet.ts` - 在 `bindBankCard` 和 `updateBankCard` 函数中添加客户端验证

**测试：**
- 手动测试（前后端集成测试）

---

## Task 1: 添加后端错误码

**文件：**
- Modify: `taketaxi/pkg/errcode/errcode.go:33-34` (在 ErrBankCardNoTooLong 后添加)
- Modify: `taketaxi/pkg/errcode/errcode.go:81-82` (在 codeMessages 映射中添加)

- [ ] **步骤 1: 添加错误码常量**

在 `errcode.go` 第 33 行后添加：

```go
ErrBankCardNoTooLong Code = 40006 // 银行卡号超长
ErrInvalidAccountName Code = 40007 // 持卡人姓名格式不正确
```

- [ ] **步骤 2: 添加错误码描述**

在 `errcode.go` 第 81 行后添加：

```go
ErrBankCardNoTooLong:   "银行卡号超长",
ErrInvalidAccountName:  "持卡人姓名格式不正确，必须为2-6个汉字",
```

- [ ] **步骤 3: 验证编译**

运行：
```bash
cd taketaxi/srvDriver
go build ./...
```

预期：编译成功，无错误

- [ ] **步骤 4: 提交更改**

```bash
git add taketaxi/pkg/errcode/errcode.go
git commit -m "feat(errcode): add ErrInvalidAccountName for bank card name validation"
```

---

## Task 2: 添加后端验证逻辑

**文件：**
- Modify: `taketaxi/srvDriver/internal/service/bankCardService.go:1` (添加 regexp 导入)
- Modify: `taketaxi/srvDriver/internal/service/bankCardService.go:28-29` (在 BindBankCard 中添加验证)
- Modify: `taketaxi/srvDriver/internal/service/bankCardService.go:110-111` (在 UpdateBankCard 中添加验证)
- Modify: `taketaxi/srvDriver/internal/service/bankCardService.go:162` (文件末尾添加辅助函数)

- [ ] **步骤 1: 添加 regexp 包导入**

在 `bankCardService.go` 文件顶部的 import 块中添加：

```go
import (
	"context"
	"fmt"
	"regexp"
	"time"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/pkg/errcode"
)
```

- [ ] **步骤 2: 在 BindBankCard 方法中添加验证**

在 `bankCardService.go` 第 28 行（卡号长度检查之后）插入：

```go
if len(req.BankCardNo) > model.BankCardNoMaxLength {
	return nil, errcode.New(errcode.ErrBankCardNoTooLong)
}

// 验证持卡人姓名格式（2-6个汉字）
if !isValidChineseName(req.AccountName) {
	return nil, errcode.New(errcode.ErrInvalidAccountName)
}

// 拒绝信用卡（信用卡不能接收转账，打款必失败）
```

- [ ] **步骤 3: 在 UpdateBankCard 方法中添加验证**

在 `bankCardService.go` 第 110 行（卡号长度检查之后）插入：

```go
if len(req.BankCardNo) > model.BankCardNoMaxLength {
	return nil, errcode.New(errcode.ErrBankCardNoTooLong)
}

// 验证持卡人姓名格式（2-6个汉字）
if !isValidChineseName(req.AccountName) {
	return nil, errcode.New(errcode.ErrInvalidAccountName)
}

// 拒绝信用卡
```

- [ ] **步骤 4: 添加验证辅助函数**

在 `bankCardService.go` 文件末尾（第 162 行后）添加：

```go
// isValidChineseName 验证姓名是否为2-6个汉字
func isValidChineseName(name string) bool {
	matched, _ := regexp.MatchString(`^[\x{4e00}-\x{9fa5}]{2,6}$`, name)
	return matched
}
```

- [ ] **步骤 5: 验证编译**

运行：
```bash
cd taketaxi/srvDriver
go build ./...
```

预期：编译成功，无错误

- [ ] **步骤 6: 提交更改**

```bash
git add taketaxi/srvDriver/internal/service/bankCardService.go
git commit -m "feat(service): add account name validation in bank card binding"
```

---

## Task 3: 添加前端验证逻辑

**文件：**
- Modify: `driverfrontend/src/app/api/wallet.ts:31-52` (bindBankCard 函数)
- Modify: `driverfrontend/src/app/api/wallet.ts:54-75` (updateBankCard 函数)

- [ ] **步骤 1: 添加验证辅助函数**

在 `wallet.ts` 文件顶部（第 1 行后）添加：

```typescript
const API_BASE = "http://localhost:8080";

// 验证持卡人姓名格式（2-6个汉字）
function validateAccountName(name: string): boolean {
  const regex = /^[一-龥]{2,6}$/;
  return regex.test(name);
}
```

- [ ] **步骤 2: 在 bindBankCard 函数中添加验证**

修改 `bindBankCard` 函数（第 31-52 行）：

```typescript
export async function bindBankCard(params: {
  driver_id: number;
  bank_name: string;
  bank_code?: string;
  bank_card_no: string;
  account_name: string;
  card_type?: number;
  branch_name?: string;
}): Promise<{ success: boolean; message: string } | null> {
  // 客户端验证：持卡人姓名格式
  if (!validateAccountName(params.account_name)) {
    return {
      success: false,
      message: "持卡人姓名必须为2-6个汉字"
    };
  }

  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/bankcard`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("BindBankCard API failed:", err);
    return null;
  }
}
```

- [ ] **步骤 3: 在 updateBankCard 函数中添加验证**

修改 `updateBankCard` 函数（第 54-75 行）：

```typescript
export async function updateBankCard(params: {
  driver_id: number;
  bank_name: string;
  bank_code?: string;
  bank_card_no: string;
  account_name: string;
  card_type?: number;
  branch_name?: string;
}): Promise<{ success: boolean; message: string } | null> {
  // 客户端验证：持卡人姓名格式
  if (!validateAccountName(params.account_name)) {
    return {
      success: false,
      message: "持卡人姓名必须为2-6个汉字"
    };
  }

  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/bankcard/update`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("UpdateBankCard API failed:", err);
    return null;
  }
}
```

- [ ] **步骤 4: 验证 TypeScript 编译**

运行：
```bash
cd driverfrontend
npm run build
```

预期：编译成功，无 TypeScript 错误

- [ ] **步骤 5: 提交更改**

```bash
git add driverfrontend/src/app/api/wallet.ts
git commit -m "feat(api): add client-side account name validation for bank card"
```

---

## Task 4: 集成测试

**文件：**
- Test: 手动测试前后端集成

- [ ] **步骤 1: 启动后端服务**

运行：
```bash
cd taketaxi/srvDriver
go run ./cmd/main.go -config=configs/config.yaml
```

预期：服务启动在 :8001

运行：
```bash
cd taketaxi/bffDriver
go run ./cmd/main.go -config=configs/config.yaml
```

预期：服务启动在 :8080

- [ ] **步骤 2: 启动前端开发服务器**

运行：
```bash
cd driverfrontend
npm run dev
```

预期：前端启动在 http://localhost:5173

- [ ] **步骤 3: 测试前端验证 - 有效姓名**

使用浏览器开发者工具控制台测试：

```javascript
// 测试有效姓名（2个汉字）
const result1 = await fetch('http://localhost:8080/api/v1/driver/bankcard', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    driver_id: 1,
    bank_name: '工商银行',
    bank_card_no: '6222021234567890123',
    account_name: '张三',
    card_type: 1
  })
});
console.log(await result1.json());
```

预期：如果司机已实名认证且姓名匹配，返回 `{success: true}`；如果姓名不匹配，返回错误码 50024

- [ ] **步骤 4: 测试前端验证 - 姓名过短**

```javascript
// 测试姓名过短（1个汉字）
import { bindBankCard } from './api/wallet';
const result = await bindBankCard({
  driver_id: 1,
  bank_name: '工商银行',
  bank_card_no: '6222021234567890123',
  account_name: '张',
  card_type: 1
});
console.log(result);
```

预期：返回 `{success: false, message: "持卡人姓名必须为2-6个汉字"}`（前端验证拦截，不发送请求）

- [ ] **步骤 5: 测试前端验证 - 姓名过长**

```javascript
// 测试姓名过长（7个汉字）
const result = await bindBankCard({
  driver_id: 1,
  bank_name: '工商银行',
  bank_card_no: '6222021234567890123',
  account_name: '爱新觉罗启星明',
  card_type: 1
});
console.log(result);
```

预期：返回 `{success: false, message: "持卡人姓名必须为2-6个汉字"}`

- [ ] **步骤 6: 测试前端验证 - 包含非汉字**

```javascript
// 测试包含英文字母
const result = await bindBankCard({
  driver_id: 1,
  bank_name: '工商银行',
  bank_card_no: '6222021234567890123',
  account_name: 'Zhang San',
  card_type: 1
});
console.log(result);
```

预期：返回 `{success: false, message: "持卡人姓名必须为2-6个汉字"}`

- [ ] **步骤 7: 测试后端验证 - 绕过前端直接调用 API**

使用 curl 直接调用后端 API，绕过前端验证：

```bash
curl -X PUT http://localhost:8080/api/v1/driver/bankcard \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": 1,
    "bank_name": "工商银行",
    "bank_card_no": "6222021234567890123",
    "account_name": "张",
    "card_type": 1
  }'
```

预期：返回 `{"success":false,"message":"持卡人姓名格式不正确，必须为2-6个汉字"}`（后端验证生效）

- [ ] **步骤 8: 测试 UpdateBankCard 方法**

```bash
curl -X PUT http://localhost:8080/api/v1/driver/bankcard/update \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": 1,
    "bank_name": "工商银行",
    "bank_card_no": "6222021234567890123",
    "account_name": "李",
    "card_type": 1
  }'
```

预期：返回 `{"success":false,"message":"持卡人姓名格式不正确，必须为2-6个汉字"}`

- [ ] **步骤 9: 验证有效姓名范围**

测试边界情况：

```bash
# 2个汉字（最小有效）
curl -X PUT http://localhost:8080/api/v1/driver/bankcard \
  -H "Content-Type: application/json" \
  -d '{"driver_id": 1, "bank_name": "工商银行", "bank_card_no": "6222021234567890123", "account_name": "张三", "card_type": 1}'

# 6个汉字（最大有效）
curl -X PUT http://localhost:8080/api/v1/driver/bankcard \
  -H "Content-Type: application/json" \
  -d '{"driver_id": 1, "bank_name": "工商银行", "bank_card_no": "6222021234567890123", "account_name": "爱新觉罗启星", "card_type": 1}'
```

预期：两个请求都通过格式验证（可能因其他业务规则失败，但不应该是格式错误）

- [ ] **步骤 10: 记录测试结果**

创建测试报告文档：

```bash
cat > docs/test-results-bankcard-name-validation.md << 'EOF'
# 银行卡持卡人姓名验证测试报告

**测试日期：** 2026-05-03
**测试人员：** [填写]

## 测试用例

| 测试场景 | 输入姓名 | 预期结果 | 实际结果 | 状态 |
|---------|---------|---------|---------|------|
| 前端验证-有效姓名(2字) | "张三" | 通过验证 | [填写] | [✅/❌] |
| 前端验证-有效姓名(6字) | "爱新觉罗启星" | 通过验证 | [填写] | [✅/❌] |
| 前端验证-姓名过短 | "张" | 拦截并提示错误 | [填写] | [✅/❌] |
| 前端验证-姓名过长 | "爱新觉罗启星明" | 拦截并提示错误 | [填写] | [✅/❌] |
| 前端验证-包含英文 | "Zhang San" | 拦截并提示错误 | [填写] | [✅/❌] |
| 后端验证-姓名过短 | "张" | 返回错误码40007 | [填写] | [✅/❌] |
| 后端验证-姓名过长 | "爱新觉罗启星明" | 返回错误码40007 | [填写] | [✅/❌] |
| 后端验证-包含数字 | "张三123" | 返回错误码40007 | [填写] | [✅/❌] |

## 测试结论

[填写测试结论]
EOF
```

- [ ] **步骤 11: 提交测试文档**

```bash
git add docs/test-results-bankcard-name-validation.md
git commit -m "docs: add bank card name validation test report"
```

---

## 验收标准

1. ✅ 后端错误码 `ErrInvalidAccountName` (40007) 已添加
2. ✅ `BindBankCard` 方法包含姓名格式验证
3. ✅ `UpdateBankCard` 方法包含姓名格式验证
4. ✅ 前端 `bindBankCard` 函数包含客户端验证
5. ✅ 前端 `updateBankCard` 函数包含客户端验证
6. ✅ 验证规则：2-6个汉字，正则表达式 `^[一-龥]{2,6}$`
7. ✅ 前端验证失败时返回友好错误提示
8. ✅ 后端验证失败时返回错误码 40007
9. ✅ 所有测试用例通过
10. ✅ 代码已提交到 git

## 注意事项

1. **正则表达式差异**：Go 使用 `\x{4e00}` 语法，JavaScript/TypeScript 使用 `一` 语法
2. **错误码顺序**：确保 40007 未被占用，如有冲突需调整
3. **验证顺序**：格式验证在实名认证姓名一致性校验之前执行
4. **前端验证**：客户端验证提供即时反馈，但不能替代后端验证
5. **测试环境**：需要确保司机已完成实名认证才能测试完整流程
