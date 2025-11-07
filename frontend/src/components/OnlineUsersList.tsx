"use client";

import { Users, X } from "lucide-react";

interface OnlineUser {
  id: number;
  username: string;
  name: string;
}

interface OnlineUsersListProps {
  onlineUsers: OnlineUser[];
  onClose: () => void;
}

export function OnlineUsersList({
  onlineUsers,
  onClose,
}: OnlineUsersListProps) {
  console.log("OnlineUsersList rendering with onlineUsers:", onlineUsers);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5" />
            <h2 className="text-xl font-bold">Online Users</h2>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              ({onlineUsers.length})
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
          {onlineUsers.length === 0 ? (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              <Users className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p>No users online</p>
            </div>
          ) : (
            <div className="space-y-2">
              {onlineUsers.map((user) => (
                <OnlineUserItem key={user.id} user={user} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

interface OnlineUserItemProps {
  user: OnlineUser;
}

function OnlineUserItem({ user }: OnlineUserItemProps) {
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
      {/* Avatar */}
      <div className="relative">
        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-green-400 to-blue-500 flex items-center justify-center text-white font-semibold">
          {user.name.charAt(0).toUpperCase()}
        </div>
        {/* Online Status Indicator */}
        <div className="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white dark:border-gray-800 bg-green-500" />
      </div>

      {/* User Info */}
      <div className="flex-1 min-w-0">
        <div className="font-medium text-gray-900 dark:text-gray-100 truncate">
          {user.name}
        </div>
        <div className="text-sm text-gray-500 dark:text-gray-400 truncate">
          @{user.username}
        </div>
      </div>

      {/* Status Badge */}
      <div className="text-xs px-2 py-1 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">
        Online
      </div>
    </div>
  );
}
