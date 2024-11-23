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
    <div
      className="flex flex-col justify-center align-middle m-auto px-14 mx-14 text-md"
      style={{ minHeight: "70vh" }}
    >
      <h1 className="text-6xl mb-8 text-center">Pricing</h1>
      <p>
        Wakana is free to use forever. You'll have 1 month dashboard history in
        addition to all features like goals, invoices and per project views as
        well as shareables for websites like github - all without paying a cent.
      </p>
      <br />
      <p>
        To access premium features like unlimited dashboard history, you might
        want to upgrade to the paid plan. This plan keeps your data forever and
        makes it accessible to you every time.
      </p>

      <div className="mt-12">
        <h1 className="text-center text-4xl my-5 mt-12">For Individuals</h1>
        <div className="flex flex-wrap justify-center gap-8 mt-12">
          {Prices.map((price) => (
            <PricingCard key={price.title} {...price} />
          ))}
        </div>
      </div>
    </div>
  );
}
