import { PxlKitIcon } from "@pxlkit/core";
import { Cloud, CloudSun, Rain, Sun } from "@pxlkit/weather";

export function PxlkitTest() {
  return (
    <div style={{ display: "flex", gap: 16 }}>
      <PxlKitIcon
        icon={Sun}
        size={48}
      />
      <PxlKitIcon
        icon={CloudSun}
        size={48}
      />
      <PxlKitIcon
        icon={Cloud}
        size={48}
      />
      <PxlKitIcon
        icon={Rain}
        size={48}
      />
    </div>
  );
}
