"use client";

import { useState, useEffect } from "react";
import NextImage from "next/image";

interface StickerPickerProps {
  onSelectSticker: (stickerUrl: string) => void;
  isOpen: boolean;
  onClose: () => void;
}

export function StickerPicker({
  onSelectSticker,
  isOpen,
  onClose,
}: StickerPickerProps) {
  const [stickers, setStickers] = useState<string[]>([]);
  const [loadedStickers, setLoadedStickers] = useState<Set<number>>(new Set());

  console.log("StickerPicker render, isOpen:", isOpen, "stickers:", stickers);

  useEffect(() => {
    // Load stickers from /stickers folder
    // For now, we'll assume stickers are named sticker1.svg, sticker2.svg, etc.
    // In a real app, you might fetch this from an API
    const stickerFiles = [];
    for (let i = 1; i <= 10; i++) {
      // Assume up to 10 stickers for demo
      stickerFiles.push(`/stickers/sticker${i}.svg`);
    }
    console.log("Setting stickers:", stickerFiles);
    setStickers(stickerFiles);
  }, []);

  const handleStickerLoad = (index: number) => {
    setLoadedStickers((prev) => new Set(prev).add(index));
  };

  const handleStickerError = (index: number) => {
    setLoadedStickers((prev) => {
      const newSet = new Set(prev);
      newSet.delete(index);
      return newSet;
    });
  };

  if (!isOpen) return null;

  return (
    <div className="absolute bottom-full left-0 mb-2 p-4 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-lg w-80 z-50">
      <div className="grid grid-cols-5 gap-3 max-h-64 overflow-y-auto">
        {stickers.map((sticker, index) => {
          const isLoaded = loadedStickers.has(index);
          return (
            <button
              key={index}
              onClick={() => {
                if (!isLoaded) return; // Prevent clicking empty slots
                console.log("Sticker selected:", sticker);
                onSelectSticker(sticker);
                onClose();
              }}
              disabled={!isLoaded}
              className={`p-2 rounded-lg transition-all ${
                isLoaded
                  ? "hover:scale-110 active:scale-95 cursor-pointer"
                  : "cursor-default opacity-0"
              }`}
            >
              <div className="w-12 h-12 flex items-center justify-center">
                {sticker.endsWith(".svg") ? (
                  <img
                    src={sticker}
                    alt={`Sticker ${index + 1}`}
                    className="w-10 h-10 object-contain"
                    onLoad={() => handleStickerLoad(index)}
                    onError={(e) => {
                      console.log("Failed to load sticker:", sticker);
                      handleStickerError(index);
                      (e.target as HTMLImageElement).style.display = "none";
                    }}
                  />
                ) : (
                  <NextImage
                    src={sticker}
                    alt={`Sticker ${index + 1}`}
                    width={40}
                    height={40}
                    unoptimized
                    className="w-10 h-10 object-contain"
                    onLoad={() => handleStickerLoad(index)}
                    onError={(e) => {
                      console.log("Failed to load sticker:", sticker);
                      handleStickerError(index);
                      (e.target as HTMLImageElement).style.display = "none";
                    }}
                  />
                )}
              </div>
            </button>
          );
        })}
      </div>
      <button
        onClick={onClose}
        className="mt-3 w-full text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 py-1"
      >
        Close
      </button>
    </div>
  );
}
