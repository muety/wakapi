"use client";

import { useState } from "react";

import { MultiSelect, MultiSelectOption } from "@/components/multi-select";

interface iProps {
  options: MultiSelectOption[];
  onSelectedOptionsChanged: (value: string[]) => any;
  title?: string;
  placeholder: string;
  defaultValue?: string[];
}

function WMultiSelect({
  options,
  onSelectedOptionsChanged,
  title,
  placeholder,
  defaultValue = [],
}: iProps) {
  const [selectedOption, setSelectedOption] = useState<string[]>(defaultValue);

  return (
    <div className="main-bg">
      {title && (
        <h1 className="mb-4" style={{ fontSize: "30px", fontWeight: 400 }}>
          {title}
        </h1>
      )}
      <MultiSelect
        options={options}
        onValueChange={(value: string[]) => {
          setSelectedOption(value);
          onSelectedOptionsChanged(value);
        }}
        defaultValue={selectedOption}
        placeholder={placeholder}
        variant="inverted"
      />
    </div>
  );
}

export default WMultiSelect;
