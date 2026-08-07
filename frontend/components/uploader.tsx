import { useState, useRef, useCallback } from "react";
import { ImagePlus, X } from "lucide-react";
import { compressImage } from "@/lib/image-compress";

interface UploaderProps {
  onImageChange: (file: File | null) => void;
  disabled?: boolean;
}

export function Uploader({ onImageChange, disabled }: UploaderProps) {
  const [preview, setPreview] = useState<string | null>(null);
  const [isCompressing, setIsCompressing] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFile = async (file: File) => {
    if (!file.type.startsWith("image/")) return;
    
    // Create immediate preview for UX
    const objectUrl = URL.createObjectURL(file);
    setPreview(objectUrl);
    
    setIsCompressing(true);
    try {
      const compressed = await compressImage(file);
      onImageChange(compressed);
    } catch (err) {
      console.error("Compression failed", err);
      // Fallback to original if small enough, or just let backend reject
      if (file.size < 5 * 1024 * 1024) {
        onImageChange(file);
      } else {
        alert("Ảnh quá lớn và không thể nén. Vui lòng chọn ảnh khác.");
        clearImage();
      }
    } finally {
      setIsCompressing(false);
    }
  };

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const onDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const onDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (disabled) return;
    
    const file = e.dataTransfer.files?.[0];
    if (file) handleFile(file);
  }, [disabled]);

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
  };

  const clearImage = () => {
    setPreview(null);
    onImageChange(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <div className="w-full relative">
      <input
        type="file"
        ref={fileInputRef}
        onChange={onChange}
        accept="image/jpeg, image/png, image/webp"
        className="hidden"
        disabled={disabled || isCompressing}
      />
      
      {preview ? (
        <div className="relative border border-line rounded-xl overflow-hidden bg-surface flex items-center justify-center h-[180px]">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img 
            src={preview} 
            alt="Preview" 
            className={`max-h-full object-contain ${isCompressing ? 'opacity-50 blur-sm' : ''} transition-all duration-300`} 
          />
          {isCompressing && (
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="px-3 py-1 bg-ink/70 text-white text-xs font-semibold rounded-full shadow-sm">
                Đang xử lý ảnh...
              </span>
            </div>
          )}
          <button
            type="button"
            onClick={clearImage}
            disabled={disabled}
            className="absolute top-2 right-2 w-8 h-8 bg-ink/50 hover:bg-ink/70 text-white rounded-full flex items-center justify-center transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
      ) : (
        <div 
          onClick={() => !disabled && fileInputRef.current?.click()}
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
          className={`border-2 border-dashed rounded-xl bg-surface p-[26px_16px] text-center transition-all duration-150 ${
            isDragging 
              ? 'border-brand bg-brand-tint scale-[1.02]' 
              : 'border-[#BBD4D2] hover:border-brand hover:bg-brand-tint'
          } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
          role="button"
          tabIndex={0}
          aria-label="Tải ảnh màn hình"
        >
          <div className="w-[46px] h-[46px] rounded-xl bg-brand-tint text-brand flex items-center justify-center mx-auto mb-3 pointer-events-none">
            <ImagePlus className="w-6 h-6" />
          </div>
          <p className="text-[17px] font-semibold m-0 mb-0.5 pointer-events-none">Tải ảnh màn hình</p>
          <p className="text-[14px] text-muted m-0 mb-2.5 pointer-events-none">Kéo thả hoặc chạm để chọn</p>
          <p className="text-[12px] text-[#94A3B8] tracking-[0.02em] pointer-events-none">JPG · PNG · tối đa 5MB</p>
        </div>
      )}
    </div>
  );
}
