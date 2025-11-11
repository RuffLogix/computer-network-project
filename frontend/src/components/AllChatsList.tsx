"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { API_BASE_URL } from "@/constants";
import { AuthService } from "@/lib/auth";
import { Hash, Users, X, Lock, Globe } from "lucide-react";
import { Chat } from "@/types";
import { JSX } from "react";

interface ChatMember {
  id: number;
  chat_id: number;
  user_id: number;
  role: string;
  joined_at: string;
  user?: {
    id: string;
    numeric_id: number;
    username: string;
    name: string;
  };
  is_online: boolean;
}

interface ChatWithMembers extends Chat {
  members: ChatMember[];
}

interface AllChatsListProps {
  onClose: () => void;
  onChatSelect?: (chatId: number) => void;
}

export function AllChatsList({ onClose, onChatSelect }: AllChatsListProps) {
  const t = useTranslations();
  const [chats, setChats] = useState<ChatWithMembers[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchAllChats = async () => {
      try {
        const token = AuthService.getToken();
        const response = await fetch(`${API_BASE_URL}/api/all-chats`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error("Failed to fetch chats");
        }

        const data = await response.json();
        setChats(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setLoading(false);
      }
    };

    fetchAllChats();
  }, []);

  const getChatIcon = (chatType: string, isPublic: boolean) => {
    if (chatType === "individual") {
      return <Users className="w-4 h-4" />;
    }
    if (isPublic) {
      return <Globe className="w-4 h-4" />;
    }
    return <Lock className="w-4 h-4" />;
  };

  const getChatTypeLabel = (chatType: string, isPublic: boolean) => {
    if (chatType === "individual") return "Individual";
    if (isPublic) return "Public Group";
    return "Private Group";
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            <Hash className="w-5 h-5" />
            <h2 className="text-xl font-bold">{t("chat.allChats")}</h2>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              ({chats.length})
            </span>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4">
          {loading ? (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              Loading chat groups...
            </div>
          ) : error ? (
            <div className="text-center py-8 text-red-500 dark:text-red-400">
              Error: {error}
            </div>
          ) : chats.length === 0 ? (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              <Hash className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p>No chat groups found</p>
            </div>
          ) : (
            <div className="space-y-4">
              {chats.map((chat) => (
                <ChatGroupItem
                  key={chat.id}
                  chat={chat}
                  onSelect={onChatSelect}
                  getChatIcon={getChatIcon}
                  getChatTypeLabel={getChatTypeLabel}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

interface ChatGroupItemProps {
  chat: ChatWithMembers;
  onSelect?: (chatId: number) => void;
  getChatIcon: (chatType: string, isPublic: boolean) => JSX.Element;
  getChatTypeLabel: (chatType: string, isPublic: boolean) => string;
}

function ChatGroupItem({
  chat,
  onSelect,
  getChatIcon,
  getChatTypeLabel,
}: ChatGroupItemProps) {
  const onlineMembers = chat.members.filter((m) => m.is_online);
  const offlineMembers = chat.members.filter((m) => !m.is_online);

  return (
    <div
      className={`border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors cursor-pointer ${
        onSelect ? "cursor-pointer" : ""
      }`}
      onClick={() => onSelect?.(chat.id)}
    >
      {/* Chat Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div className="text-gray-500 dark:text-gray-400">
            {getChatIcon(chat.type, chat.is_public)}
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-gray-900 dark:text-gray-100 truncate">
              {chat.name}
            </h3>
            <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
              {chat.description || "No description"}
            </p>
          </div>
        </div>
        <div className="text-xs px-2 py-1 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 ml-2">
          {getChatTypeLabel(chat.type, chat.is_public)}
        </div>
      </div>

      {/* Members */}
      <div className="space-y-2">
        <div className="text-sm font-medium text-gray-700 dark:text-gray-300">
          Members ({chat.members.length})
        </div>

        {/* Online Members */}
        {onlineMembers.length > 0 && (
          <div>
            <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
              Online ({onlineMembers.length})
            </div>
            <div className="flex flex-wrap gap-1">
              {onlineMembers.slice(0, 5).map((member) => (
                <div
                  key={member.user_id}
                  className="flex items-center gap-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 px-2 py-1 rounded-full text-xs"
                >
                  <div className="w-2 h-2 rounded-full bg-green-500" />
                  {member.user?.name || `User ${member.user_id}`}
                </div>
              ))}
              {onlineMembers.length > 5 && (
                <div className="text-xs text-gray-500 dark:text-gray-400 px-2 py-1">
                  +{onlineMembers.length - 5} more
                </div>
              )}
            </div>
          </div>
        )}

        {/* Offline Members */}
        {offlineMembers.length > 0 && (
          <div>
            <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">
              Offline ({offlineMembers.length})
            </div>
            <div className="flex flex-wrap gap-1">
              {offlineMembers.slice(0, 5).map((member) => (
                <div
                  key={member.user_id}
                  className="flex items-center gap-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-2 py-1 rounded-full text-xs"
                >
                  <div className="w-2 h-2 rounded-full bg-gray-400" />
                  {member.user?.name || `User ${member.user_id}`}
                </div>
              ))}
              {offlineMembers.length > 5 && (
                <div className="text-xs text-gray-500 dark:text-gray-400 px-2 py-1">
                  +{offlineMembers.length - 5} more
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
