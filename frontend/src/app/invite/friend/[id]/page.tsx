"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { API_BASE_URL } from "@/constants";
import { AuthService } from "@/lib/auth";

export default function FriendInvitePage() {
  const router = useRouter();
  const params = useParams();
  const [status, setStatus] = useState<"loading" | "success" | "error">(
    "loading"
  );
  const [message, setMessage] = useState("");
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!isMounted) return;

    const acceptInvitation = async () => {
      const code = params.id as string;
      if (!code) {
        setStatus("error");
        setMessage("Invalid invitation code");
        return;
      }

      try {
        if (!AuthService.isAuthenticated()) {
          setStatus("error");
          setMessage(
            "You need to log in to accept this invitation. Redirecting to login..."
          );
          router.push(
            `/login?redirect=${encodeURIComponent(`/invite/friend/${code}`)}`
          );
          return;
        }

        const response = await fetch(
          `${API_BASE_URL}/api/invitations/friend/${code}/accept`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              ...AuthService.getAuthHeader(),
            },
          }
        );

        if (response.status === 401) {
          router.push(
            `/login?redirect=${encodeURIComponent(`/invite/friend/${code}`)}`
          );
          return;
        }

        if (response.ok) {
          setStatus("success");
          setMessage("Friend invitation accepted successfully!");
          // Redirect to main page after 2 seconds
          setTimeout(() => {
            router.push("/");
          }, 2000);
        } else {
          const errorData = await response.json().catch(() => ({}));
          setStatus("error");
          setMessage(errorData.error || "Failed to accept invitation");
        }
      } catch (error) {
        console.error("Error accepting invitation:", error);
        setStatus("error");
        setMessage("Network error. Please try again.");
      }
    };

    acceptInvitation();
  }, [params.id, router, isMounted]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full bg-white dark:bg-gray-800 rounded-lg shadow-lg p-8">
        <div className="text-center">
          {status === "loading" && (
            <>
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                Accepting Invitation
              </h2>
              <p className="text-gray-600 dark:text-gray-400">
                Please wait while we process your friend invitation...
              </p>
            </>
          )}

          {status === "success" && (
            <>
              <div className="w-12 h-12 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg
                  className="w-6 h-6 text-green-600 dark:text-green-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                Invitation Accepted!
              </h2>
              <p className="text-gray-600 dark:text-gray-400 mb-4">{message}</p>
              <p className="text-sm text-gray-500 dark:text-gray-500">
                Redirecting to chat...
              </p>
            </>
          )}

          {status === "error" && (
            <>
              <div className="w-12 h-12 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg
                  className="w-6 h-6 text-red-600 dark:text-red-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                Invitation Failed
              </h2>
              <p className="text-gray-600 dark:text-gray-400 mb-4">{message}</p>
              <button
                onClick={() => router.push("/")}
                className="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                Go to Chat
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
