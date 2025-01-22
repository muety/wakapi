import PricingCard from "../pricing-card";

export const Prices = [
  {
    title: "Free Plan",
    price: 0,
    period: "mo",
    features: [
      "1 month dashboard history",
      "weekly email reports",
      "unlimited programming goals",
      "public leaderboard",
      "invoices generated from code stats for clients",
      "dashboard history for every project",
      "wakatime data migration and integration",
    ],
    ctaText: "Free Forever",
  },
  {
    title: "Premium Plan",
    price: 5.99,
    period: "mo",
    features: [
      "everything in free plan",
      "unlimited dashboard history",
      "weekly email reports to up to five other emails",
      "priority support",
    ],
    ctaText: "Upgrade to Premium",
    ctaClassName: "bg-green-600 hover:bg-green-700 font-bold",
  },
];

export function Pricing() {
  return (
    <div className="m-auto mx-14 flex flex-col justify-center px-14 align-middle">
      <h1 className="mb-8 text-center text-6xl">Pricing</h1>
      <p>
        Wakana is free to use forever. You&apos;ll have 1 month dashboard
        history in addition to all features like goals, invoices and per project
        views as well as shareables for websites like github - all without
        paying a cent.
      </p>
      <br />
      <p>
        To access premium features like unlimited dashboard history, you might
        want to upgrade to the paid plan. This plan keeps your data forever and
        makes it accessible to you every time.
      </p>

      <div className="mt-12">
        <h1 className="my-5 mt-12 text-center text-4xl">For Individuals</h1>
        <div className="mt-12 flex flex-wrap justify-center gap-8">
          {Prices.map((price) => (
            <PricingCard key={price.title} {...price} />
          ))}
        </div>
      </div>
    </div>
  );
}
