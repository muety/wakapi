"use client";

import React, { useEffect, useRef, useState } from "react";

interface Option {
  id: string;
  label: string;
}

interface TypeformSelectProps {
  options: Option[];
  onChange: (selectedOption: Option) => void;
}

export default function TypeformSelect({
  options,
  onChange,
}: TypeformSelectProps) {
  const [selectedOption, setSelectedOption] = useState<Option | null>(null);
  const [focusedIndex, setFocusedIndex] = useState<number>(-1);
  const buttonRefs = useRef<(HTMLButtonElement | null)[]>([]);

  useEffect(() => {
    buttonRefs.current = buttonRefs.current.slice(0, options.length);
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
      buttonRefs.current[focusedIndex]?.focus();
    }
  }, [focusedIndex]);

  return (
    <div
      className=" items-center gap-2 space-y-4"
      role="radiogroup"
      aria-label="Select an option"
      onKeyDown={handleKeyDown}
    >
      {options.map((option, index) => (
        <button
          key={option.id}
          ref={(el) => (buttonRefs.current[index] = el) as any}
          className={`
            relative h-10 w-full rounded-lg p-1 text-left text-lg font-medium transition-all duration-200 ease-in-out focus:outline-none
            focus:ring-1 focus:ring-primary focus:ring-offset-2
            ${
              selectedOption?.id === option.id
                ? "bg-primary text-primary-foreground"
                : "bg-background text-foreground hover:bg-secondary/50"
            }
          `}
          role="radio"
          aria-checked={selectedOption?.id === option.id}
          onClick={() => handleSelect(option)}
        >
          <span className="relative z-10">{option.label}</span>
          <Ripple />
        </button>
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
      className="pointer-events-none absolute inset-0 overflow-hidden rounded-lg"
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
