export type ChatType = "individual" | "private_group" | "public_group";
export type MessageType =
  | "text"
  | "image"
  | "video"
  | "sticker"
  | "file"
  | "system";
export type ReactionType = "like" | "love" | "laugh" | "wow" | "sad" | "angry";
export type FriendshipStatus = "pending" | "accepted" | "rejected" | "blocked";
export type NotificationType =
  | "friend_request"
  | "friend_accepted"
  | "group_invitation"
  | "message_reaction"
  | "message_reply"
  | "group_member_joined";
export type NotificationStatus = "unread" | "read" | "accepted" | "rejected";

export interface User {
  id: string; // MongoDB ObjectID as string
  numeric_id: number; // Numeric ID for chat operations
  username: string;
  name: string;
  email?: string; // Optional for guest users
  avatar?: string;
  is_guest: boolean; // New field to identify guest users
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

export interface Chat {
  id: number;
  type: ChatType;
  name: string;
  description?: string;
  is_public: boolean;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: number;
  chat_id: number;
  content: string;
  type: MessageType;
  media_url?: string;
  file_name?: string;
  file_size?: number;
  reply_to_id?: number;
  reply_to?: Message;
  reactions?: Reaction[];
  created_at: string;
  updated_at: string;
  created_by: number;
  created_by_user?: User;
}

export interface Reaction {
  id: number;
  message_id: number;
  type: ReactionType;
  count: number;
  user_ids: number[];
  created_at: string;
  updated_at: string;
}

export interface ChatMember {
  id: number;
  chat_id: number;
  user_id: number;
  role: "admin" | "member";
  joined_at: string;
}

export interface ChatInvitation {
  id: number;
  chat_id: number;
  code: string;
  expires_at?: string;
  max_uses?: number;
  used_count: number;
  created_by: number;
  created_at: string;
  is_active: boolean;
}

export interface FriendInvitation {
  id: number;
  code: string;
  user_id: number;
  expires_at?: string;
  max_uses?: number;
  used_count: number;
  created_at: string;
  is_active: boolean;
}

export interface Friendship {
  id: number;
  user_id: number;
  friend_id: number;
  status: FriendshipStatus;
  created_at: string;
  updated_at: string;
  user?: User;
  friend?: User;
}

export interface Notification {
  id: number;
  recipient_id: number;
  sender_id: number;
  type: NotificationType;
  status: NotificationStatus;
  title: string;
  message: string;
  reference_id?: number;
  created_at: string;
  updated_at: string;
  sender?: User;
}

export type EventType =
  | "connect"
  | "join"
  | "leave"
  | "send_message"
  | "delete_message"
  | "edit_message"
  | "add_reaction"
  | "remove_reaction"
  | "typing"
  | "notification"
  | "friend_invite"
  | "group_invite"
  | "user_online"
  | "user_offline"
  | "online_users_list";

export interface Event {
  type: EventType;
  data: Record<string, unknown>;
  created_by: number;
}

export interface Friend extends User {
  is_online: boolean;
}

export interface OnlineUser {
  id: number;
  username: string;
  name: string;
}
