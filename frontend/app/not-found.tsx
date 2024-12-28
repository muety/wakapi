import Link from "next/link";

export default function Custom404() {
  return (
    <div className="flex flex-col justify-center items-center h-screen bg-gray-900 text-center">
      <h1 className="text-6xl font-bold md:text-4xl">404</h1>
      <p className="text-lg mt-2 mb-6 text-gray-400 md:text-base">
        Oops! The page you are looking for does not exist.
      </p>
      <Link
        href="/"
        className="px-6 py-3 font-bold rounded-lg shadow-md transition"
      >
        Go Back Home
      </Link>
    </div>
  );
}
