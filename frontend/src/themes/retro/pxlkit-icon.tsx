import {
  AnimatedPxlKitIcon,
  PxlKitIcon,
  type AnimatedPxlKitData,
  type PxlKitData,
} from "@pxlkit/core";

type PxlkitIconProps = {
  icon: PxlKitData;
};

type AnimatedIconProps = {
  icon: AnimatedPxlKitData;
};

export function RetroIcon({ icon }: PxlkitIconProps) {
  return <PxlKitIcon icon={icon} />;
}

export function AnimatedIcon({ icon }: AnimatedIconProps) {
  return <AnimatedPxlKitIcon icon={icon} />;
}
