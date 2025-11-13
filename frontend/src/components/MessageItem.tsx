"use client";

import NextImage from "next/image";
import { useState } from "react";
import {
  Reply,
  Smile,
  MoreVertical,
  Edit,
  Trash,
  Download,
  FileText,
} from "lucide-react";
import { API_BASE_URL, REACTION_EMOJIS } from "@/constants";
import { Message, Reaction } from "@/types";

interface MessageItemProps {
  message: Message;
  reactions?: Reaction[];
  isOwnMessage: boolean;
  userId: number;
  onReply?: (message: Message) => void;
  onReact?: (messageId: number, reactionType: string) => void;
  onEdit?: (messageId: number, content: string) => void;
  onDelete?: (messageId: number) => void;
}

export function MessageItem({
  message,
  reactions = [],
  isOwnMessage,
  userId,
  onReply,
  onReact,
  onEdit,
  onDelete,
}: MessageItemProps) {
  const [showReactions, setShowReactions] = useState(false);
  const [showMenu, setShowMenu] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(message.content);

  const handleEdit = () => {
    if (!editContent.trim()) {
      return;
    }
    onEdit?.(message.id, editContent);
    setIsEditing(false);
  };

  const formatFileSize = (bytes?: number): string => {
    if (!bytes || bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  const getFileIcon = (fileName?: string) => {
    if (!fileName) return <FileText className="w-8 h-8" />;
    const ext = fileName.split(".").pop()?.toLowerCase();
    const iconMap: Record<string, string> = {
      pdf: "📄",
      doc: "📘",
      docx: "📘",
      xls: "📗",
      xlsx: "📗",
      ppt: "📙",
      pptx: "📙",
      txt: "📝",
      zip: "🗜️",
      rar: "🗜️",
      "7z": "🗜️",
    };
    return iconMap[ext || ""] || <FileText className="w-8 h-8" />;
  };

  const handleDownload = async () => {
    if (!message.media_url) return;

    try {
      const response = await fetch(`${API_BASE_URL}${message.media_url}`);
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = message.file_name || "download";
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error("Download failed:", error);
      alert("Failed to download file");
    }
  };

  const renderMedia = () => {
    if (!message.media_url) return null;

    if (message.type === "image") {
      return (
        <NextImage
          src={`${API_BASE_URL}${message.media_url}`}
          alt="Shared image"
          width={400}
          height={300}
          unoptimized
          className="h-auto max-w-sm rounded-lg"
        />
      );
    }

    if (message.type === "sticker") {
      // Check if it's an SVG
      if (message.media_url.endsWith(".svg")) {
        return (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={message.media_url}
            alt="Sticker"
            className="h-24 w-24 object-contain"
            draggable={false}
          />
        );
      }
      return (
        <NextImage
          src={message.media_url}
          alt="Sticker"
          width={100}
          height={100}
          unoptimized
          className="h-auto max-w-xs"
        />
      );
    }

    if (message.type === "video") {
      return (
        <video
          src={`${API_BASE_URL}${message.media_url}`}
          controls
          className="max-w-sm rounded-lg"
        />
      );
    }

    if (message.type === "file") {
      return (
        <div className="flex items-center gap-3 p-3 bg-white dark:bg-gray-900 rounded-lg border border-gray-300 dark:border-gray-600 max-w-sm">
          <div className="text-4xl">{getFileIcon(message.file_name)}</div>
          <div className="flex-1 min-w-0">
            <div className="font-medium text-sm truncate">
              {message.file_name || "Unknown file"}
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400">
              {formatFileSize(message.file_size)}
            </div>
          </div>
          <button
            onClick={handleDownload}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full transition-colors"
            title="Download file"
          >
            <Download className="w-5 h-5" />
          </button>
        </div>
      );
    }

    return null;
  };

  const groupedReactions = reactions.reduce((acc, reaction) => {
    acc[reaction.type] = reaction;
    return acc;
  }, {} as Record<string, Reaction>);

  if (message.type === "system") {
    return (
      <div className="flex justify-center mb-4">
        <div className="px-4 py-1 text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 rounded-full">
          {message.created_by_user && (
            <span className="font-semibold">
              {message.created_by_user.name}
            </span>
          )}{" "}
          {message.content}
        </div>
      </div>
    );
  }

  return (
    <div
      className={`flex ${
        isOwnMessage ? "justify-end" : "justify-start"
      } mb-4 group`}
    >
      <div
        className={`max-w-[70%] ${
          isOwnMessage ? "items-end" : "items-start"
        } flex flex-col`}
      >
        {message.reply_to && (
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 px-3 py-1 bg-gray-100 dark:bg-gray-800 rounded">
            <Reply className="w-3 h-3 inline mr-1" />
            Replying to: {message.reply_to.content.substring(0, 50)}...
          </div>
        )}

        <div
          className={`px-4 py-2 rounded-lg ${
            isOwnMessage
              ? "bg-blue-500 text-white"
              : "bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          }`}
        >
          {message.created_by_user && !isOwnMessage && (
            <div className="text-xs font-semibold mb-1">
              {message.created_by_user.name}
            </div>
          )}

          {isEditing ? (
            <div>
              <input
                type="text"
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                className="bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-2 py-1 rounded"
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleEdit();
                  if (e.key === "Escape") setIsEditing(false);
                }}
              />
              <button onClick={handleEdit} className="ml-2 text-xs underline">
                Save
              </button>
            </div>
          ) : (
            <>
              {renderMedia()}
              {message.content && <div>{message.content}</div>}
            </>
          )}

          <div className="text-xs opacity-70 mt-1">
            {new Date(message.created_at).toLocaleTimeString()}
          </div>
        </div>

        {/* Reactions */}
        {Object.keys(groupedReactions).length > 0 && (
          <div className="flex gap-1 mt-1">
            {Object.entries(groupedReactions).map(([type, reaction]) => {
              const hasUserReacted =
                reaction.user_ids?.includes(userId) ?? false;
              return (
                <button
                  key={type}
                  className={`px-2 py-0.5 rounded-full text-xs flex items-center gap-1 hover:bg-gray-200 dark:hover:bg-gray-700 ${
                    hasUserReacted
                      ? "bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200"
                      : "bg-gray-100 dark:bg-gray-800"
                  }`}
                  onClick={() => onReact?.(message.id, type)}
                >
                  <span>
                    {REACTION_EMOJIS[type as keyof typeof REACTION_EMOJIS]}
                  </span>
                  <span>{reaction.count}</span>
                </button>
              );
            })}
          </div>
        )}

        {/* Action buttons */}
        <div className="opacity-0 group-hover:opacity-100 transition-opacity flex gap-2 mt-1">
          <button
            onClick={() => setShowReactions(!showReactions)}
            className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
          >
            <Smile className="w-4 h-4" />
          </button>
          <button
            onClick={() => onReply?.(message)}
            className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
          >
            <Reply className="w-4 h-4" />
          </button>
          {isOwnMessage && (
            <div className="relative">
              <button
                onClick={() => setShowMenu(!showMenu)}
                className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
              >
                <MoreVertical className="w-4 h-4" />
              </button>
              {showMenu && (
                <div className="absolute right-0 mt-1 bg-white dark:bg-gray-800 shadow-lg rounded-lg overflow-hidden z-10">
                  <button
                    onClick={() => {
                      setIsEditing(true);
                      setShowMenu(false);
                    }}
                    className="flex items-center gap-2 px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left"
                  >
                    <Edit className="w-4 h-4" />
                    Edit
                  </button>
                  <button
                    onClick={() => {
                      onDelete?.(message.id);
                      setShowMenu(false);
                    }}
                    className="flex items-center gap-2 px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left text-red-600"
                  >
                    <Trash className="w-4 h-4" />
                    Delete
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Reaction picker */}
        {showReactions && (
          <div className="flex gap-2 mt-2 p-2 bg-white dark:bg-gray-800 shadow-lg rounded-lg">
            {Object.entries(REACTION_EMOJIS).map(([type, emoji]) => (
              <button
                key={type}
                onClick={() => {
                  onReact?.(message.id, type);
                  setShowReactions(false);
                }}
                className="text-2xl hover:scale-125 transition-transform"
              >
                {emoji}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
