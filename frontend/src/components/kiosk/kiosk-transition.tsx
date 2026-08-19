import { useEffect, useRef, useState, type ReactNode } from "react";

type KioskTransitionProps = {
  mode: string;
  children: ReactNode;
};

type TransitionDirection = "forward" | "backward";

type TransitionState = {
  mode: string;
  content: ReactNode;
  phase: "idle" | "exiting" | "entering";
  direction: TransitionDirection;
};

const EXIT_DURATION = 200;
const ENTER_DURATION = 350;

export function KioskTransition({ mode, children }: KioskTransitionProps) {
  const [transition, setTransition] = useState<TransitionState>({
    mode,
    content: children,
    phase: "idle",
    direction: "forward",
  });

  const pendingContentRef = useRef<ReactNode>(children);

  const previousModeRef = useRef(mode);

  useEffect(() => {
    pendingContentRef.current = children;
  });

  useEffect(() => {
    if (previousModeRef.current === mode) {
      return;
    }

    const previousMode = previousModeRef.current;

    previousModeRef.current = mode;

    const direction: TransitionDirection =
      previousMode === "home" && mode === "notification"
        ? "forward"
        : "backward";

    setTransition((current) => ({
      ...current,
      phase: "exiting",
      direction,
    }));

    const exitTimer = window.setTimeout(() => {
      setTransition({
        mode,
        content: pendingContentRef.current,
        phase: "entering",
        direction,
      });
    }, EXIT_DURATION);

    return () => {
      window.clearTimeout(exitTimer);
    };
  }, [mode]);

  useEffect(() => {
    if (transition.phase !== "entering") {
      return;
    }

    const enterTimer = window.setTimeout(() => {
      setTransition((current) => ({
        ...current,
        phase: "idle",
      }));
    }, ENTER_DURATION);

    return () => {
      window.clearTimeout(enterTimer);
    };
  }, [transition.phase]);

  return (
    <div
      className="kiosk-transition"
      data-phase={transition.phase}
      data-mode={transition.mode}
      data-direction={transition.direction}
    >
      {transition.content}
    </div>
  );
}
