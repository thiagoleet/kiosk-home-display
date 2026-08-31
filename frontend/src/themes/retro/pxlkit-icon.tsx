import {
  AnimatedPxlKitIcon,
  PxlKitIcon,
  type AnimatedPxlKitData,
  type PxlKitData,
} from "@pxlkit/core";

type PxlkitIconProps = {
  icon: PxlKitData;
  size?: number;
};

type AnimatedIconProps = {
  icon: AnimatedPxlKitData;
  size?: number;
};

export function RetroIcon({ icon, size }: PxlkitIconProps) {
  return (
    <PxlKitIcon
      icon={icon}
      size={size}
    />
  );
}

export function AnimatedIcon({ icon, size }: AnimatedIconProps) {
  return (
    <AnimatedPxlKitIcon
      icon={icon}
      size={size}
    />
  );
}
