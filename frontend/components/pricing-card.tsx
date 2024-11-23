import { cn } from "@/lib/utils";
import { Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

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
    <Card className="w-full sm:w-[300px] flex flex-col overflow-hidden">
      <CardHeader className="border-b border-gray-200 dark:border-gray-700 space-y-4 sm:space-y-6 pb-6">
        <CardTitle className="text-xl sm:text-2xl font-bold text-center text-gray-900 dark:text-white">
          {title}
        </CardTitle>
        <div className="flex flex-col items-center space-y-4">
          <div className="flex items-baseline">
            <span className="text-4xl sm:text-5xl font-bold text-gray-900 dark:text-white">
              ${price}
            </span>
            <span className="text-lg sm:text-xl ml-1 text-gray-600 dark:text-gray-400">
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
      <CardContent className="flex-grow pt-4 sm:pt-6">
        <h3 className="text-base sm:text-lg font-semibold mb-2 sm:mb-4 text-gray-900 dark:text-white">
          What's included:
        </h3>
        <ul className="space-y-2 sm:space-y-4">
          {features.map((feature, index) => (
            <li
              key={index}
              className="flex items-center text-sm sm:text-base text-gray-700 dark:text-gray-300"
            >
              <Check className="h-4 w-4 sm:h-5 sm:w-5 text-green-500 mr-2 sm:mr-3 flex-shrink-0" />
              <span>{feature}</span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}
