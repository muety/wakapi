import React from "react";

interface iProps {
  options: string[];
  onChange?: (value: string) => void;
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
    <div className="code-more-anchor px-3 py-2 text-sm" onClick={clickHandler}>
      <span>{value || options[index]}</span>
    </div>
  );
}
