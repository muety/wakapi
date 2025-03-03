import {
  BarChart3,
  Clock,
  FileText,
  Flag,
  Layers,
  Repeat,
  Trophy,
  Users,
} from "lucide-react";

import { FadeOnView } from "./fade-on-view";
import { FeatureCard } from "./feature-card";
import { StatCard } from "./start-card";
import Link from "next/link";

const features = [
  {
    icon: Flag,
    title: "Goals",
    description:
      "Create and track programming goals with customizable targets. Set daily, weekly, or monthly coding objectives and monitor your progress.",
    iconColor: "text-blue-400",
    iconBgColor: "bg-blue-900/30",
  },
  {
    icon: FileText,
    title: "Invoices",
    description:
      "Generate professional invoices based on your programming activity for any project. Automatically track billable hours and create detailed reports.",
    iconColor: "text-green-400",
    iconBgColor: "bg-green-900/30",
  },
  {
    icon: Layers,
    title: "Projects",
    description:
      "View comprehensive analytics for individual projects. Track time spent, languages used, and productivity metrics to optimize your workflow.",
    iconColor: "text-purple-400",
    iconBgColor: "bg-purple-900/30",
  },
  {
    icon: Users,
    title: "Client Management",
    description:
      "Create and manage freelance clients with ease. Organize projects by client, track billable hours, and maintain client-specific settings.",
    iconColor: "text-yellow-400",
    iconBgColor: "bg-yellow-900/30",
  },
  {
    icon: Trophy,
    title: "Public Leaderboards",
    description:
      "See how you measure up against other developers. Compare coding metrics, languages, and productivity with the global developer community.",
    iconColor: "text-red-400",
    iconBgColor: "bg-red-900/30",
  },
  {
    icon: Repeat,
    title: "Wakatime Relay",
    description:
      "Set up Wakatime integration to relay all heartbeats. Perfect for trying our platform while maintaining your existing Wakatime setup and data.",
    iconColor: "text-teal-400",
    iconBgColor: "bg-teal-900/30",
  },
];

const stats = [
  {
    icon: Clock,
    value: "10,000+",
    label: "Hours Tracked",
    iconColor: "text-blue-400",
  },
  {
    icon: Users,
    value: "5,000+",
    label: "Active Developers",
    iconColor: "text-green-400",
  },
  {
    icon: Layers,
    value: "25,000+",
    label: "Projects Managed",
    iconColor: "text-purple-400",
  },
  {
    icon: BarChart3,
    value: "1M+",
    label: "Data Points",
    iconColor: "text-yellow-400",
  },
];

export default function FeaturesSection() {
  return (
    <div className="w-full text-white py-16 px-4 md:py-24">
      <div className="container mx-auto">
        <FadeOnView>
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Powerful Features for Developers
            </h2>
            <p className="text-gray-400 max-w-2xl mx-auto">
              Track your coding activity, manage clients, create invoice, set
              productivity goals and observe your productivity with ease.
            </p>
          </div>
        </FadeOnView>

        <FadeOnView>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {features.map((feature, index) => (
              <FeatureCard key={index} {...feature} />
            ))}
          </div>
        </FadeOnView>

        {/* Stats Section */}
        <div className="mt-20 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 sr-only">
          {stats.map((stat, index) => (
            <StatCard key={index} {...stat} />
          ))}
        </div>

        {/* CTA Section */}
        <FadeOnView>
          <div className="mt-16 text-center">
            <Link href='/login' className="bg-white hover:text-gray-500 text-black font-medium py-5 px-8 rounded-sm transition-colors">
              Get Started for Free
            </Link>
            <p className="mt-6 text-gray-400">
              No credit card required. Start tracking in minutes.
            </p>
          </div>
        </FadeOnView>
      </div>
    </div>
  );
}
