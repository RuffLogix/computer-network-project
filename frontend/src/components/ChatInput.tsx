"use client";

import { useState, useRef, useEffect } from "react";
import { Send, X, Smile } from "lucide-react";
import { MediaUpload } from "./MediaUpload";
import { StickerPicker } from "./StickerPicker";
import { Message } from "@/types";

interface ChatInputProps {
  userId: number;
  onSendMessage: (
    content: string,
    type?: string,
    mediaUrl?: string,
    replyToId?: number
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
  const [mediaType, setMediaType] = useState<
    "image" | "video" | "sticker" | "text"
  >("text");
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
    if (!message.trim() && !mediaUrl) return;

    onSendMessage(
      message,
      mediaUrl ? mediaType : "text",
      mediaUrl || undefined,
      replyTo?.id
    );

    setMessage("");
    setMediaUrl("");
    setMediaType("text");
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
    type: "image" | "video" | "sticker"
  ) => {
    setMediaUrl(url);
    setMediaType(type);
  };

  const handleStickerSelect = (stickerUrl: string) => {
    onSendMessage("", "sticker", stickerUrl);
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

      <div className="flex items-end gap-2">
        <div className="flex items-center gap-1">
          <MediaUpload userId={userId} onUpload={handleMediaUpload} />
          <div className="relative">
            <button
              onClick={() => {
                console.log(
                  "Sticker button clicked, current state:",
                  showStickerPicker
                );
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

      {mediaUrl && (
        <div className="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {mediaType === "image" ? "📷" : mediaType === "video" ? "🎥" : "😊"}{" "}
          {mediaType === "sticker" ? "Sticker attached" : "Media attached"}
        </div>
      )}
    </div>
  );
}
