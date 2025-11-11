"use client";

import NextImage from "next/image";
import { useRef, useState } from "react";
import { Image as ImageIcon, X, Upload } from "lucide-react";
import {
  MAX_FILE_SIZE,
  MAX_MEDIA_SIZE,
  ALLOWED_IMAGE_TYPES,
  ALLOWED_VIDEO_TYPES,
  API_BASE_URL,
} from "@/constants";

interface MediaUploadProps {
  userId: number;
  onUpload: (url: string, type: "image" | "video", preview?: string) => void;
}

export function MediaUpload({ onUpload, userId }: MediaUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [fileType, setFileType] = useState<"image" | "video" | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const isImage = ALLOWED_IMAGE_TYPES.includes(file.type);
    const isVideo = ALLOWED_VIDEO_TYPES.includes(file.type);

    if (!isImage && !isVideo) {
      alert("Please select an image or video file");
      return;
    }

    const maxSize = isVideo ? MAX_MEDIA_SIZE : MAX_FILE_SIZE;
    if (file.size > maxSize) {
      alert(`File size must be less than ${isVideo ? "50MB" : "10MB"}`);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      return;
    }

    setUploading(true);

    // Create local preview first
    const reader = new FileReader();
    reader.onload = async (e) => {
      const previewUrl = e.target?.result as string;
      setPreview(previewUrl);
      setFileType(isImage ? "image" : "video");

      // Now upload to server
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
          // Pass both server URL and preview
          onUpload(data.url, isImage ? "image" : "video", previewUrl);
          setPreview(null);
          setFileType(null);
        } else {
          throw new Error("Malformed upload response");
        }
      } catch (error) {
        console.error("Upload failed:", error);
        alert("Failed to upload file");
        setPreview(null);
        setFileType(null);
      } finally {
        setUploading(false);
        if (fileInputRef.current) {
          fileInputRef.current.value = "";
        }
      }
    };
    reader.readAsDataURL(file);
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
        className="p-2 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
        disabled={uploading}
        title="Upload image or video"
      >
        {uploading ? (
          <Upload className="w-5 h-5 animate-pulse" />
        ) : (
          <ImageIcon className="w-5 h-5" />
        )}
      </button>
    </div>
  );
}
