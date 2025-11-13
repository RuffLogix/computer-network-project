"use client";

import { useRef, useState } from "react";
import { Upload, Paperclip } from "lucide-react";
import { MAX_FILE_SIZE, ALLOWED_FILE_TYPES, API_BASE_URL } from "@/constants";

interface FileUploadProps {
  userId: number;
  onUpload: (url: string, fileName: string, fileSize: number) => void;
}

export function FileUpload({ onUpload, userId }: FileUploadProps) {
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > MAX_FILE_SIZE) {
      alert("File size must be less than 10MB");
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      return;
    }

    // Check MIME type or file extension
    const fileExtension = file.name.split(".").pop()?.toLowerCase();
    const allowedExtensions = [
      "pdf",
      "doc",
      "docx",
      "xls",
      "xlsx",
      "ppt",
      "pptx",
      "txt",
      "zip",
      "rar",
      "7z",
      "csv",
      "json",
      "xml",
      "html",
      "css",
      "js",
      "ts",
      "jsx",
      "tsx",
      "py",
      "java",
      "cpp",
      "c",
      "h",
      "md",
      "rtf",
    ];

    const isValidType = ALLOWED_FILE_TYPES.includes(file.type);
    const isValidExtension =
      fileExtension && allowedExtensions.includes(fileExtension);

    if (!isValidType && !isValidExtension) {
      alert(
        "Please select a valid file type:\n" +
          "Documents: PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, RTF\n" +
          "Text: TXT, CSV, JSON, XML, HTML, CSS, MD\n" +
          "Code: JS, TS, JSX, TSX, PY, JAVA, CPP, C, H\n" +
          "Archives: ZIP, RAR, 7Z"
      );
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      return;
    }

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
        onUpload(data.url, file.name, file.size);
      } else {
        throw new Error("Malformed upload response");
      }
    } catch (error) {
      console.error("Upload failed:", error);
      alert("Failed to upload file");
    } finally {
      setUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  return (
    <div className="relative">
      <input
        ref={fileInputRef}
        type="file"
        onChange={handleFileSelect}
        className="hidden"
        disabled={uploading}
      />
      <button
        onClick={() => fileInputRef.current?.click()}
        className="p-2 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
        disabled={uploading}
        title="Upload file"
      >
        {uploading ? (
          <Upload className="w-5 h-5 animate-pulse" />
        ) : (
          <Paperclip className="w-5 h-5" />
        )}
      </button>
    </div>
  );
}
