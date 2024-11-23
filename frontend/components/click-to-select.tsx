import React from "react";

interface iProps {
  options: string[];
  onChange?: Function;
  value?: string;
}

export function ClickToSelect({ options, onChange, value }: iProps) {
  const [index, setIndex] = React.useState(0);

  const clickHandler = (event: { preventDefault: () => void }) => {
    event.preventDefault();
    const newIndex = (index + 1) % options.length;
    const newValue = options[index];
    if (onChange) {
      onChange(newValue);
    }
    setIndex(newIndex);
  };
  return (
    <div className="code-more-anchor text-sm px-3 py-2" onClick={clickHandler}>
      <span>{value || options[index]}</span>
    </div>
  );
}
