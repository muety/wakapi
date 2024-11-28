export function ProjectCard({
  title,
  duration,
}: {
  title: string;
  duration: string;
}) {
  return (
    <div className="project-wrapper cursor-pointer hover:border-white/15 hover:bg-white/[4%]">
      <a className="project-card" href={"/projects/" + title}>
        <div className="project-content">
          <h3 className="text-2xl font-normal">{title}</h3>
          <h4 className="text-lg text-gray-500" style={{ color: "#777777" }}>
            {duration}
          </h4>
        </div>
      </a>
    </div>
  );
}
