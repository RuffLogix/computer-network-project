"use client";

import { useState } from "react";
import { Copy, Check, Link as LinkIcon } from "lucide-react";

interface InvitationLinkProps {
  code: string;
  type: "chat" | "friend";
  expiresAt?: string;
  maxUses?: number;
  usedCount?: number;
}

export function InvitationLink({
  code,
  type,
  expiresAt,
  maxUses,
  usedCount = 0,
}: InvitationLinkProps) {
  const [copied, setCopied] = useState(false);

  const baseUrl = typeof window !== "undefined" ? window.location.origin : "";
  const inviteLink = `${baseUrl}/invite/${type}/${code}`;

  const handleCopy = async () => {
    await navigator.clipboard.writeText(inviteLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isExpired = expiresAt && new Date(expiresAt) < new Date();
  const isMaxedOut = maxUses && usedCount >= maxUses;

  return (
    <div className="p-4 bg-gray-100 dark:bg-gray-800 rounded-lg">
      <div className="flex items-center gap-2 mb-2">
        <LinkIcon className="w-5 h-5 text-blue-500" />
        <h3 className="font-semibold">
          {type === "chat" ? "Group" : "Friend"} Invitation Link
        </h3>
      </div>

      <div className="flex items-center gap-2 mb-3">
        <input
          type="text"
          value={inviteLink}
          readOnly
          className="flex-1 px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-sm"
        />
        <button
          onClick={handleCopy}
          className="p-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          disabled={isExpired || !!isMaxedOut}
        >
          {copied ? (
            <Check className="w-5 h-5" />
          ) : (
            <Copy className="w-5 h-5" />
          )}
        </button>
      </div>

      <div className="text-xs text-gray-600 dark:text-gray-400 space-y-1">
        {expiresAt && (
          <div className={isExpired ? "text-red-500" : ""}>
            {isExpired ? "Expired" : "Expires"}:{" "}
            {new Date(expiresAt).toLocaleString()}
          </div>
        )}
        {maxUses && (
          <div className={isMaxedOut ? "text-red-500" : ""}>
            Uses: {usedCount} / {maxUses}
          </div>
        )}
        {!maxUses && !expiresAt && (
          <div className="text-green-600 dark:text-green-400">
            ✓ No expiration or usage limit
          </div>
        )}
      </div>

      {(isExpired || isMaxedOut) && (
        <div className="mt-2 text-sm text-red-600 dark:text-red-400">
          This invitation link is no longer valid
        </div>
      )}
    </div>
  );
}
