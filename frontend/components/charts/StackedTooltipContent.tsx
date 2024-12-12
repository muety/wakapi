import { startCase } from "lodash";

import { convertSecondsToHoursAndMinutes, deTransparentize } from "@/lib/utils";

export interface TooltipPayload {
  name: string;
  value: number;
  color: string;
}

export interface TooltipRowProp {
  payload: TooltipPayload;
  total: number;
  showPercentage?: boolean;
}

export function TooltipRow({ payload, total, showPercentage }: TooltipRowProp) {
  const labelColor = payload.color
    ? deTransparentize(payload.color).color
    : payload.color;
  return (
    <div
      key={payload.name}
      style={{
        zIndex: 1000,
        fontSize: "12px",
        borderRadius: "5px",
        padding: 5,
        gap: 10,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
      }}
    >
      <div
        className="flex"
        style={{
          fontSize: "12px",
          padding: 5,
          height: 8,
          display: "flex",
          alignItems: "center",
        }}
      >
        <div
          style={{
            backgroundColor: labelColor,
            minHeight: 8,
            minWidth: 8,
            maxWidth: 8,
            marginRight: 5,
          }}
          className="rounded"
        ></div>
        {total === payload.value ? (
          <strong>{startCase(payload.name)} </strong>
        ) : (
          startCase(payload.name)
        )}{" "}
      </div>
      <div style={{ textAlign: "left" }} className="flex justify-start">
        {convertSecondsToHoursAndMinutes(payload.value)}
        {showPercentage && (
          <span className="pl-1">
            ({((payload.value / total) * 100).toFixed(2)}%)
          </span>
        )}
      </div>
    </div>
  );
}

export function StackedTooltipContent(props: any) {
  if (props.payload.length > 0) {
    const total =
      props.payload.find((payload: any) => payload["name"] === "total")
        ?.value || 0;
    const selectedPayload = props.payload.filter(
      (p: { name: string }) => !["key", "total"].includes(p.name)
    );
    return (
      <div className="custom-tooltip">
        <div
          className="custom-tooltip-header text-center shadow"
          style={{ color: "white" }}
        >
          {props.label}
        </div>
        <TooltipRow
          payload={{ name: "Total", value: total, color: "#fff" }}
          total={total}
        />
        {selectedPayload
          .filter((p: any) => p.name !== "key") //  && p.value >= 60
          .filter((p: any) => p?.value > 0) // TODO: update?
          .map((payload: any, index: number) => (
            <TooltipRow payload={payload} total={total} key={index} />
          ))}
      </div>
    );
  }
  return null;
}

export function StackedTooltipContentForCategories(props: any) {
  if (props.payload.length > 0) {
    let total = 0;

    props.payload.forEach((payload: any) => {
      Object.keys(payload).forEach((key: any) => {
        if (typeof payload[key] === "number") {
          total += payload[key];
        }
      });
    });
    const getTitleFromPayload = () => {
      const target = props.payload[props.label];
      return target ? startCase(target.dataKey || "") : "";
    };
    const labelTitle =
      typeof props.label === "string" ? props.label : getTitleFromPayload();
    return (
      <div className="custom-tooltip">
        <div className="custom-tooltip-header shadow">{labelTitle}</div>
        {props.payload
          .filter((p: any) => p.name !== "key")
          .map((payload: any, index: number) => (
            <TooltipRow
              payload={payload}
              key={index}
              total={total}
              showPercentage={true}
            />
          ))}
      </div>
    );
  }
  return null;
}
