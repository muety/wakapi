"use client";

import React, { useEffect, useRef, useState } from "react";

import { Card, CardContent } from "@/components/ui/card";

interface Option {
  id: string;
  label: string;
}

interface CardSelectProps {
  options: Option[];
  onChange: (selectedOption: Option) => void;
}

export default function CardSelect({ options, onChange }: CardSelectProps) {
  const [selectedOption, setSelectedOption] = useState<Option | null>(null);
  const [focusedIndex, setFocusedIndex] = useState<number>(-1);
  const cardRefs = useRef<(HTMLDivElement | null)[]>([]);

  useEffect(() => {
    cardRefs.current = cardRefs.current.slice(0, options.length);
  }, [options]);

  const handleSelect = (option: Option) => {
    setSelectedOption(option);
    onChange(option);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case "ArrowDown":
        setFocusedIndex((prevIndex) => (prevIndex + 1) % options.length);
        break;
      case "ArrowUp":
        setFocusedIndex(
          (prevIndex) => (prevIndex - 1 + options.length) % options.length
        );
        break;
      case "Enter":
      case " ":
        if (focusedIndex !== -1) {
          handleSelect(options[focusedIndex]);
        }
        break;
    }
  };

  useEffect(() => {
    if (focusedIndex !== -1) {
      cardRefs.current[focusedIndex]?.focus();
    }
  }, [focusedIndex]);

  return (
    <div
      className="grid grid-cols-3 gap-4"
      role="listbox"
      aria-label="Select an option"
      onKeyDown={handleKeyDown}
    >
      {options.map((option, index) => (
        <Card
          key={option.id}
          ref={(el) => (cardRefs.current[index] = el) as any}
          className={`
            cursor-pointer transition-all duration-300 ease-in-out
            ${
              selectedOption?.id === option.id
                ? "ring-2 ring-primary"
                : "hover:shadow-md"
            }
            focus:outline-none focus:ring-2 focus:ring-primary
          `}
          role="option"
          aria-selected={selectedOption?.id === option.id}
          tabIndex={0}
          onClick={() => handleSelect(option)}
        >
          <CardContent className="relative overflow-hidden p-4">
            <span className="text-sm font-medium">{option.label}</span>
            <Ripple />
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function Ripple() {
  const [ripples, setRipples] = useState<
    { x: number; y: number; size: number }[]
  >([]);

  const addRipple = (e: React.MouseEvent<HTMLDivElement>) => {
    const rippleContainer = e.currentTarget.getBoundingClientRect();
    const size =
      rippleContainer.width > rippleContainer.height
        ? rippleContainer.width
        : rippleContainer.height;
    const x = e.clientX - rippleContainer.left - size / 2;
    const y = e.clientY - rippleContainer.top - size / 2;
    const newRipple = { x, y, size };

    setRipples((prevRipples) => [...prevRipples, newRipple]);
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      setRipples([]);
    }, 1000);

    return () => clearTimeout(timer);
  }, [ripples]);

  return (
    <div
      className="pointer-events-none absolute inset-0"
      onMouseDown={addRipple}
    >
      {ripples.map((ripple, index) => (
        <span
          key={index}
          style={{
            top: ripple.y,
            left: ripple.x,
            width: ripple.size,
            height: ripple.size,
          }}
          className="animate-ripple absolute scale-0 rounded-full bg-white opacity-30"
        />
      ))}
    </div>
  );
}
