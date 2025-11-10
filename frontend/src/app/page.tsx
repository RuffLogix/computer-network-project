"use client";

import { useCallback, useEffect, useMemo, useState, useRef } from "react";
import { useRouter } from "next/navigation";
import { useWebSocket } from "@/hooks/useWebSocket";
import { MessageItem } from "@/components/MessageItem";
import { ChatInput } from "@/components/ChatInput";
import { NotificationItem } from "@/components/NotificationItem";
import { InvitationLink } from "@/components/InvitationLink";
import { ThemeToggle } from "@/components/ThemeToggle";
import { CreateChatModal } from "@/components/CreateChatModal";
import { FriendsList, Friend } from "@/components/FriendsList";
import { OnlineUsersList } from "@/components/OnlineUsersList";
import { AllChatsList } from "@/components/AllChatsList";
import { Chat, Message, User } from "@/types";
import { API_BASE_URL } from "@/constants";
import { Bell, Users, Plus, Hash, LogOut, UserPlus, Globe } from "lucide-react";
import { AuthService } from "@/lib/auth";

type GeneratedInvite = {
  type: "friend" | "chat";
  code: string;
  expires_at?: string;
  max_uses?: number;
  used_count?: number;
  chat_id?: number;
};

const DEFAULT_INVITE_EXPIRY_SECONDS = 7 * 24 * 60 * 60;

export default function Home() {
  const router = useRouter();
  const [userId, setUserId] = useState(0);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isAuthReady, setIsAuthReady] = useState(false);
  const [showCreateChatModal, setShowCreateChatModal] = useState(false);
  const joinedRooms = useRef<Set<number>>(new Set());
  const messagesEndRef = useRef<HTMLDivElement>(null);
  // Check authentication on mount
  useEffect(() => {
    if (!AuthService.isAuthenticated()) {
      router.push("/login");
      return;
    }
    const user = AuthService.getUser();
    setCurrentUser(user);
    // Use numeric_id for chat operations
    setUserId(user?.numeric_id || 0);
    setIsAuthReady(true);
  }, [router]);
  const [chats, setChats] = useState<Chat[]>([]);
  const [publicChats, setPublicChats] = useState<Chat[]>([]);
  const [selectedChatId, setSelectedChatId] = useState<number | null>(null);
  const [replyTo, setReplyTo] = useState<Message | null>(null);
  const [showNotifications, setShowNotifications] = useState(false);
  const [showAddFriendModal, setShowAddFriendModal] = useState(false);
  const [showGroupInviteModal, setShowGroupInviteModal] = useState(false);
  const [showFriendsList, setShowFriendsList] = useState(false);
  const [showOnlineUsersList, setShowOnlineUsersList] = useState(false);
  const [showAllChatsList, setShowAllChatsList] = useState(false);
  const [friends, setFriends] = useState<Friend[]>([]);
  const [loadingChats, setLoadingChats] = useState(false);
  const [onlineUsersDetails, setOnlineUsersDetails] = useState<User[]>([]);
  const [friendAddMethod, setFriendAddMethod] = useState<"direct" | "link">(
    "direct"
  );
  const [directAddIdentifier, setDirectAddIdentifier] = useState("");
  const [inviteChatId, setInviteChatId] = useState<number | null>(null);
  const [inviteIsLoading, setInviteIsLoading] = useState(false);
  const [generatedInvite, setGeneratedInvite] =
    useState<GeneratedInvite | null>(null);

  const fetchWithUser = useCallback(
    (path: string, init: RequestInit = {}) => {
      const headers = new Headers(init.headers || {});

      // Add authentication token
      const token = AuthService.getToken();
      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }
      headers.set("X-User-ID", String(userId));

      if (
        init.body &&
        !(init.body instanceof FormData) &&
        !headers.has("Content-Type")
      ) {
        headers.set("Content-Type", "application/json");
      }

      const url = path.startsWith("http") ? path : `${API_BASE_URL}${path}`;
      return fetch(url, {
        ...init,
        headers,
      });
    },
    [userId]
  );

  const {
    isConnected,
    loadChatHistory,
    refreshNotifications,
    sendMessage,
    editMessage,
    deleteMessage,
    addReaction,
    joinChat,
    leaveChat,
    sendTyping,
    messages,
    notifications,
    reactions,
    typingUsers,
    onlineUsers,
    allOnlineUsers,
  } = useWebSocket(userId);

  const scrollToBottom = useCallback(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, []);

  // Debug log for connection state
  // useEffect(() => {
  //   console.log("Page: WebSocket connection state changed:", isConnected);
  // }, [isConnected]);

  const fetchChats = useCallback(async () => {
    setLoadingChats(true);
    try {
      const response = await fetchWithUser("/api/chats");
      if (!response.ok) {
        throw new Error("Failed to load chats");
      }

      const data: Chat[] = (await response.json()) || [];
      setChats(data);
    } catch (error) {
      console.error("Failed to load chats:", error);
    } finally {
      setLoadingChats(false);
    }
  }, [fetchWithUser]);

  const fetchPublicChats = useCallback(async () => {
    try {
      const url = `${API_BASE_URL}/api/chats/public`;
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Failed to load public chats");
      }

      const data: Chat[] = (await response.json()) || [];
      setPublicChats(data);
    } catch (error) {
      console.error("Failed to load public chats:", error);
    }
  }, []);

  const fetchFriends = useCallback(async () => {
    try {
      const response = await fetchWithUser("/api/friends");
      if (!response.ok) {
        if (response.status === 401) {
          return;
        }
        throw new Error("Failed to load friends");
      }

      const data = (await response.json()) || [];
      console.log("Fetched friends:", data);
      setFriends(data);
    } catch (error) {
      console.error("Failed to load friends:", error);
      setFriends([]);
    }
  }, [fetchWithUser]);

  useEffect(() => {
    if (!isAuthReady) return;

    fetchPublicChats();

    if (userId !== 0) {
      fetchChats();
      fetchFriends();
    }
  }, [fetchChats, fetchPublicChats, fetchFriends, isAuthReady, userId]);

  useEffect(() => {
    // console.log("Online users updated:", Array.from(onlineUsers));
    setFriends((prevFriends) => {
      const updated = prevFriends.map((friend) => ({
        ...friend,
        is_online: onlineUsers.has(friend.numeric_id),
      }));
      // console.log("Updated friends with online status:", updated);
      return updated;
    });
  }, [onlineUsers]);

  // Fetch online user details when allOnlineUsers changes
  useEffect(() => {
    const fetchOnlineUserDetails = async () => {
      if (allOnlineUsers.length === 0) {
        setOnlineUsersDetails([]);
        return;
      }

      try {
        const userDetailsPromises = allOnlineUsers.map(async (userId) => {
          const response = await fetchWithUser(`/api/users/${userId}`);
          if (response.ok) {
            return await response.json();
          }
          return null;
        });

        const userDetails = await Promise.all(userDetailsPromises);
        const validUsers = userDetails.filter((user) => user !== null);
        setOnlineUsersDetails(validUsers);
      } catch (error) {
        console.error("Failed to fetch online user details:", error);
        setOnlineUsersDetails([]);
      }
    };

    fetchOnlineUserDetails();
  }, [allOnlineUsers, fetchWithUser]);

  useEffect(() => {
    if (selectedChatId === null) return;
    loadChatHistory(selectedChatId);
  }, [selectedChatId, loadChatHistory]);

  useEffect(() => {
    if (!showAddFriendModal) {
      setGeneratedInvite(null);
      setInviteIsLoading(false);
      setFriendAddMethod("direct");
      setDirectAddIdentifier("");
    }
  }, [showAddFriendModal]);

  useEffect(() => {
    if (!showGroupInviteModal) {
      setGeneratedInvite(null);
      setInviteIsLoading(false);
      return;
    }

    const preferredChat = chats?.find(
      (chat) => chat.type === "private_group" && chat.id === selectedChatId
    );

    if (preferredChat) {
      setInviteChatId(preferredChat.id);
      return;
    }

    const firstPrivate = chats?.find((chat) => chat.type === "private_group");
    if (firstPrivate) {
      setInviteChatId(firstPrivate.id);
    }
  }, [showGroupInviteModal, chats, selectedChatId]);

  useEffect(() => {
    if (!isConnected || selectedChatId === null) return;

    const currentChatId = selectedChatId; // Capture current value

    if (!joinedRooms.current.has(currentChatId)) {
      joinChat(currentChatId);
      joinedRooms.current.add(currentChatId);
    }

    return () => {
      if (currentChatId !== null && joinedRooms.current.has(currentChatId)) {
        leaveChat(currentChatId);
        joinedRooms.current.delete(currentChatId);
      }
    };
  }, [isConnected, selectedChatId, joinChat, leaveChat]);

  const handleSendMessage = useCallback(
    (
      content: string,
      type?: string,
      mediaUrl?: string,
      replyToId?: number,
      fileName?: string,
      fileSize?: number
    ) => {
      if (!selectedChatId) return;
      sendMessage(
        selectedChatId,
        content,
        type,
        mediaUrl,
        replyToId,
        fileName,
        fileSize
      );
      setReplyTo(null);
      // Scroll to bottom when sending a message
      setTimeout(() => {
        scrollToBottom();
      }, 100);
    },
    [selectedChatId, sendMessage, scrollToBottom]
  );

  const handleJoinPublicChat = useCallback(
    async (chatId: number) => {
      try {
        const response = await fetchWithUser(`/api/chats/${chatId}/join`, {
          method: "POST",
        });
        if (!response.ok) {
          throw new Error("Unable to join chat");
        }
        // Refresh chat lists to show the newly joined chat
        await fetchChats();
        await fetchPublicChats();
        // Select the chat - the useEffect will handle joining via WebSocket
        setSelectedChatId(chatId);
      } catch (error) {
        console.error("Failed to join public chat:", error);
      }
    },
    [fetchWithUser, fetchChats, fetchPublicChats]
  );

  const handleAcceptNotification = useCallback(
    async (id: number) => {
      try {
        const response = await fetchWithUser(
          `/api/notifications/${id}/accept`,
          {
            method: "POST",
          }
        );
        if (!response.ok) {
          let errorMessage = "Failed to accept notification";
          try {
            const errorData = await response.json();
            if (errorData.error) {
              errorMessage = errorData.error;
            }
          } catch (_e) {
            // If we can't parse the error, use the default message
          }
          throw new Error(errorMessage);
        }
        await refreshNotifications();
        await fetchFriends(); // Refresh friends list after accepting notification
      } catch (error) {
        console.error("Failed to accept notification:", error);
        alert(
          error instanceof Error
            ? error.message
            : "Failed to accept notification"
        );
      }
    },
    [fetchWithUser, refreshNotifications, fetchFriends]
  );

  const handleRejectNotification = useCallback(
    async (id: number) => {
      try {
        const response = await fetchWithUser(
          `/api/notifications/${id}/reject`,
          {
            method: "POST",
          }
        );
        if (!response.ok) {
          throw new Error("Failed to reject notification");
        }
        await refreshNotifications();
      } catch (error) {
        console.error("Failed to reject notification:", error);
      }
    },
    [fetchWithUser, refreshNotifications]
  );

  const handleMarkNotificationRead = useCallback(
    async (id: number) => {
      try {
        const response = await fetchWithUser(`/api/notifications/${id}/read`, {
          method: "PUT",
        });
        if (!response.ok) {
          throw new Error("Failed to mark as read");
        }
        await refreshNotifications();
      } catch (error) {
        console.error("Failed to mark notification as read:", error);
      }
    },
    [fetchWithUser, refreshNotifications]
  );

  const handleLogout = useCallback(() => {
    AuthService.logout();
    router.push("/login");
  }, [router]);

  const handleLeaveGroup = useCallback(
    async (chatId: number) => {
      if (!confirm("Are you sure you want to leave this group?")) {
        return;
      }

      try {
        const response = await fetchWithUser(
          `/api/chats/${chatId}/members/${userId}`,
          {
            method: "DELETE",
          }
        );
        if (!response.ok) {
          throw new Error("Failed to leave group");
        }

        leaveChat(chatId);
        joinedRooms.current.delete(chatId);

        if (selectedChatId === chatId) {
          setSelectedChatId(null);
        }

        // Refresh chat lists
        await fetchChats();
        await fetchPublicChats();
      } catch (error) {
        console.error("Failed to leave group:", error);
        alert("Failed to leave group. Please try again.");
      }
    },
    [
      fetchWithUser,
      userId,
      leaveChat,
      selectedChatId,
      fetchChats,
      fetchPublicChats,
    ]
  );

  const handleCreateChat = useCallback(
    async (
      name: string,
      description: string,
      isPublic: boolean,
      type: "individual" | "private_group" | "public_group"
    ) => {
      try {
        const response = await fetchWithUser("/api/chats", {
          method: "POST",
          body: JSON.stringify({
            name,
            description,
            type,
            is_public: isPublic,
          }),
        });
        if (!response.ok) {
          throw new Error("Failed to create chat");
        }
        await fetchChats();
        await fetchPublicChats();
      } catch (error) {
        console.error("Failed to create chat:", error);
        throw error;
      }
    },
    [fetchWithUser, fetchChats, fetchPublicChats]
  );

  const handleAddFriend = useCallback(async () => {
    setInviteIsLoading(true);
    setGeneratedInvite(null);
    try {
      if (friendAddMethod === "direct") {
        if (!directAddIdentifier.trim()) {
          alert("Please enter a user ID or username");
          return;
        }

        const response = await fetchWithUser(
          "/api/invitations/friend/request",
          {
            method: "POST",
            body: JSON.stringify({
              target_identifier: directAddIdentifier.trim(),
            }),
          }
        );
        if (!response.ok) {
          const errorData = await response.json();
          const errorMessage =
            errorData.error || "Failed to send friend request";
          alert(errorMessage);
          return;
        }

        setShowAddFriendModal(false);
        setDirectAddIdentifier(""); // Clear the input
        alert("Friend request sent successfully!");
        // Refresh friends list
        fetchFriends();
      } else {
        const response = await fetchWithUser("/api/invitations/friend", {
          method: "POST",
          body: JSON.stringify({ expires_in: DEFAULT_INVITE_EXPIRY_SECONDS }),
        });
        if (!response.ok) {
          const errorData = await response.json();
          const errorMessage =
            errorData.error || "Failed to create friend invitation";
          alert(errorMessage);
          return;
        }

        const invitation = await response.json();
        setGeneratedInvite({
          type: "friend",
          code: invitation.code,
          expires_at: invitation.expires_at,
          max_uses: invitation.max_uses ?? undefined,
          used_count: invitation.used_count,
        });
      }
    } catch (error) {
      console.error("Failed to add friend:", error);
      alert(error instanceof Error ? error.message : "Failed to add friend");
    } finally {
      setInviteIsLoading(false);
    }
  }, [fetchWithUser, friendAddMethod, directAddIdentifier, fetchFriends]);

  const handleGenerateGroupInvite = useCallback(async () => {
    setInviteIsLoading(true);
    setGeneratedInvite(null);
    try {
      const targetChatId = inviteChatId ?? selectedChatId;
      if (!targetChatId) {
        throw new Error("Select a private group to generate an invitation");
      }

      const response = await fetchWithUser("/api/invitations/chat", {
        method: "POST",
        body: JSON.stringify({
          chat_id: targetChatId,
          expires_in: DEFAULT_INVITE_EXPIRY_SECONDS,
        }),
      });
      if (!response.ok) {
        throw new Error("Failed to create chat invitation");
      }

      const invitation = await response.json();
      setGeneratedInvite({
        type: "chat",
        code: invitation.code,
        expires_at: invitation.expires_at,
        max_uses: invitation.max_uses ?? undefined,
        used_count: invitation.used_count,
        chat_id: invitation.chat_id,
      });
    } catch (error) {
      console.error("Failed to generate group invitation:", error);
      alert(
        error instanceof Error ? error.message : "Failed to generate invitation"
      );
    } finally {
      setInviteIsLoading(false);
    }
  }, [fetchWithUser, inviteChatId, selectedChatId]);

  const privateChats = useMemo(
    () => chats?.filter((chat) => chat.type === "private_group") || [],
    [chats]
  );

  const chatMessages = useMemo(() => {
    if (selectedChatId === null) {
      return [];
    }
    return (
      messages?.filter((message) => message.chat_id === selectedChatId) || []
    );
  }, [messages, selectedChatId]);

  // Scroll to bottom when chat messages change
  useEffect(() => {
    if (chatMessages.length > 0) {
      // Small delay to ensure DOM is updated
      setTimeout(() => {
        scrollToBottom();
      }, 100);
    }
  }, [chatMessages, scrollToBottom]);

  const activeTypingCount = selectedChatId
    ? typingUsers.get(selectedChatId)?.size ?? 0
    : 0;

  const unreadCount =
    notifications?.filter((n) => n.status === "unread").length || 0;

  const isMemberOf = useCallback(
    (chatId: number) => chats?.some((chat) => chat.id === chatId) || false,
    [chats]
  );

  const handleReact = useCallback(
    (msgId: number, type: string) => {
      if (!selectedChatId) return;

      // Backend handles toggle logic now
      addReaction(msgId, type, selectedChatId);
    },
    [selectedChatId, addReaction]
  );

  if (!isAuthReady) {
    return (
      <div className="h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-600 dark:text-gray-400">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col">
      <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold">Chat App</h1>
            <div
              className={`flex items-center gap-2 ${
                isConnected ? "text-green-500" : "text-red-500"
              }`}
            >
              <div className="w-2 h-2 rounded-full bg-current" />
              <span className="text-sm">
                {isConnected ? "Connected" : "Disconnected"}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {currentUser && !currentUser.is_guest && (
              <div className="text-sm text-gray-600 dark:text-gray-400">
                {currentUser.username}
              </div>
            )}
            {currentUser && currentUser.is_guest && (
              <div className="text-sm text-amber-600 dark:text-amber-400">
                Guest User
              </div>
            )}

            {/* Only show create and invite buttons to authenticated users */}
            {currentUser && !currentUser.is_guest && (
              <>
                <button
                  onClick={() => setShowCreateChatModal(true)}
                  title="Create Group Chat"
                  className="p-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                >
                  <Plus className="w-5 h-5" />
                </button>

                <button
                  onClick={() => setShowAddFriendModal(true)}
                  title="Add Friend"
                  className="p-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                >
                  <UserPlus className="w-5 h-5" />
                </button>

                <button
                  onClick={() => {
                    console.log(
                      "Opening friends list, current friends:",
                      friends
                    );
                    setShowFriendsList(true);
                  }}
                  title="Friends List"
                  className="relative p-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 transition-colors"
                >
                  <Users className="w-5 h-5" />
                  {friends.filter((f) => f.is_online).length > 0 && (
                    <span className="absolute -top-1 -right-1 bg-green-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                      {friends.filter((f) => f.is_online).length}
                    </span>
                  )}
                </button>

                <button
                  onClick={() => setShowOnlineUsersList(true)}
                  title="Online Users"
                  className="relative p-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                >
                  <Globe className="w-5 h-5" />
                </button>

                <button
                  onClick={() => setShowAllChatsList(true)}
                  title="All Chat Groups"
                  className="relative p-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                >
                  <Hash className="w-5 h-5" />
                </button>
              </>
            )}

            <button
              onClick={() => setShowNotifications((prev) => !prev)}
              className="relative p-2 bg-gray-200 dark:bg-gray-700 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              <Bell className="w-5 h-5" />
              {unreadCount > 0 && (
                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                  {unreadCount}
                </span>
              )}
            </button>

            <button
              onClick={handleLogout}
              title="Logout"
              className="p-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
            >
              <LogOut className="w-5 h-5" />
            </button>

            <ThemeToggle />
          </div>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        <aside className="w-80 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
          <div className="p-4">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <Users className="w-5 h-5" />
              Your Chats
            </h2>

            {loadingChats ? (
              <div className="text-sm text-gray-500 dark:text-gray-400">
                Loading chats...
              </div>
            ) : chats.length === 0 ? (
              <div className="text-sm text-gray-500 dark:text-gray-400">
                {currentUser?.is_guest ? (
                  <>
                    No chats yet. Join a public group below to start chatting!
                  </>
                ) : (
                  <>
                    No chats yet. Join a public group below or create a new one.
                  </>
                )}
              </div>
            ) : (
              <div className="space-y-2">
                {chats.map((chat) => (
                  <button
                    key={chat.id}
                    onClick={() => setSelectedChatId(chat.id)}
                    className={`w-full text-left px-3 py-2 rounded-lg border transition-colors ${
                      selectedChatId === chat.id
                        ? "border-blue-500 bg-blue-50 dark:bg-blue-500/10"
                        : "border-transparent hover:border-gray-300 dark:hover:border-gray-600"
                    }`}
                  >
                    <div className="font-semibold truncate">{chat.name}</div>
                    <div className="text-xs text-gray-500 dark:text-gray-400 capitalize">
                      {chat.type.replace("_", " ")}
                    </div>
                  </button>
                ))}
              </div>
            )}

            <div className="mt-6">
              <h3 className="text-sm font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-2 flex items-center gap-2">
                <Hash className="w-4 h-4" /> Public Groups
              </h3>
              {publicChats.length === 0 ? (
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  No public groups available yet.
                </div>
              ) : (
                <div className="space-y-2">
                  {publicChats.map((chat) => {
                    const member = isMemberOf(chat.id);
                    return (
                      <div
                        key={chat.id}
                        className="border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 flex items-center justify-between gap-2"
                      >
                        <div>
                          <div className="font-medium truncate">
                            {chat.name}
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">
                            {member ? "Joined" : "Public group"}
                          </div>
                        </div>
                        <button
                          onClick={() =>
                            member
                              ? setSelectedChatId(chat.id)
                              : handleJoinPublicChat(chat.id)
                          }
                          className={`px-3 py-1 rounded-md text-sm transition-colors ${
                            member
                              ? "bg-gray-200 dark:bg-gray-700"
                              : "bg-blue-500 text-white hover:bg-blue-600"
                          }`}
                        >
                          {member ? "Open" : "Join"}
                        </button>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </aside>

        <main className="flex-1 flex flex-col bg-gray-50 dark:bg-gray-900">
          {selectedChatId !== null && (
            <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-3 flex items-center justify-between">
              <h3 className="font-semibold">
                {chats.find((c) => c.id === selectedChatId)?.name ||
                  publicChats.find((c) => c.id === selectedChatId)?.name ||
                  "Chat"}
              </h3>
              {(() => {
                const currentChat = chats.find((c) => c.id === selectedChatId);
                const isPrivateGroup = currentChat?.type === "private_group";
                const isPublicGroup = currentChat?.type === "public_group";
                const isGlobalChat = currentChat?.name === "Global Chat";

                if (isPrivateGroup) {
                  return (
                    <div className="flex gap-2">
                      <button
                        onClick={() => setShowGroupInviteModal(true)}
                        className="px-3 py-1 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                      >
                        Invite
                      </button>
                      <button
                        onClick={() => handleLeaveGroup(selectedChatId)}
                        className="px-3 py-1 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
                      >
                        Leave
                      </button>
                    </div>
                  );
                } else if (isPublicGroup && !isGlobalChat) {
                  return (
                    <button
                      onClick={() => handleLeaveGroup(selectedChatId)}
                      className="px-3 py-1 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
                    >
                      Leave
                    </button>
                  );
                }
                return null;
              })()}
            </div>
          )}
          <div className="flex-1 overflow-y-auto p-6 pb-4">
            {selectedChatId === null ? (
              <div className="h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
                Select a chat to start messaging.
              </div>
            ) : chatMessages.length === 0 ? (
              <div className="h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
                No messages yet. Start a conversation!
              </div>
            ) : (
              <div className="space-y-2 mb-4">
                {chatMessages.map((message) => {
                  const messageReactions =
                    reactions.get(message.id) ?? message.reactions ?? [];
                  return (
                    <MessageItem
                      key={message.id}
                      message={message}
                      reactions={messageReactions}
                      isOwnMessage={message.created_by === userId}
                      userId={userId}
                      onReply={setReplyTo}
                      onReact={handleReact}
                      onEdit={(msgId, content) =>
                        selectedChatId &&
                        editMessage(msgId, content, selectedChatId)
                      }
                      onDelete={(msgId) =>
                        selectedChatId && deleteMessage(msgId, selectedChatId)
                      }
                    />
                  );
                })}
              </div>
            )}

            {activeTypingCount > 0 && (
              <div className="text-sm text-gray-500 dark:text-gray-400 mt-2 mb-4">
                {activeTypingCount === 1
                  ? "Someone is typing..."
                  : `${activeTypingCount} people are typing...`}
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {selectedChatId !== null ? (
            <ChatInput
              userId={userId}
              onSendMessage={handleSendMessage}
              onTyping={(isTyping) =>
                selectedChatId && sendTyping(selectedChatId, isTyping)
              }
              replyTo={replyTo}
              onCancelReply={() => setReplyTo(null)}
            />
          ) : (
            <div className="p-4 bg-gray-100 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 text-center text-gray-500 dark:text-gray-400">
              Select a chat to start messaging
            </div>
          )}
        </main>

        {showNotifications && (
          <aside className="w-96 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 overflow-y-auto">
            <div className="p-4 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold">Notifications</h2>
            </div>
            {notifications.length === 0 ? (
              <div className="p-4 text-center text-gray-500 dark:text-gray-400">
                No notifications
              </div>
            ) : (
              <div>
                {notifications.map((notification) => (
                  <NotificationItem
                    key={notification.id}
                    notification={notification}
                    onAccept={handleAcceptNotification}
                    onReject={handleRejectNotification}
                    onMarkRead={handleMarkNotificationRead}
                  />
                ))}
              </div>
            )}
          </aside>
        )}
      </div>

      {/* Add Friend Modal */}
      {showAddFriendModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-bold">Add Friend</h2>
              <button
                onClick={() => setShowAddFriendModal(false)}
                className="text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
              >
                ✕
              </button>
            </div>

            <div className="flex gap-2 border-b border-gray-200 dark:border-gray-700">
              <button
                onClick={() => setFriendAddMethod("direct")}
                className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                  friendAddMethod === "direct"
                    ? "text-blue-500 border-b-2 border-blue-500"
                    : "text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
                }`}
              >
                Add Friend
              </button>
              <button
                onClick={() => setFriendAddMethod("link")}
                className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                  friendAddMethod === "link"
                    ? "text-blue-500 border-b-2 border-blue-500"
                    : "text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
                }`}
              >
                Create Link
              </button>
            </div>

            {friendAddMethod === "direct" ? (
              <div className="space-y-4">
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  You can add a friend by their User ID or Username
                </p>
                <input
                  type="text"
                  value={directAddIdentifier}
                  onChange={(e) => setDirectAddIdentifier(e.target.value)}
                  placeholder="Enter User ID or Username"
                  className="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  onKeyPress={(e) => {
                    if (e.key === "Enter" && directAddIdentifier.trim()) {
                      handleAddFriend();
                    }
                  }}
                />
                <button
                  onClick={handleAddFriend}
                  disabled={inviteIsLoading || !directAddIdentifier.trim()}
                  className="w-full px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 dark:disabled:bg-gray-700 disabled:cursor-not-allowed transition-colors"
                >
                  {inviteIsLoading ? "Sending..." : "Send Friend Request"}
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Share this link with people you want to add as friends
                </p>
                {!generatedInvite ? (
                  <button
                    onClick={handleAddFriend}
                    disabled={inviteIsLoading}
                    className="w-full px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 dark:disabled:bg-gray-700 transition-colors"
                  >
                    {inviteIsLoading ? "Generating..." : "Generate Invite Link"}
                  </button>
                ) : (
                  <InvitationLink
                    code={generatedInvite.code}
                    type={generatedInvite.type}
                    expiresAt={generatedInvite.expires_at}
                    maxUses={generatedInvite.max_uses}
                    usedCount={generatedInvite.used_count}
                  />
                )}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Group Invite Modal */}
      {showGroupInviteModal && selectedChatId && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-bold">Invite to Group</h2>
              <button
                onClick={() => setShowGroupInviteModal(false)}
                className="text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
              >
                ✕
              </button>
            </div>

            {privateChats.length === 0 ? (
              <div className="text-center py-4">
                <p className="text-gray-500 dark:text-gray-400">
                  No private groups available
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Select Group</label>
                  <select
                    value={inviteChatId ?? selectedChatId}
                    onChange={(e) => setInviteChatId(Number(e.target.value))}
                    className="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {privateChats.map((chat) => (
                      <option key={chat.id} value={chat.id}>
                        {chat.name}
                      </option>
                    ))}
                  </select>
                </div>

                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Share this link to invite people to the group
                </p>

                {!generatedInvite ? (
                  <button
                    onClick={handleGenerateGroupInvite}
                    disabled={inviteIsLoading}
                    className="w-full px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 dark:disabled:bg-gray-700 transition-colors"
                  >
                    {inviteIsLoading ? "Generating..." : "Generate Invite Link"}
                  </button>
                ) : (
                  <InvitationLink
                    code={generatedInvite.code}
                    type={generatedInvite.type}
                    expiresAt={generatedInvite.expires_at}
                    maxUses={generatedInvite.max_uses}
                    usedCount={generatedInvite.used_count}
                  />
                )}
              </div>
            )}
          </div>
        </div>
      )}

      {showCreateChatModal && (
        <CreateChatModal
          onClose={() => setShowCreateChatModal(false)}
          onCreate={handleCreateChat}
        />
      )}

      {showFriendsList && (
        <FriendsList
          friends={friends}
          onClose={() => setShowFriendsList(false)}
        />
      )}

      {showOnlineUsersList && (
        <OnlineUsersList
          onlineUsers={onlineUsersDetails.map((user) => ({
            id: user.numeric_id,
            username: user.username,
            name: user.name,
          }))}
          onClose={() => setShowOnlineUsersList(false)}
        />
      )}

      {showAllChatsList && (
        <AllChatsList
          onClose={() => setShowAllChatsList(false)}
          onChatSelect={(chatId) => {
            setSelectedChatId(chatId);
            setShowAllChatsList(false);
          }}
        />
      )}
    </div>
  );
}
