"use client";

import { useState, useRef, useEffect } from "react";
import { Send, X, Smile, FileText } from "lucide-react";
import { MediaUpload } from "./MediaUpload";
import { FileUpload } from "./FileUpload";
import { StickerPicker } from "./StickerPicker";
import { Message } from "@/types";

interface ChatInputProps {
  userId: number;
  onSendMessage: (
    content: string,
    type?: string,
    mediaUrl?: string,
    replyToId?: number,
    fileName?: string,
    fileSize?: number
  ) => void;
  onTyping: (isTyping: boolean) => void;
  replyTo?: Message | null;
  onCancelReply?: () => void;
}

export function ChatInput({
  userId,
  onSendMessage,
  onTyping,
  replyTo,
  onCancelReply,
}: ChatInputProps) {
  const [message, setMessage] = useState("");
  const [mediaUrl, setMediaUrl] = useState<string>("");
  const [mediaPreview, setMediaPreview] = useState<string>("");
  const [mediaType, setMediaType] = useState<
    "image" | "video" | "sticker" | "file" | "text"
  >("text");
  const [fileName, setFileName] = useState<string>("");
  const [fileSize, setFileSize] = useState<number>(0);
  const [showStickerPicker, setShowStickerPicker] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (replyTo) {
      inputRef.current?.focus();
    }
  }, [replyTo]);

  const handleInputChange = (value: string) => {
    setMessage(value);

    onTyping(true);

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    typingTimeoutRef.current = setTimeout(() => {
      onTyping(false);
    }, 1000);
  };

  const handleSend = () => {
    // Allow sending if there's a message OR media (including files)
    if (!message.trim() && !mediaUrl) return;

    onSendMessage(
      message,
      mediaUrl ? mediaType : "text",
      mediaUrl || undefined,
      replyTo?.id,
      fileName || undefined,
      fileSize || undefined
    );

    setMessage("");
    setMediaUrl("");
    setMediaPreview("");
    setMediaType("text");
    setFileName("");
    setFileSize(0);
    onTyping(false);

    if (onCancelReply) {
      onCancelReply();
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleMediaUpload = (
    url: string,
    type: "image" | "video" | "sticker",
    preview?: string
  ) => {
    setMediaUrl(url);
    setMediaType(type);
    setMediaPreview(preview || url);
  };

  const handleFileUpload = (url: string, name: string, size: number) => {
    setMediaUrl(url);
    setMediaType("file");
    setFileName(name);
    setFileSize(size);
  };

  const handleStickerSelect = (stickerUrl: string) => {
    onSendMessage("", "sticker", stickerUrl);
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  const clearAttachment = () => {
    setMediaUrl("");
    setMediaPreview("");
    setMediaType("text");
    setFileName("");
    setFileSize(0);
  };

  return (
    <div className="border-t border-gray-200 dark:border-gray-700 p-4">
      {replyTo && (
        <div className="flex items-center justify-between mb-2 p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
          <div className="text-sm">
            <span className="text-gray-500 dark:text-gray-400">
              Replying to:{" "}
            </span>
            <span className="font-semibold">
              {replyTo.content.substring(0, 50)}...
            </span>
          </div>
          <button
            onClick={onCancelReply}
            className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {mediaUrl && (
        <div className="mb-2">
          <div className="p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
            <div className="flex items-start justify-between gap-2 mb-2">
              <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 min-w-0 flex-1">
                {mediaType === "image" && (
                  <>
                    <span className="text-xl flex-shrink-0">📷</span>
                    <span className="truncate">Image attached</span>
                  </>
                )}
                {mediaType === "video" && (
                  <>
                    <span className="text-xl flex-shrink-0">🎥</span>
                    <span className="truncate">Video attached</span>
                  </>
                )}
                {mediaType === "file" && (
                  <>
                    <FileText className="w-4 h-4 flex-shrink-0" />
                    <div className="flex flex-col min-w-0 flex-1">
                      <span className="font-medium text-xs truncate">
                        {fileName}
                      </span>
                      <span className="text-xs text-gray-500 dark:text-gray-500">
                        {formatFileSize(fileSize)}
                      </span>
                    </div>
                  </>
                )}
                {mediaType === "sticker" && (
                  <>
                    <span className="text-xl flex-shrink-0">😊</span>
                    <span className="truncate">Sticker attached</span>
                  </>
                )}
              </div>
              <button
                onClick={clearAttachment}
                className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-full transition-colors flex-shrink-0"
              >
                <X className="w-3 h-3" />
              </button>
            </div>
            {(mediaType === "image" || mediaType === "video") &&
              mediaPreview && (
                <div className="mt-2">
                  {mediaType === "image" && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={mediaPreview}
                      alt="Preview"
                      className="max-h-32 w-auto rounded border border-gray-300 dark:border-gray-600 object-contain"
                      onError={() => {
                        console.error("Image failed to load:", mediaPreview);
                      }}
                    />
                  )}
                  {mediaType === "video" && (
                    <video
                      src={mediaPreview}
                      controls
                      className="max-h-32 w-auto rounded border border-gray-300 dark:border-gray-600"
                      onError={() => {
                        console.error("Video failed to load:", mediaPreview);
                      }}
                    />
                  )}
                </div>
              )}
          </div>
        </div>
      )}

      <div className="flex items-end gap-2">
        <div className="flex items-center gap-1">
          <MediaUpload userId={userId} onUpload={handleMediaUpload} />
          <FileUpload userId={userId} onUpload={handleFileUpload} />
          <div className="relative">
            <button
              onClick={() => {
                setShowStickerPicker(!showStickerPicker);
              }}
              className="p-2 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800"
              title="Add sticker"
            >
              <Smile className="w-5 h-5" />
            </button>
            <StickerPicker
              onSelectSticker={handleStickerSelect}
              isOpen={showStickerPicker}
              onClose={() => setShowStickerPicker(false)}
            />
          </div>
        </div>

        <div className="flex-1">
          <input
            ref={inputRef}
            type="text"
            value={message}
            onChange={(e) => handleInputChange(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type a message..."
            className="w-full px-4 py-2 bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <button
          onClick={handleSend}
          disabled={!message.trim() && !mediaUrl}
          className="p-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-300 dark:disabled:bg-gray-700 disabled:cursor-not-allowed transition-colors"
        >
          <Send className="w-5 h-5" />
        </button>
      </div>
    </div>
  );
}
