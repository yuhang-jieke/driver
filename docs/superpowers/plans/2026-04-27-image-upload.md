# 前端图片上传功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐前端图片上传功能，支持司机头像、车辆照片、证件资料上传

**Architecture:** 纯 React + Canvas 实现裁剪，无第三方依赖。API 封装层处理上传请求，ImageUploader 组件处理图片选择、裁剪、上传全流程。

**Tech Stack:** React 18, TypeScript, Canvas API, Tailwind CSS, Vite

---

## 文件结构

```
driverfrontend/src/app/
├── api/
│   └── upload.ts              # 新建：上传 API 封装
├── components/
│   ├── ImageUploader.tsx      # 新建：图片上传组件
│   ├── ImageCropper.tsx       # 新建：裁剪组件
│   └── DriverApp.tsx          # 修改：集成上传功能
└── store.tsx                  # 修改：扩展 avatar 字段
```

---

### Task 1: 创建上传 API 封装

**Files:**
- Create: `driverfrontend/src/app/api/upload.ts`

- [ ] **Step 1: 创建 upload.ts 文件，定义类型和常量**

```typescript
const API_BASE = "http://localhost:8080";

// 业务类型
export type BizType = 'avatar' | 'idcard' | 'license' | 'vehicle' | 'face';

// 上传结果
export interface UploadResult {
  url: string;
  path: string;
  file_size: number;
}

// 上传响应
interface UploadResponse {
  code: number;
  msg: string;
  data?: UploadResult;
}

// 校验常量（与后端一致）
const MAX_FILE_SIZE = 2 * 1024 * 1024;  // 2MB
const ALLOWED_TYPES = ['image/jpeg', 'image/png'];

// 校验结果
interface ValidationResult {
  valid: boolean;
  error?: string;
}
```

- [ ] **Step 2: 实现 validateFile 校验函数**

```typescript
export function validateFile(file: File): ValidationResult {
  if (!ALLOWED_TYPES.includes(file.type)) {
    return { valid: false, error: '仅支持 jpg/png 格式' };
  }
  if (file.size > MAX_FILE_SIZE) {
    return { valid: false, error: '图片大小不能超过 2MB' };
  }
  return { valid: true };
}
```

- [ ] **Step 3: 实现 dataURLtoFile 转换函数**

```typescript
export function dataURLtoFile(dataURL: string, filename: string): File {
  const arr = dataURL.split(',');
  const mime = arr[0].match(/:(.*?);/)?.[1] || 'image/png';
  const bstr = atob(arr[1]);
  let n = bstr.length;
  const u8arr = new Uint8Array(n);
  while (n--) {
    u8arr[n] = bstr.charCodeAt(n);
  }
  return new File([u8arr], filename, { type: mime });
}
```

- [ ] **Step 4: 实现 uploadImage 主函数**

```typescript
export async function uploadImage(file: File, bizType: BizType): Promise<UploadResult> {
  const validation = validateFile(file);
  if (!validation.valid) {
    throw new Error(validation.error);
  }

  const formData = new FormData();
  formData.append('file', file);
  formData.append('biz_type', bizType);

  const response = await fetch(`${API_BASE}/api/v1/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    throw new Error('上传失败，请重试');
  }

  const result: UploadResponse = await response.json();
  if (result.code !== 0 || !result.data) {
    throw new Error(result.msg || '上传失败');
  }

  return result.data;
}
```

- [ ] **Step 5: 提交代码**

```bash
git add driverfrontend/src/app/api/upload.ts
git commit -m "feat: add upload API wrapper with validation"
```

---

### Task 2: 创建图片裁剪组件

**Files:**
- Create: `driverfrontend/src/app/components/ImageCropper.tsx`

- [ ] **Step 1: 创建 ImageCropper.tsx，定义 Props 和状态**

```typescript
import { useState, useRef, useEffect, useCallback } from 'react';
import { X, Check, RotateCcw } from 'lucide-react';

interface ImageCropperProps {
  imageSrc: string;
  aspectRatio: 'square' | 'portrait' | 'free';
  onConfirm: (croppedDataURL: string) => void;
  onCancel: () => void;
}

export function ImageCropper({ imageSrc, aspectRatio, onConfirm, onCancel }: ImageCropperProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [imageSize, setImageSize] = useState({ width: 0, height: 0 });
  const [containerSize, setContainerSize] = useState({ width: 300, height: 300 });
```

- [ ] **Step 2: 实现图片加载和 Canvas 初始化**

```typescript
  // 加载图片
  useEffect(() => {
    const img = new Image();
    img.onload = () => {
      setImageSize({ width: img.width, height: img.height });
      // 计算适应容器的缩放比例
      const containerW = containerRef.current?.clientWidth || 300;
      const containerH = aspectRatio === 'portrait' ? containerW * 1.5 : containerW;
      setContainerSize({ width: containerW, height: containerH });
      const scaleW = containerW / img.width;
      const scaleH = containerH / img.height;
      const fitScale = Math.min(scaleW, scaleH, 1);
      setScale(fitScale);
      setPosition({ x: 0, y: 0 });
    };
    img.src = imageSrc;
  }, [imageSrc, aspectRatio]);

  // 绘制 Canvas
  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext('2d');
    if (!canvas || !ctx) return;

    canvas.width = containerSize.width;
    canvas.height = containerSize.height;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    const img = new Image();
    img.onload = () => {
      ctx.save();
      ctx.translate(position.x, position.y);
      ctx.scale(scale, scale);
      ctx.drawImage(img, 0, 0);
      ctx.restore();

      // 绘制裁剪框
      if (aspectRatio === 'square') {
        const size = Math.min(canvas.width, canvas.height) * 0.8;
        const x = (canvas.width - size) / 2;
        const y = (canvas.height - size) / 2;
        
        // 遮罩层
        ctx.fillStyle = 'rgba(0,0,0,0.5)';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // 裁剪区域
        ctx.clearRect(x, y, size, size);
        ctx.save();
        ctx.translate(position.x, position.y);
        ctx.scale(scale, scale);
        ctx.drawImage(img, 0, 0);
        ctx.restore();
        
        // 圆形边框
        ctx.beginPath();
        ctx.arc(x + size/2, y + size/2, size/2, 0, Math.PI * 2);
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 2;
        ctx.stroke();
      }
    };
    img.src = imageSrc;
  }, [imageSrc, scale, position, containerSize, aspectRatio]);

  useEffect(() => {
    draw();
  }, [draw]);
```

- [ ] **Step 3: 实现拖拽和缩放手势**

```typescript
  // 触摸/鼠标拖拽
  const handleStart = (clientX: number, clientY: number) => {
    setIsDragging(true);
    setDragStart({ x: clientX - position.x, y: clientY - position.y });
  };

  const handleMove = (clientX: number, clientY: number) => {
    if (!isDragging) return;
    setPosition({
      x: clientX - dragStart.x,
      y: clientY - dragStart.y,
    });
  };

  const handleEnd = () => {
    setIsDragging(false);
  };

  // 缩放
  const handleZoom = (delta: number) => {
    setScale(prev => Math.max(0.5, Math.min(3, prev + delta)));
  };

  const resetTransform = () => {
    const containerW = containerRef.current?.clientWidth || 300;
    const containerH = aspectRatio === 'portrait' ? containerW * 1.5 : containerW;
    const scaleW = containerW / imageSize.width;
    const scaleH = containerH / imageSize.height;
    const fitScale = Math.min(scaleW, scaleH, 1);
    setScale(fitScale);
    setPosition({ x: 0, y: 0 });
  };
```

- [ ] **Step 4: 实现裁剪确认逻辑**

```typescript
  const handleConfirm = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const outputCanvas = document.createElement('canvas');
    const outputCtx = outputCanvas.getContext('2d');
    if (!outputCtx) return;

    const outputSize = 400; // 输出图片尺寸
    outputCanvas.width = outputSize;
    outputCanvas.height = aspectRatio === 'portrait' ? outputSize * 1.5 : outputSize;

    const img = new Image();
    img.onload = () => {
      outputCtx.save();
      outputCtx.translate(position.x * (imageSize.width / containerSize.width), position.y * (imageSize.height / containerSize.height));
      outputCtx.scale(imageSize.width / containerSize.width / scale * scale, imageSize.height / containerSize.height / scale * scale);
      
      if (aspectRatio === 'square') {
        // 圆形裁剪
        outputCtx.beginPath();
        outputCtx.arc(outputSize / 2, outputSize / 2, outputSize / 2, 0, Math.PI * 2);
        outputCtx.clip();
      }
      
      outputCtx.drawImage(img, 0, 0);
      outputCtx.restore();

      const dataURL = outputCanvas.toDataURL('image/jpeg', 0.9);
      onConfirm(dataURL);
    };
    img.src = imageSrc;
  };
```

- [ ] **Step 5: 实现 UI 渲染**

```typescript
  return (
    <div className="fixed inset-0 bg-black z-50 flex flex-col">
      {/* 顶部栏 */}
      <div className="flex items-center justify-between px-4 py-3 bg-black/50">
        <button onClick={onCancel} className="text-white">
          <X className="w-6 h-6" />
        </button>
        <span className="text-white font-medium">裁剪图片</span>
        <button onClick={handleConfirm} className="text-emerald-400">
          <Check className="w-6 h-6" />
        </button>
      </div>

      {/* 裁剪区域 */}
      <div 
        ref={containerRef}
        className="flex-1 flex items-center justify-center overflow-hidden"
        onMouseDown={e => handleStart(e.clientX, e.clientY)}
        onMouseMove={e => handleMove(e.clientX, e.clientY)}
        onMouseUp={handleEnd}
        onMouseLeave={handleEnd}
        onTouchStart={e => handleStart(e.touches[0].clientX, e.touches[0].clientY)}
        onTouchMove={e => {
          e.preventDefault();
          handleMove(e.touches[0].clientX, e.touches[0].clientY);
        }}
        onTouchEnd={handleEnd}
      >
        <canvas 
          ref={canvasRef}
          className="max-w-full max-h-full"
          style={{ width: containerSize.width, height: containerSize.height }}
        />
      </div>

      {/* 底部工具栏 */}
      <div className="flex items-center justify-center gap-6 px-4 py-4 bg-black/50">
        <button onClick={() => handleZoom(-0.1)} className="text-white text-sm px-3 py-2 bg-white/10 rounded-lg">
          缩小
        </button>
        <button onClick={resetTransform} className="text-white">
          <RotateCcw className="w-6 h-6" />
        </button>
        <button onClick={() => handleZoom(0.1)} className="text-white text-sm px-3 py-2 bg-white/10 rounded-lg">
          放大
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 6: 提交代码**

```bash
git add driverfrontend/src/app/components/ImageCropper.tsx
git commit -m "feat: add ImageCropper component with canvas-based cropping"
```

---

### Task 3: 创建图片上传主组件

**Files:**
- Create: `driverfrontend/src/app/components/ImageUploader.tsx`

- [ ] **Step 1: 创建 ImageUploader.tsx，定义 Props 和状态**

```typescript
import { useState, useRef } from 'react';
import { Camera, ImagePlus, Loader2 } from 'lucide-react';
import { uploadImage, validateFile, dataURLtoFile, BizType, UploadResult } from '../api/upload';
import { ImageCropper } from './ImageCropper';

interface ImageUploaderProps {
  bizType: BizType;
  aspectRatio?: 'square' | 'portrait' | 'free';
  currentImage?: string;
  onUploadSuccess: (result: UploadResult) => void;
  onUploadError?: (error: string) => void;
}

type Stage = 'idle' | 'crop' | 'uploading';

export function ImageUploader({
  bizType,
  aspectRatio = 'free',
  currentImage,
  onUploadSuccess,
  onUploadError,
}: ImageUploaderProps) {
  const [stage, setStage] = useState<Stage>('idle');
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null);
```

- [ ] **Step 2: 实现文件选择处理**

```typescript
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const validation = validateFile(file);
    if (!validation.valid) {
      setError(validation.error || '文件校验失败');
      onUploadError?.(validation.error || '文件校验失败');
      return;
    }

    const reader = new FileReader();
    reader.onload = (event) => {
      setSelectedImage(event.target?.result as string);
      setStage('crop');
    };
    reader.readAsDataURL(file);

    // 重置 input
    e.target.value = '';
  };

  const handleCameraSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleFileSelect(e);
  };
```

- [ ] **Step 3: 实现裁剪确认和上传逻辑**

```typescript
  const handleCropConfirm = async (croppedDataURL: string) => {
    setStage('uploading');
    setError(null);

    try {
      const file = dataURLtoFile(croppedDataURL, `upload_${Date.now()}.jpg`);
      const result = await uploadImage(file, bizType);
      onUploadSuccess(result);
      setStage('idle');
      setSelectedImage(null);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '上传失败，请重试';
      setError(errorMsg);
      onUploadError?.(errorMsg);
      setStage('idle');
    }
  };

  const handleCropCancel = () => {
    setStage('idle');
    setSelectedImage(null);
  };
```

- [ ] **Step 4: 实现 UI 渲染**

```typescript
  // 裁剪阶段
  if (stage === 'crop' && selectedImage) {
    return (
      <ImageCropper
        imageSrc={selectedImage}
        aspectRatio={aspectRatio}
        onConfirm={handleCropConfirm}
        onCancel={handleCropCancel}
      />
    );
  }

  // 上传中阶段
  if (stage === 'uploading') {
    return (
      <div className="flex flex-col items-center justify-center gap-3 p-6 bg-gray-50 rounded-2xl">
        <Loader2 className="w-8 h-8 text-emerald-500 animate-spin" />
        <span className="text-sm text-gray-500">正在上传...</span>
      </div>
    );
  }

  // 空闲阶段
  return (
    <div className="relative">
      {/* 当前图片或占位 */}
      <div 
        onClick={() => fileInputRef.current?.click()}
        className={`relative overflow-hidden cursor-pointer group ${
          aspectRatio === 'square' 
            ? 'w-20 h-20 rounded-full' 
            : aspectRatio === 'portrait'
            ? 'w-full h-32 rounded-xl'
            : 'w-full h-24 rounded-xl'
        } ${currentImage ? 'bg-gray-100' : 'bg-gray-100 border-2 border-dashed border-gray-300'}`}
      >
        {currentImage ? (
          <img 
            src={currentImage} 
            alt="已上传图片" 
            className={`w-full h-full object-cover ${aspectRatio === 'square' ? 'rounded-full' : ''}`}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-gray-400">
            <ImagePlus className="w-6 h-6" />
          </div>
        )}
        
        {/* 悬浮遮罩 */}
        <div className={`absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center ${
          aspectRatio === 'square' ? 'rounded-full' : ''
        }`}>
          <span className="text-white text-xs">点击更换</span>
        </div>
      </div>

      {/* 相机按钮（仅头像模式显示） */}
      {aspectRatio === 'square' && (
        <button 
          onClick={() => cameraInputRef.current?.click()}
          className="absolute bottom-0 right-0 w-7 h-7 bg-emerald-500 rounded-full flex items-center justify-center shadow-md"
        >
          <Camera className="w-4 h-4 text-white" />
        </button>
      )}

      {/* 错误提示 */}
      {error && (
        <div className="mt-2 text-xs text-rose-500">{error}</div>
      )}

      {/* 隐藏的文件输入 */}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png"
        className="hidden"
        onChange={handleFileSelect}
      />
      <input
        ref={cameraInputRef}
        type="file"
        accept="image/jpeg,image/png"
        capture="environment"
        className="hidden"
        onChange={handleCameraSelect}
      />
    </div>
  );
}
```

- [ ] **Step 5: 提交代码**

```bash
git add driverfrontend/src/app/components/ImageUploader.tsx
git commit -m "feat: add ImageUploader component with camera support"
```

---

### Task 4: 集成头像上传到 DriverMe

**Files:**
- Modify: `driverfrontend/src/app/components/DriverApp.tsx` (DriverMe 组件)

- [ ] **Step 1: 在 DriverMe 组件中导入 ImageUploader**

在 DriverApp.tsx 文件顶部添加导入：

```typescript
import { ImageUploader } from './ImageUploader';
import { UploadResult } from '../api/upload';
```

- [ ] **Step 2: 修改 DriverMe 组件 Props 添加 avatar 和 onAvatarChange**

找到 DriverMe 组件定义（约第 737 行），修改 Props 类型并添加头像上传状态：

```typescript
function DriverMe({ driver, onNav, onWithdraw, onToast, avatar, onAvatarChange }: any) {
```

- [ ] **Step 3: 替换头像显示区域为 ImageUploader 组件**

找到头像显示区域（约第 743-744 行），将：

```typescript
<div className="w-16 h-16 rounded-full bg-white/30 border-2 border-white/50 flex items-center justify-center text-3xl">👨‍✈️</div>
```

替换为：

```typescript
<ImageUploader
  bizType="avatar"
  aspectRatio="square"
  currentImage={avatar}
  onUploadSuccess={(result: UploadResult) => {
    onAvatarChange(result.url);
    onToast('头像更新成功');
  }}
  onUploadError={(error) => onToast(error)}
/>
```

- [ ] **Step 4: 提交代码**

```bash
git add driverfrontend/src/app/components/DriverApp.tsx
git commit -m "feat: integrate avatar upload in DriverMe"
```

---

### Task 5: 集成车辆照片上传到 DriverCarView

**Files:**
- Modify: `driverfrontend/src/app/components/DriverApp.tsx` (DriverCarView 组件)

- [ ] **Step 1: 在 DriverCarView 组件中添加车辆照片状态**

找到 DriverCarView 组件定义（约第 1002 行），添加状态：

```typescript
function DriverCarView({ driver, onBack, onToast }: any) {
  const [vehiclePhoto, setVehiclePhoto] = useState<string | null>(null);
```

- [ ] **Step 2: 在车辆信息卡下方添加车辆照片上传区域**

在渐变车辆卡片（约第 1010-1016 行）后面，添加车辆照片卡片：

```typescript
      {/* 车辆照片 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b border-gray-50">
          <div className="text-sm font-medium text-gray-700">车辆照片</div>
          <div className="text-[10px] text-gray-400 mt-0.5">上传车辆外观照片</div>
        </div>
        <div className="p-4">
          <ImageUploader
            bizType="vehicle"
            aspectRatio="free"
            currentImage={vehiclePhoto || undefined}
            onUploadSuccess={(result: UploadResult) => {
              setVehiclePhoto(result.url);
              onToast('车辆照片上传成功');
            }}
            onUploadError={(error) => onToast(error)}
          />
        </div>
      </div>
```

- [ ] **Step 3: 提交代码**

```bash
git add driverfrontend/src/app/components/DriverApp.tsx
git commit -m "feat: integrate vehicle photo upload in DriverCarView"
```

---

### Task 6: 在设置页添加证件资料上传区域

**Files:**
- Modify: `driverfrontend/src/app/components/DriverApp.tsx` (DriverSettingsView 组件)

- [ ] **Step 1: 在 DriverSettingsView 组件中添加证件状态**

找到 DriverSettingsView 组件定义（约第 1119 行），添加状态：

```typescript
function DriverSettingsView({ driver, onBack, onToggleOnline, onToast }: any) {
  const [sound, setSound] = useState(true);
  const [autoNav, setAutoNav] = useState(true);
  const [nightMode, setNightMode] = useState(false);
  const [changeMobileOpen, setChangeMobileOpen] = useState(false);
  const [changePasswordOpen, setChangePasswordOpen] = useState(false);
  // 证件照片状态
  const [idCardFront, setIdCardFront] = useState<string | null>(null);
  const [idCardBack, setIdCardBack] = useState<string | null>(null);
  const [licensePhoto, setLicensePhoto] = useState<string | null>(null);
```

- [ ] **Step 2: 在账户区块下方添加证件资料区块**

在账户区块（约第 1171-1185 行）之后，添加证件资料区块：

```typescript
      {/* 证件资料 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">证件资料</div>
        
        {/* 身份证正面 */}
        <div className="px-4 py-3 border-b border-gray-50">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-700">身份证正面</span>
            {idCardFront && <span className="text-[10px] text-emerald-500">已上传</span>}
          </div>
          <ImageUploader
            bizType="idcard"
            aspectRatio="portrait"
            currentImage={idCardFront || undefined}
            onUploadSuccess={(result: UploadResult) => {
              setIdCardFront(result.url);
              onToast('身份证正面上传成功');
            }}
            onUploadError={(error) => onToast(error)}
          />
        </div>
        
        {/* 身份证反面 */}
        <div className="px-4 py-3 border-b border-gray-50">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-700">身份证反面</span>
            {idCardBack && <span className="text-[10px] text-emerald-500">已上传</span>}
          </div>
          <ImageUploader
            bizType="idcard"
            aspectRatio="portrait"
            currentImage={idCardBack || undefined}
            onUploadSuccess={(result: UploadResult) => {
              setIdCardBack(result.url);
              onToast('身份证反面上传成功');
            }}
            onUploadError={(error) => onToast(error)}
          />
        </div>
        
        {/* 驾驶证 */}
        <div className="px-4 py-3">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-700">驾驶证</span>
            {licensePhoto && <span className="text-[10px] text-emerald-500">已上传</span>}
          </div>
          <ImageUploader
            bizType="license"
            aspectRatio="portrait"
            currentImage={licensePhoto || undefined}
            onUploadSuccess={(result: UploadResult) => {
              setLicensePhoto(result.url);
              onToast('驾驶证上传成功');
            }}
            onUploadError={(error) => onToast(error)}
          />
        </div>
      </div>
```

- [ ] **Step 3: 提交代码**

```bash
git add driverfrontend/src/app/components/DriverApp.tsx
git commit -m "feat: add ID card and license upload in settings"
```

---

### Task 7: 更新 DriverApp 主组件传递 avatar 状态

**Files:**
- Modify: `driverfrontend/src/app/components/DriverApp.tsx` (DriverApp 组件)

- [ ] **Step 1: 在 DriverApp 组件中添加 avatar 状态**

找到 DriverApp 组件定义（约第 79 行），添加状态：

```typescript
export function DriverApp() {
  const { orders, drivers, currentDriverId, setDriverOnline, acceptOrder, updateOrderStatus } = useStore();
  const [tab, setTab] = useState<Tab>("home");
  const [meStage, setMeStage] = useState<MeStage>("main");
  const [toast, setToast] = useState<string | null>(null);
  const [selectedHistoryId, setSelectedHistoryId] = useState<string | null>(null);
  const [cancellationAlert, setCancellationAlert] = useState<{ name: string } | null>(null);
  const [withdrawalOpen, setWithdrawalOpen] = useState(false);
  const [sosOpen, setSosOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const [avatar, setAvatar] = useState<string | null>(null);  // 新增
```

- [ ] **Step 2: 更新 DriverMe 调用传递 avatar props**

找到 DriverMe 调用（约第 146-148 行），修改为：

```typescript
{tab === "me" && meStage === "main" && (
  <DriverMe 
    driver={driver} 
    avatar={avatar || undefined} 
    onAvatarChange={setAvatar}
    onNav={s => setMeStage(s as MeStage)}
    onWithdraw={() => setWithdrawalOpen(true)} 
    onToast={showToast} 
  />
)}
```

- [ ] **Step 3: 提交代码**

```bash
git add driverfrontend/src/app/components/DriverApp.tsx
git commit -m "feat: pass avatar state to DriverMe component"
```

---

### Task 8: 手动测试验证

**Files:**
- 无文件修改

- [ ] **Step 1: 启动前端开发服务器**

```bash
cd driverfrontend && npm run dev
```

Expected: 服务器启动在 http://localhost:5173

- [ ] **Step 2: 启动后端服务（如果未运行）**

```bash
cd taketaxi/bffDriver/cmd && go run main.go
```

Expected: 后端服务启动在 http://localhost:8080

- [ ] **Step 3: 测试头像上传**

1. 打开浏览器访问前端
2. 进入"我的"页面
3. 点击头像区域
4. 选择一张 jpg/png 图片
5. 在裁剪界面调整并确认
6. 验证头像更新成功

- [ ] **Step 4: 测试车辆照片上传**

1. 进入"我的" -> "我的车辆"
2. 点击车辆照片上传区域
3. 选择图片并上传
4. 验证上传成功

- [ ] **Step 5: 测试证件上传**

1. 进入"我的" -> "设置"
2. 在证件资料区域上传身份证正反面、驾驶证
3. 验证上传成功

- [ ] **Step 6: 测试错误处理**

1. 尝试上传超过 2MB 的图片 → 应提示"图片大小不能超过 2MB"
2. 尝试上传非 jpg/png 文件 → 应提示"仅支持 jpg/png 格式"
3. 关闭后端服务后上传 → 应提示上传失败

---

### Task 9: 最终提交和清理

**Files:**
- 无新文件

- [ ] **Step 1: 确认所有改动已提交**

```bash
git status
```

Expected: 工作目录干净

- [ ] **Step 2: 查看提交历史**

```bash
git log --oneline -10
```

Expected: 看到所有功能的提交记录

---

## 总结

| Task | 内容 | 文件 |
|------|------|------|
| 1 | 上传 API 封装 | api/upload.ts |
| 2 | 图片裁剪组件 | components/ImageCropper.tsx |
| 3 | 图片上传组件 | components/ImageUploader.tsx |
| 4 | 头像上传集成 | DriverApp.tsx (DriverMe) |
| 5 | 车辆照片集成 | DriverApp.tsx (DriverCarView) |
| 6 | 证件上传集成 | DriverApp.tsx (DriverSettingsView) |
| 7 | 状态传递 | DriverApp.tsx (DriverApp) |
| 8 | 手动测试 | - |
| 9 | 最终清理 | - |
