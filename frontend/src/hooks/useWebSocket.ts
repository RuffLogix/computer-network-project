"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { Event, Message, Notification, Reaction } from "@/types";
import { WS_BASE_URL, EVENT_TYPES, API_BASE_URL } from "@/constants";

export interface WebSocketHookReturn {
  isConnected: boolean;
  loadChatHistory: (chatId: number) => Promise<void>;
  refreshNotifications: () => Promise<void>;
  sendMessage: (
    chatId: number,
    content: string,
    type?: string,
    mediaUrl?: string,
    replyToId?: number
  ) => void;
  editMessage: (messageId: number, content: string, chatId: number) => void;
  deleteMessage: (messageId: number, chatId: number) => void;
  addReaction: (messageId: number, type: string, chatId: number) => void;
  joinChat: (chatId: number) => void;
  leaveChat: (chatId: number) => void;
  sendTyping: (chatId: number, isTyping: boolean) => void;
  messages: Message[];
  notifications: Notification[];
  reactions: Map<number, Reaction[]>;
  typingUsers: Map<number, Set<number>>;
  onlineUsers: Set<number>;
  allOnlineUsers: number[];
}

export function useWebSocket(userId: number): WebSocketHookReturn {
  const ws = useRef<WebSocket | null>(null);
  const pendingEvents = useRef<Event[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [reactions, setReactions] = useState<Map<number, Reaction[]>>(
    new Map()
  );
  const [typingUsers, setTypingUsers] = useState<Map<number, Set<number>>>(
    new Map()
  );
  const [onlineUsers, setOnlineUsers] = useState<Set<number>>(new Set());
  const [allOnlineUsers, setAllOnlineUsers] = useState<number[]>([]);

  const fetchWithUser = useCallback(
    (path: string, options: RequestInit = {}) => {
      const headers = new Headers(options.headers || {});

      // Add authentication token
      const token =
        typeof window !== "undefined"
          ? localStorage.getItem("auth_token")
          : null;
      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }

      if (userId) {
        headers.set("X-User-ID", String(userId));
      }

      if (
        options.body &&
        !(options.body instanceof FormData) &&
        !headers.has("Content-Type")
      ) {
        headers.set("Content-Type", "application/json");
      }

      const requestInit: RequestInit = {
        ...options,
        headers,
      };

      const url = path.startsWith("http") ? path : `${API_BASE_URL}${path}`;
      return fetch(url, requestInit);
    },
    [userId]
  );

  const flushPendingEvents = useCallback(() => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      return;
    }

    while (pendingEvents.current.length > 0) {
      const nextEvent = pendingEvents.current.shift();
      if (nextEvent) {
        ws.current.send(JSON.stringify(nextEvent));
      }
    }
  }, []);

  const handleEvent = useCallback((event: Event) => {
    switch (event.type) {
      case EVENT_TYPES.SEND_MESSAGE:
        if (event.data.message) {
          const newMessage = event.data.message as Message;
          setMessages((prev) => {
            const exists = prev.some((msg) => msg.id === newMessage.id);
            if (exists) {
              return prev;
            }
            return [...prev, newMessage];
          });
        }
        break;

      case EVENT_TYPES.EDIT_MESSAGE:
        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === (event.data.message_id as number)
              ? { ...msg, content: event.data.content as string }
              : msg
          )
        );
        break;

      case EVENT_TYPES.DELETE_MESSAGE:
        setMessages((prev) =>
          prev.filter((msg) => msg.id !== (event.data.message_id as number))
        );
        break;

      case EVENT_TYPES.ADD_REACTION:
        if (event.data.reaction) {
          const reaction = event.data.reaction as Reaction;
          setReactions((prev) => {
            const newMap = new Map(prev);
            const msgReactions = newMap.get(reaction.message_id) || [];
            // Find existing reaction of same type
            const existingIndex = msgReactions.findIndex(
              (r) => r.type === reaction.type
            );
            if (existingIndex >= 0) {
              msgReactions[existingIndex] = reaction;
            } else {
              msgReactions.push(reaction);
            }
            newMap.set(reaction.message_id, [...msgReactions]);
            return newMap;
          });
        }
        break;

      case EVENT_TYPES.REMOVE_REACTION:
        if (event.data.reaction) {
          const reaction = event.data.reaction as Reaction;
          setReactions((prev) => {
            const newMap = new Map(prev);
            const msgReactions = newMap.get(reaction.message_id) || [];
            // Find and update the reaction of same type
            const existingIndex = msgReactions.findIndex(
              (r) => r.type === reaction.type
            );
            if (existingIndex >= 0) {
              msgReactions[existingIndex] = reaction;
              newMap.set(reaction.message_id, [...msgReactions]);
            }
            return newMap;
          });
        }
        break;

      case EVENT_TYPES.TYPING: {
        const chatId = event.data.chat_id as number;
        const typingUserId = event.data.user_id as number;
        const isTyping = event.data.is_typing as boolean;

        setTypingUsers((prev) => {
          const newMap = new Map(prev);
          const chatTyping = newMap.get(chatId) || new Set<number>();

          if (isTyping) {
            chatTyping.add(typingUserId);
          } else {
            chatTyping.delete(typingUserId);
          }

          newMap.set(chatId, chatTyping);
          return newMap;
        });
        break;
      }

      case EVENT_TYPES.NOTIFICATION:
        if (event.data.notification) {
          const incoming = event.data.notification as Notification;
          setNotifications((prev) => {
            const currentNotifications = prev || [];
            const copy = [...currentNotifications];
            const existingIndex = copy.findIndex((n) => n.id === incoming.id);
            if (existingIndex !== -1) {
              copy[existingIndex] = incoming;
              return copy;
            }
            return [...copy, incoming];
          });
        }
        break;

      case EVENT_TYPES.USER_ONLINE:
        if (event.data.user_id) {
          const onlineUserId = event.data.user_id as number;
          setOnlineUsers((prev) => {
            const newSet = new Set(prev);
            newSet.add(onlineUserId);
            return newSet;
          });
        }
        break;

      case EVENT_TYPES.USER_OFFLINE:
        if (event.data.user_id) {
          const offlineUserId = event.data.user_id as number;
          setOnlineUsers((prev) => {
            const newSet = new Set(prev);
            newSet.delete(offlineUserId);
            return newSet;
          });
        }
        break;

      case EVENT_TYPES.ONLINE_USERS_LIST:
        if (event.data.online_users) {
          const onlineUsersList = event.data.online_users as number[];
          setAllOnlineUsers(onlineUsersList);
        }
        break;
    }
  }, []);

  useEffect(() => {
    if (!userId) {
      console.log("WebSocket: No userId, not connecting");
      if (ws.current) {
        ws.current.close();
        ws.current = null;
      }
      pendingEvents.current = [];
      setIsConnected(false);
      return;
    }

    console.log("WebSocket: Initializing connection for userId:", userId);
    let reconnectTimeout: NodeJS.Timeout | null = null;
    let shouldReconnect = true;

    const connect = () => {
      // Note: WebSocket doesn't support custom headers in browser
      // Token needs to be passed via query parameter or sent after connection
      const token =
        typeof window !== "undefined"
          ? localStorage.getItem("auth_token")
          : null;
      const wsUrl = token
        ? `${WS_BASE_URL}/ws?token=${encodeURIComponent(token)}`
        : `${WS_BASE_URL}/ws`;

      console.log("WebSocket: Connecting to", wsUrl);
      ws.current = new WebSocket(wsUrl);

      ws.current.onopen = () => {
        console.log("WebSocket connected for user:", userId);
        setIsConnected(true);

        if (ws.current) {
          const connectEvent: Event = {
            type: EVENT_TYPES.CONNECT,
            created_by: userId,
            data: { userId },
          };
          console.log("Sending connect event:", connectEvent);
          ws.current.send(JSON.stringify(connectEvent));
        }

        flushPendingEvents();
      };

      ws.current.onclose = () => {
        console.log("WebSocket disconnected");
        setIsConnected(false);
        // Reconnect after 3 seconds if we should still be connected
        if (shouldReconnect && userId) {
          console.log("WebSocket: Will reconnect in 3 seconds");
          reconnectTimeout = setTimeout(connect, 3000);
        }
      };

      ws.current.onerror = (error) => {
        console.error("WebSocket error:", error);
        setIsConnected(false);
      };

      ws.current.onmessage = (event) => {
        try {
          const data: Event = JSON.parse(event.data);
          handleEvent(data);
        } catch (error) {
          console.error("Error parsing WebSocket message:", error);
        }
      };
    };

    connect();

    return () => {
      console.log("WebSocket: Cleaning up connection for userId:", userId);
      shouldReconnect = false;
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
      if (ws.current) {
        ws.current.close();
        ws.current = null;
      }
      setIsConnected(false);
    };
  }, [flushPendingEvents, handleEvent, userId]);

  const loadChatHistory = useCallback(
    async (chatId: number) => {
      if (!chatId) return;

      try {
        const response = await fetchWithUser(`/api/chats/${chatId}/messages`);
        if (!response.ok) {
          throw new Error(`Failed to load messages for chat ${chatId}`);
        }

        const data: Message[] = (await response.json()) || [];
        setMessages(data);

        const reactionMap = new Map<number, Reaction[]>();
        data.forEach((message) => {
          if (Array.isArray(message.reactions) && message.reactions.length) {
            reactionMap.set(message.id, message.reactions);
          }
        });
        setReactions(reactionMap);
      } catch (error) {
        console.error("Failed to load chat history:", error);
      }
    },
    [fetchWithUser]
  );

  const refreshNotifications = useCallback(async () => {
    if (!userId) return;

    try {
      const response = await fetchWithUser(`/api/notifications`);
      if (!response.ok) {
        throw new Error("Failed to load notifications");
      }

      const data: Notification[] = (await response.json()) || [];
      setNotifications(data);
    } catch (error) {
      console.error("Failed to refresh notifications:", error);
    }
  }, [fetchWithUser, userId]);

  useEffect(() => {
    refreshNotifications();
  }, [refreshNotifications]);

  const sendEvent = useCallback((event: Event) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(event));
      return;
    }

    pendingEvents.current.push(event);
    console.warn(
      "WebSocket not ready. Queuing event until connection opens:",
      event.type
    );
  }, []);

  const sendMessage = useCallback(
    (
      chatId: number,
      content: string,
      type: string = "text",
      mediaUrl?: string,
      replyToId?: number
    ) => {
      sendEvent({
        type: EVENT_TYPES.SEND_MESSAGE,
        data: {
          chat_id: chatId,
          content,
          type,
          media_url: mediaUrl,
          reply_to_id: replyToId,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  const editMessage = useCallback(
    (messageId: number, content: string, chatId: number) => {
      sendEvent({
        type: EVENT_TYPES.EDIT_MESSAGE,
        data: {
          message_id: messageId,
          content,
          chat_id: chatId,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  const deleteMessage = useCallback(
    (messageId: number, chatId: number) => {
      sendEvent({
        type: EVENT_TYPES.DELETE_MESSAGE,
        data: {
          message_id: messageId,
          chat_id: chatId,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  const addReaction = useCallback(
    (messageId: number, type: string, chatId: number) => {
      setReactions((prev) => {
        const newMap = new Map(prev);
        const msgReactions = newMap.get(messageId) || [];
        const existingIndex = msgReactions.findIndex((r) => r.type === type);

        if (existingIndex >= 0) {
          const existing = msgReactions[existingIndex];
          const hasUserReacted = existing.user_ids.includes(userId);

          if (hasUserReacted) {
            const newUserIds = existing.user_ids.filter((id) => id !== userId);
            if (newUserIds.length === 0) {
              msgReactions.splice(existingIndex, 1);
            } else {
              msgReactions[existingIndex] = {
                ...existing,
                count: existing.count - 1,
                user_ids: newUserIds,
              };
            }
          } else {
            msgReactions[existingIndex] = {
              ...existing,
              count: existing.count + 1,
              user_ids: [...existing.user_ids, userId],
            };
          }
        } else {
          msgReactions.push({
            id: Date.now(), // Temporary ID
            message_id: messageId,
            type: type as any,
            count: 1,
            user_ids: [userId],
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          });
        }

        newMap.set(messageId, [...msgReactions]);
        return newMap;
      });

      sendEvent({
        type: EVENT_TYPES.ADD_REACTION,
        data: {
          message_id: messageId,
          type,
          chat_id: chatId,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  const joinChat = useCallback(
    (chatId: number) => {
      console.log(`Joining chat ${chatId} for user ${userId}`);
      sendEvent({
        type: EVENT_TYPES.JOIN,
        data: {
          chat_id: chatId,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  const leaveChat = useCallback(
    (chatId: number) => {
      sendEvent({
        type: EVENT_TYPES.LEAVE,
        data: {
          chat_id: chatId,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  const sendTyping = useCallback(
    (chatId: number, isTyping: boolean) => {
      sendEvent({
        type: EVENT_TYPES.TYPING,
        data: {
          chat_id: chatId,
          is_typing: isTyping,
        },
        created_by: userId,
      });
    },
    [sendEvent, userId]
  );

  return {
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
  };
}
