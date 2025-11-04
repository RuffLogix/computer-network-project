import { ReactionType } from "@/types";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
export const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";

export const REACTION_EMOJIS: Record<ReactionType, string> = {
  like: "👍",
  love: "❤️",
  laugh: "😂",
  wow: "😮",
  sad: "😢",
  angry: "😠",
};

export const MESSAGE_TYPES = {
  TEXT: "text",
  IMAGE: "image",
  VIDEO: "video",
} as const;

export const CHAT_TYPES = {
  INDIVIDUAL: "individual",
  PRIVATE_GROUP: "private_group",
  PUBLIC_GROUP: "public_group",
} as const;

export const EVENT_TYPES = {
  CONNECT: "connect",
  JOIN: "join",
  LEAVE: "leave",
  SEND_MESSAGE: "send_message",
  DELETE_MESSAGE: "delete_message",
  EDIT_MESSAGE: "edit_message",
  ADD_REACTION: "add_reaction",
  REMOVE_REACTION: "remove_reaction",
  TYPING: "typing",
  NOTIFICATION: "notification",
  FRIEND_INVITE: "friend_invite",
  GROUP_INVITE: "group_invite",
  USER_ONLINE: "user_online",
  USER_OFFLINE: "user_offline",
} as const;

export const NOTIFICATION_TYPES = {
  FRIEND_REQUEST: "friend_request",
  FRIEND_ACCEPTED: "friend_accepted",
  GROUP_INVITATION: "group_invitation",
  MESSAGE_REACTION: "message_reaction",
  MESSAGE_REPLY: "message_reply",
  GROUP_MEMBER_JOINED: "group_member_joined",
} as const;

export const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
export const ALLOWED_IMAGE_TYPES = [
  "image/jpeg",
  "image/png",
  "image/gif",
  "image/webp",
];
export const ALLOWED_VIDEO_TYPES = [
  "video/mp4",
  "video/webm",
  "video/ogg",
  "video/quicktime",
];
