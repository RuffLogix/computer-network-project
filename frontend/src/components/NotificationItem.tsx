"use client";

import { Notification } from "@/types";
import { Bell, Check, X } from "lucide-react";
import { useState } from "react";

interface NotificationItemProps {
  notification: Notification;
  onAccept?: (id: number) => void;
  onReject?: (id: number) => void;
  onMarkRead?: (id: number) => void;
}

export function NotificationItem({
  notification,
  onAccept,
  onReject,
  onMarkRead,
}: NotificationItemProps) {
  const [isHovered, setIsHovered] = useState(false);

  const showActions =
    notification.type === "friend_request" ||
    notification.type === "group_invitation";

  return (
    <div
      className={`p-4 border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors ${
        notification.status === "unread" ? "bg-blue-50 dark:bg-blue-900/20" : ""
      }`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          <Bell className="w-5 h-5 text-blue-500" />
        </div>

        <div className="flex-1">
          <div className="flex items-start justify-between">
            <div>
              <h4 className="font-semibold text-sm">{notification.title}</h4>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                {notification.message}
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">
                {new Date(notification.created_at).toLocaleString()}
              </p>
            </div>

            {isHovered && notification.status === "unread" && (
              <button
                onClick={() => onMarkRead && onMarkRead(notification.id)}
                className="text-xs text-blue-500 hover:underline"
              >
                Mark as read
              </button>
            )}
          </div>

          {showActions && notification.status === "unread" && (
            <div className="flex gap-2 mt-3">
              <button
                onClick={() => onAccept && onAccept(notification.id)}
                className="flex items-center gap-1 px-3 py-1 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors text-sm"
              >
                <Check className="w-4 h-4" />
                Accept
              </button>
              <button
                onClick={() => onReject && onReject(notification.id)}
                className="flex items-center gap-1 px-3 py-1 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors text-sm"
              >
                <X className="w-4 h-4" />
                Reject
              </button>
            </div>
          )}

          {notification.status === "accepted" && (
            <div className="text-xs text-green-600 dark:text-green-400 mt-2">
              ✓ Accepted
            </div>
          )}

          {notification.status === "rejected" && (
            <div className="text-xs text-red-600 dark:text-red-400 mt-2">
              ✗ Rejected
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
