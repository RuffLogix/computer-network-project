"use client";

import { useState } from "react";
import { AuthService } from "@/lib/auth";
import { useRouter } from "next/navigation";

export function LoginForm() {
  const router = useRouter();
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [formData, setFormData] = useState({
    username: "",
    password: "",
    name: "",
    email: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      if (isLogin) {
        await AuthService.login({
          username: formData.username,
          password: formData.password,
        });
      } else {
        await AuthService.register(formData);
      }
      router.push("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed");
    } finally {
      setLoading(false);
    }
  };

  const handleGuestLogin = async () => {
    setLoading(true);
    setError("");

    try {
      await AuthService.createGuest("Guest User");
      router.push("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Guest login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md mx-auto p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-6 text-center">
        {isLogin ? "Login" : "Register"}
      </h2>

      {error && (
        <div className="mb-4 p-3 bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200 rounded">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {!isLogin && (
          <div>
            <label className="block text-sm font-medium mb-1">Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700"
              required={!isLogin}
            />
          </div>
        )}

        <div>
          <label className="block text-sm font-medium mb-1">Username</label>
          <input
            type="text"
            value={formData.username}
            onChange={(e) =>
              setFormData({ ...formData, username: e.target.value })
            }
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700"
            required
          />
        </div>

        {!isLogin && (
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) =>
                setFormData({ ...formData, email: e.target.value })
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700"
              required={!isLogin}
            />
          </div>
        )}

        <div>
          <label className="block text-sm font-medium mb-1">Password</label>
          <input
            type="password"
            value={formData.password}
            onChange={(e) =>
              setFormData({ ...formData, password: e.target.value })
            }
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700"
            required
            minLength={6}
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white rounded-md disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? "Loading..." : isLogin ? "Login" : "Register"}
        </button>
      </form>

      <div className="mt-4 text-center">
        <button
          onClick={() => setIsLogin(!isLogin)}
          className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400"
        >
          {isLogin ? "Need an account? Register" : "Have an account? Login"}
        </button>
      </div>

      <div className="mt-6 pt-6 border-t border-gray-300 dark:border-gray-600">
        <button
          onClick={handleGuestLogin}
          disabled={loading}
          className="w-full py-2 px-4 bg-gray-600 hover:bg-gray-700 text-white rounded-md disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? "Loading..." : "Continue as Guest"}
        </button>
        <p className="mt-2 text-xs text-gray-500 dark:text-gray-400 text-center">
          Guests can only access public chats
        </p>
      </div>
    </div>
  );
}
