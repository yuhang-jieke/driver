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

// 校验文件
export function validateFile(file: File): ValidationResult {
  if (!ALLOWED_TYPES.includes(file.type)) {
    return { valid: false, error: '仅支持 jpg/png 格式' };
  }
  if (file.size > MAX_FILE_SIZE) {
    return { valid: false, error: '图片大小不能超过 2MB' };
  }
  return { valid: true };
}

// DataURL 转 File
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

// 上传图片
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
