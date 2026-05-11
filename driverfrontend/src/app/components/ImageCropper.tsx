import { useState, useRef, useEffect, useCallback } from 'react';
import { X, Check, RotateCcw, ZoomIn, ZoomOut } from 'lucide-react';

interface ImageCropperProps {
  imageSrc: string;
  aspectRatio: 'square' | 'portrait' | 'free';
  onConfirm: (croppedDataURL: string) => void;
  onCancel: () => void;
}

export function ImageCropper({ imageSrc, aspectRatio, onConfirm, onCancel }: ImageCropperProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const imageRef = useRef<HTMLImageElement | null>(null);

  // State
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [imageSize, setImageSize] = useState({ width: 0, height: 0 });
  const [containerSize, setContainerSize] = useState({ width: 0, height: 0 });
  const [imageLoaded, setImageLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Ref for drag handling
  const positionRef = useRef(position);

  useEffect(() => {
    positionRef.current = position;
  }, [position]);

  // Load image
  useEffect(() => {
    const img = new Image();
    let cancelled = false;

    img.onload = () => {
      if (cancelled) return;
      imageRef.current = img;
      setImageSize({ width: img.width, height: img.height });
      setImageLoaded(true);
      setError(null);

      // Initialize scale to fit container
      if (containerRef.current) {
        const container = containerRef.current;
        const containerW = container.clientWidth;
        const containerH = container.clientHeight - 120; // Account for header and controls

        const scaleX = containerW / img.width;
        const scaleY = containerH / img.height;
        const initialScale = Math.min(scaleX, scaleY, 1);

        setScale(initialScale);
        setPosition({ x: 0, y: 0 });
      }
    };

    img.onerror = () => {
      if (cancelled) return;
      setError('图片加载失败');
      setImageLoaded(false);
    };

    img.src = imageSrc;

    return () => {
      cancelled = true;
    };
  }, [imageSrc]);

  // Update container size
  useEffect(() => {
    const updateContainerSize = () => {
      if (containerRef.current) {
        setContainerSize({
          width: containerRef.current.clientWidth,
          height: containerRef.current.clientHeight,
        });
      }
    };

    updateContainerSize();
    window.addEventListener('resize', updateContainerSize);
    return () => window.removeEventListener('resize', updateContainerSize);
  }, []);

  // Draw canvas
  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas || !imageLoaded) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const img = new Image();
    img.src = imageSrc;

    img.onload = () => {
      // Set canvas size
      const displayWidth = containerSize.width || 300;
      const displayHeight = (containerSize.height || 500) - 120; // Account for header and controls

      canvas.width = displayWidth;
      canvas.height = displayHeight;

      // Clear canvas
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      // Fill background
      ctx.fillStyle = '#1a1a1a';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Calculate crop area
      let cropWidth: number;
      let cropHeight: number;

      if (aspectRatio === 'square') {
        cropWidth = cropHeight = Math.min(canvas.width, canvas.height) * 0.8;
      } else if (aspectRatio === 'portrait') {
        cropWidth = Math.min(canvas.width, canvas.height) * 0.7;
        cropHeight = cropWidth * 1.4;
      } else {
        cropWidth = canvas.width * 0.9;
        cropHeight = canvas.height * 0.9;
      }

      const cropX = (canvas.width - cropWidth) / 2;
      const cropY = (canvas.height - cropHeight) / 2;

      // Draw image with transform
      ctx.save();
      ctx.translate(canvas.width / 2 + position.x, canvas.height / 2 + position.y);
      ctx.scale(scale, scale);
      ctx.drawImage(img, -img.width / 2, -img.height / 2);
      ctx.restore();

      // Draw overlay (darken outside crop area)
      ctx.fillStyle = 'rgba(0, 0, 0, 0.6)';

      // Top
      ctx.fillRect(0, 0, canvas.width, cropY);
      // Bottom
      ctx.fillRect(0, cropY + cropHeight, canvas.width, canvas.height - cropY - cropHeight);
      // Left
      ctx.fillRect(0, cropY, cropX, cropHeight);
      // Right
      ctx.fillRect(cropX + cropWidth, cropY, canvas.width - cropX - cropWidth, cropHeight);

      // Draw crop border
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
      ctx.lineWidth = 2;

      if (aspectRatio === 'square') {
        // Draw circular mask for square
        ctx.beginPath();
        ctx.arc(cropX + cropWidth / 2, cropY + cropHeight / 2, cropWidth / 2, 0, Math.PI * 2);
        ctx.stroke();

        // Clear the circle area to show the image
        ctx.save();
        ctx.globalCompositeOperation = 'destination-out';
        ctx.beginPath();
        ctx.arc(cropX + cropWidth / 2, cropY + cropHeight / 2, cropWidth / 2 - 1, 0, Math.PI * 2);
        ctx.fill();
        ctx.restore();

        // Redraw image in the circle
        ctx.save();
        ctx.beginPath();
        ctx.arc(cropX + cropWidth / 2, cropY + cropHeight / 2, cropWidth / 2 - 1, 0, Math.PI * 2);
        ctx.clip();
        ctx.translate(canvas.width / 2 + position.x, canvas.height / 2 + position.y);
        ctx.scale(scale, scale);
        ctx.drawImage(img, -img.width / 2, -img.height / 2);
        ctx.restore();
      } else {
        // Draw rectangle border
        ctx.strokeRect(cropX, cropY, cropWidth, cropHeight);

        // Clear the rectangle area
        ctx.save();
        ctx.globalCompositeOperation = 'destination-out';
        ctx.fillRect(cropX + 1, cropY + 1, cropWidth - 2, cropHeight - 2);
        ctx.restore();

        // Redraw image in the rectangle
        ctx.save();
        ctx.beginPath();
        ctx.rect(cropX + 1, cropY + 1, cropWidth - 2, cropHeight - 2);
        ctx.clip();
        ctx.translate(canvas.width / 2 + position.x, canvas.height / 2 + position.y);
        ctx.scale(scale, scale);
        ctx.drawImage(img, -img.width / 2, -img.height / 2);
        ctx.restore();
      }
    };
  }, [imageSrc, imageLoaded, scale, position, containerSize, aspectRatio]);

  // Redraw on changes
  useEffect(() => {
    draw();
  }, [draw]);

  // Drag handlers
  const handleStart = useCallback((clientX: number, clientY: number) => {
    setIsDragging(true);
    setDragStart({ x: clientX - positionRef.current.x, y: clientY - positionRef.current.y });
  }, []);

  const handleMove = useCallback((clientX: number, clientY: number) => {
    if (!isDragging) return;
    setPosition({
      x: clientX - dragStart.x,
      y: clientY - dragStart.y,
    });
  }, [isDragging, dragStart]);

  const handleEnd = useCallback(() => {
    setIsDragging(false);
  }, []);

  // Mouse events
  const handleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    handleStart(e.clientX, e.clientY);
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    handleMove(e.clientX, e.clientY);
  };

  const handleMouseUp = () => {
    handleEnd();
  };

  // Touch events
  const handleTouchStart = (e: React.TouchEvent) => {
    if (e.touches.length === 1) {
      handleStart(e.touches[0].clientX, e.touches[0].clientY);
    }
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (e.touches.length === 1) {
      handleMove(e.touches[0].clientX, e.touches[0].clientY);
    }
  };

  const handleTouchEnd = () => {
    handleEnd();
  };

  // Zoom controls
  const handleZoomIn = () => {
    setScale((prev) => Math.min(prev * 1.2, 5));
  };

  const handleZoomOut = () => {
    setScale((prev) => Math.max(prev / 1.2, 0.1));
  };

  const resetTransform = () => {
    // Reset to initial scale
    if (containerRef.current && imageSize.width > 0) {
      const containerW = containerRef.current.clientWidth;
      const containerH = containerRef.current.clientHeight - 120;
      const scaleX = containerW / imageSize.width;
      const scaleY = containerH / imageSize.height;
      const initialScale = Math.min(scaleX, scaleY, 1);
      setScale(initialScale);
    }
    setPosition({ x: 0, y: 0 });
  };

  // Confirm and crop
  const handleConfirm = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas || !imageLoaded) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const img = new Image();
    img.src = imageSrc;

    img.onload = () => {
      const displayWidth = containerSize.width || 300;
      const displayHeight = (containerSize.height || 500) - 120;

      // Calculate crop area
      let cropWidth: number;
      let cropHeight: number;

      if (aspectRatio === 'square') {
        cropWidth = cropHeight = Math.min(displayWidth, displayHeight) * 0.8;
      } else if (aspectRatio === 'portrait') {
        cropWidth = Math.min(displayWidth, displayHeight) * 0.7;
        cropHeight = cropWidth * 1.4;
      } else {
        cropWidth = displayWidth * 0.9;
        cropHeight = displayHeight * 0.9;
      }

      const cropX = (displayWidth - cropWidth) / 2;
      const cropY = (displayHeight - cropHeight) / 2;

      // Create output canvas (400px)
      const outputSize = 400;
      const outputCanvas = document.createElement('canvas');
      outputCanvas.width = outputSize;
      outputCanvas.height = aspectRatio === 'portrait' ? Math.round(outputSize * 1.4) : outputSize;

      const outputCtx = outputCanvas.getContext('2d');
      if (!outputCtx) return;

      // Calculate transform
      const scaleX = outputSize / cropWidth;
      const scaleY = outputSize / cropHeight;
      const outputScale = Math.min(scaleX, scaleY);

      // Draw cropped image
      outputCtx.save();

      if (aspectRatio === 'square') {
        // Draw circular clip for square
        outputCtx.beginPath();
        outputCtx.arc(outputSize / 2, outputSize / 2, outputSize / 2, 0, Math.PI * 2);
        outputCtx.clip();
      }

      // Transform and draw
      outputCtx.translate(outputSize / 2, outputCanvas.height / 2);
      outputCtx.scale(scale * outputScale, scale * outputScale);
      outputCtx.translate(
        (position.x / outputScale) / scale,
        (position.y / outputScale) / scale
      );

      // Recalculate for proper positioning
      outputCtx.setTransform(1, 0, 0, 1, 0, 0);
      outputCtx.translate(outputSize / 2, outputCanvas.height / 2);
      outputCtx.scale(scale * outputScale, scale * outputScale);
      outputCtx.drawImage(img, -img.width / 2 - position.x / (scale * outputScale), -img.height / 2 - position.y / (scale * outputScale));

      outputCtx.restore();

      // Get data URL
      const dataURL = outputCanvas.toDataURL('image/jpeg', 0.9);
      onConfirm(dataURL);
    };
  }, [imageSrc, imageLoaded, scale, position, containerSize, aspectRatio, onConfirm]);

  return (
    <div className="fixed inset-0 z-50 bg-black flex flex-col" ref={containerRef}>
      {/* Top bar */}
      <div className="flex items-center justify-between px-4 py-3 bg-gray-900 shrink-0">
        <button
          onClick={onCancel}
          className="w-10 h-10 flex items-center justify-center text-white"
        >
          <X className="w-6 h-6" />
        </button>
        <span className="text-white font-medium">裁剪图片</span>
        <button
          onClick={handleConfirm}
          className="w-10 h-10 flex items-center justify-center text-emerald-400"
        >
          <Check className="w-6 h-6" />
        </button>
      </div>

      {/* Canvas area */}
      <div className="flex-1 flex items-center justify-center overflow-hidden">
        <canvas
          ref={canvasRef}
          className="cursor-grab active:cursor-grabbing touch-none"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
        />
      </div>

      {/* Bottom controls */}
      <div className="flex items-center justify-center gap-8 px-4 py-4 bg-gray-900 shrink-0">
        <button
          onClick={handleZoomOut}
          className="flex flex-col items-center gap-1 text-white"
        >
          <div className="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center">
            <ZoomOut className="w-5 h-5" />
          </div>
          <span className="text-xs text-gray-400">缩小</span>
        </button>

        <button
          onClick={resetTransform}
          className="flex flex-col items-center gap-1 text-white"
        >
          <div className="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center">
            <RotateCcw className="w-5 h-5" />
          </div>
          <span className="text-xs text-gray-400">重置</span>
        </button>

        <button
          onClick={handleZoomIn}
          className="flex flex-col items-center gap-1 text-white"
        >
          <div className="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center">
            <ZoomIn className="w-5 h-5" />
          </div>
          <span className="text-xs text-gray-400">放大</span>
        </button>
      </div>
    </div>
  );
}
