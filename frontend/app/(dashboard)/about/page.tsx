export default function Page() {
  return (
    <div
      className="m-auto mx-14 flex flex-col justify-center px-14 align-middle"
      style={{ minHeight: "70vh" }}
    >
      <h1 className="mb-8 text-center text-6xl">About</h1>
      <p className="mb-5">
        The goal of this project is to provide a self-hosted version of
        wakatime.com that is open source and free to use. We rely on the same
        open source plugins and collect the same data that is available from the
        plugins.
      </p>
      <p className="mb-5">
        Unlike some of the open source alternatives, we aim for feature parity
        with wakatime first and foremost. We start with a focus on individual
        features and then proceed towards enterprise/organizational features.
      </p>
      <p className="mb-5">
        The managed version of this website is only available to paying
        customers to help fund the continued work on this project.
      </p>
      <p className="mb-5">
        Work on this project was sped up by the open source project at{" "}
        <a href="https://github.com/muety/wakapi" className="underline">
          Wakapi
        </a>
        . We are grateful for their work. We won&apos;t have built this quickly
        without starting off of that base source code. It was minimal but
        thoughtful and was still packed with a ton of features.
      </p>
    </div>
  );
}
