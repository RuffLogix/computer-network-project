"use client";

import NextImage from "next/image";
import { useRef, useState } from "react";
import { Image as ImageIcon, X, Upload } from "lucide-react";
import {
  MAX_FILE_SIZE,
  ALLOWED_IMAGE_TYPES,
  ALLOWED_VIDEO_TYPES,
  API_BASE_URL,
} from "@/constants";

interface MediaUploadProps {
  userId: number;
  onUpload: (url: string, type: "image" | "video") => void;
}

export function MediaUpload({ onUpload, userId }: MediaUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [fileType, setFileType] = useState<"image" | "video" | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > MAX_FILE_SIZE) {
      alert("File size must be less than 10MB");
      return;
    }

    const isImage = ALLOWED_IMAGE_TYPES.includes(file.type);
    const isVideo = ALLOWED_VIDEO_TYPES.includes(file.type);

    if (!isImage && !isVideo) {
      alert("Please select an image or video file");
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      setPreview(e.target?.result as string);
      setFileType(isImage ? "image" : "video");
    };
    reader.readAsDataURL(file);

    setUploading(true);
    const formData = new FormData();
    formData.append("file", file);

    try {
      const token = localStorage.getItem("auth_token");
      const response = await fetch(`${API_BASE_URL}/api/upload`, {
        method: "POST",
        headers: {
          Authorization: token ? `Bearer ${token}` : "",
          "X-User-ID": String(userId),
        },
        body: formData,
      });

      if (!response.ok) {
        throw new Error("Upload failed");
      }

      const data = await response.json();
      if (data.url) {
        onUpload(data.url, isImage ? "image" : "video");
      } else {
        throw new Error("Malformed upload response");
      }
    } catch (error) {
      console.error("Upload failed:", error);
      alert("Failed to upload file");
    } finally {
      setUploading(false);
    }
  };

  const clearPreview = () => {
    setPreview(null);
    setFileType(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <div className="relative">
      {preview ? (
        <div className="relative inline-block">
          {fileType === "image" && (
            <NextImage
              src={preview}
              alt="Preview"
              width={320}
              height={240}
              unoptimized
              className="h-auto max-w-xs rounded-lg"
            />
          )}
          {fileType === "video" && (
            <video
              src={preview}
              controls
              className="max-w-xs max-h-48 rounded-lg"
            />
          )}
          <button
            onClick={clearPreview}
            className="absolute top-2 right-2 p-1 bg-red-500 text-white rounded-full hover:bg-red-600"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      ) : (
        <div className="flex gap-2">
          <input
            ref={fileInputRef}
            type="file"
            accept={[...ALLOWED_IMAGE_TYPES, ...ALLOWED_VIDEO_TYPES].join(",")}
            onChange={handleFileSelect}
            className="hidden"
            disabled={uploading}
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            className="p-2 bg-gray-200 dark:bg-gray-700 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            disabled={uploading}
          >
            {uploading ? (
              <Upload className="w-5 h-5 animate-pulse" />
            ) : (
              <ImageIcon className="w-5 h-5" />
            )}
          </button>
        </div>
      )}
    </div>
  );
}
