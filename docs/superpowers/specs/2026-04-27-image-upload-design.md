# 前端图片上传功能设计

## 概述

补齐前端图片上传功能，支持司机头像、车辆照片、证件资料等场景，与后端已有上传接口对接。

## 后端接口

- **上传接口**: `POST /api/v1/upload`
- **删除接口**: `DELETE /api/v1/upload?path=xxx`
- **业务类型**: avatar / idcard / license / vehicle / face
- **返回格式**: `{ code, msg, data: { url, path, file_size } }`

## 校验限制

与后端保持一致：

| 限制项 | 值 |
|--------|-----|
| 最大文件大小 | 2MB |
| 允许格式 | jpg, png |
| MIME 类型 | image/jpeg, image/png |

## 文件结构

```
driverfrontend/src/app/
├── api/
│   └── upload.ts          # 上传 API 封装
├── components/
│   ├── ImageUploader.tsx  # 图片上传主组件
│   └── DriverApp.tsx      # 现有文件（集成上传）
└── store.tsx              # 现有文件（可能扩展 avatar 字段）
```

## 数据流

```
选择图片 → 预览裁剪 → 调用 upload API → 更新本地状态/显示
```

## API 封装设计

```typescript
// 业务类型
type BizType = 'avatar' | 'idcard' | 'license' | 'vehicle' | 'face';

// 上传结果
interface UploadResult {
  url: string;
  path: string;
  file_size: number;
}

// 校验常量（与后端一致）
const MAX_FILE_SIZE = 2 * 1024 * 1024;  // 2MB
const ALLOWED_TYPES = ['image/jpeg', 'image/png'];

// 主函数
async function uploadImage(file: File, bizType: BizType): Promise<UploadResult>

// 校验函数
function validateFile(file: File): { valid: boolean; error?: string }

// 裁剪后转换
function dataURLtoFile(dataURL: string, filename: string): File
```

## ImageUploader 组件设计

### Props 接口

```typescript
interface ImageUploaderProps {
  bizType: BizType;                              // 业务类型
  aspectRatio?: 'square' | 'portrait' | 'free'; // 裁剪比例
  currentImage?: string;                         // 当前图片 URL
  onUploadSuccess: (result: UploadResult) => void;
  onUploadError?: (error: string) => void;
}
```

### 组件状态

```typescript
type Stage = 'idle' | 'select' | 'crop' | 'uploading';

// idle      - 初始状态，显示占位或当前图片
// select    - 选择图片来源（相册/相机）
// crop      - 裁剪预览
// uploading - 上传中，显示进度
```

### 功能列表

| 功能 | 实现方式 |
|------|----------|
| 选择图片 | `<input type="file" accept="image/jpeg,image/png">` |
| 相机拍照 | 移动端 `capture="environment"` 属性 |
| 裁剪交互 | Canvas 绘制 + 触摸拖拽调整裁剪框 |
| 上传进度 | fetch + 状态显示 |
| 头像裁剪 | 圆形遮罩预览 |

### UI 结构

**idle 阶段：**
- 显示占位图标或当前图片
- 点击触发选择

**crop 阶段（全屏弹窗）：**
- 图片裁剪区域
- 拖拽调整裁剪框
- 取消/确认按钮

**uploading 阶段：**
- 显示上传进度
- 加载动画

## 页面集成

### 1. DriverMe 头像上传

- 头像区域改为可点击
- 点击弹出 ImageUploader，bizType='avatar'
- 上传成功后更新显示
- 裁剪比例：正方形

### 2. DriverCarView 车辆照片

- 新增"车辆照片"卡片
- 支持上传车辆外观照片
- bizType='vehicle'
- 裁剪比例：自由

### 3. 设置页新增证件区域

- 在"账户"区块下方新增"证件资料"
- 身份证正/反面、驾驶证
- bizType='idcard' / 'license'
- 裁剪比例：矩形

## 实现要点

### 裁剪实现

使用 Canvas 实现裁剪功能：

1. 加载图片到 Canvas
2. 绘制裁剪框（可拖拽调整）
3. 用户确认后裁剪并导出
4. 转换为 File 对象上传

### 错误处理

- 文件类型不正确：提示"仅支持 jpg/png 格式"
- 文件过大：提示"图片大小不能超过 2MB"
- 上传失败：提示"上传失败，请重试"

### 样式风格

与现有 DriverApp.tsx 保持一致：
- 使用 Tailwind CSS
- 渐变色按钮、圆角卡片
- emerald/teal 主题色
