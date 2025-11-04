"use client";

import { User } from "@/types";
import { Users, X } from "lucide-react";

export interface Friend extends User {
  is_online: boolean;
}

interface FriendsListProps {
  friends: Friend[];
  onClose: () => void;
}

export function FriendsList({ friends, onClose }: FriendsListProps) {
  console.log("FriendsList rendering with friends:", friends);
  const onlineFriends = friends.filter((f) => f.is_online);
  const offlineFriends = friends.filter((f) => !f.is_online);
  console.log(
    "Online friends:",
    onlineFriends,
    "Offline friends:",
    offlineFriends
  );

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5" />
            <h2 className="text-xl font-bold">Friends</h2>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              ({friends.length})
            </span>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Friends List */}
        <div className="flex-1 overflow-y-auto p-4">
          {friends.length === 0 ? (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              <Users className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p>No friends yet</p>
              <p className="text-sm mt-1">Add friends to see them here</p>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Online Friends */}
              {onlineFriends.length > 0 && (
                <div>
                  <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                    Online — {onlineFriends.length}
                  </h3>
                  <div className="space-y-2">
                    {onlineFriends.map((friend) => (
                      <FriendItem key={friend.id} friend={friend} />
                    ))}
                  </div>
                </div>
              )}

              {/* Offline Friends */}
              {offlineFriends.length > 0 && (
                <div>
                  <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                    Offline — {offlineFriends.length}
                  </h3>
                  <div className="space-y-2">
                    {offlineFriends.map((friend) => (
                      <FriendItem key={friend.id} friend={friend} />
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

interface FriendItemProps {
  friend: Friend;
}

function FriendItem({ friend }: FriendItemProps) {
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
      {/* Avatar */}
      <div className="relative">
        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white font-semibold">
          {friend.name.charAt(0).toUpperCase()}
        </div>
        {/* Online Status Indicator */}
        <div
          className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white dark:border-gray-800 ${
            friend.is_online ? "bg-green-500" : "bg-gray-400"
          }`}
        />
      </div>

      {/* Friend Info */}
      <div className="flex-1 min-w-0">
        <div className="font-medium text-gray-900 dark:text-gray-100 truncate">
          {friend.name}
        </div>
        <div className="text-sm text-gray-500 dark:text-gray-400 truncate">
          @{friend.username}
        </div>
      </div>

      {/* Status Badge */}
      <div
        className={`text-xs px-2 py-1 rounded-full ${
          friend.is_online
            ? "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400"
            : "bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400"
        }`}
      >
        {friend.is_online ? "Online" : "Offline"}
      </div>
    </div>
  );
}
