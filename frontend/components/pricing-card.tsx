import { Check } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface PricingCardProps {
  title: string;
  price: number;
  period: string;
  features: string[];
  ctaText: string;
  ctaClassName?: string;
}

export default function PricingCard({
  title,
  price,
  period,
  features,
  ctaText,
  ctaClassName = "",
}: PricingCardProps) {
  return (
    <Card className="flex w-full flex-col overflow-hidden sm:w-[300px]">
      <CardHeader className="space-y-4 border-b border-gray-200 pb-6 dark:border-gray-700 sm:space-y-6">
        <CardTitle className="text-center text-xl font-bold text-gray-900 dark:text-white sm:text-2xl">
          {title}
        </CardTitle>
        <div className="flex flex-col items-center space-y-4">
          <div className="flex items-baseline">
            <span className="text-4xl font-bold text-gray-900 dark:text-white sm:text-5xl">
              ${price}
            </span>
            <span className="ml-1 text-lg text-gray-600 dark:text-gray-400 sm:text-xl">
              /{period}
            </span>
          </div>
          <Button
            className={cn(
              "w-full bg-blue-600 hover:bg-blue-700 text-white",
              ctaClassName
            )}
          >
            {ctaText}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="grow pt-4 sm:pt-6">
        <h3 className="mb-2 text-base font-semibold text-gray-900 dark:text-white sm:mb-4 sm:text-lg">
          What&apos;s included:
        </h3>
        <ul className="space-y-2 sm:space-y-4">
          {features.map((feature, index) => (
            <li
              key={index}
              className="flex items-center text-sm text-gray-700 dark:text-gray-300 sm:text-base"
            >
              <Check className="mr-2 size-4 shrink-0 text-green-500 sm:mr-3 sm:size-5" />
              <span>{feature}</span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}
