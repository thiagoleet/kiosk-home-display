import { useEffect, useRef, useState, type ReactNode } from "react";

type KioskTransitionProps = {
  mode: string;
  children: ReactNode;
};

type TransitionState = {
  mode: string;
  content: ReactNode;
  phase: "idle" | "exiting" | "entering";
};

const EXIT_DURATION = 200;
const ENTER_DURATION = 350;

export function KioskTransition({ mode, children }: KioskTransitionProps) {
  const [transition, setTransition] = useState<TransitionState>({
    mode,
    content: children,
    phase: "idle",
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

    previousModeRef.current = mode;

    setTransition((current) => ({
      ...current,
      phase: "exiting",
    }));

    const exitTimer = window.setTimeout(() => {
      setTransition({
        mode,
        content: pendingContentRef.current,
        phase: "entering",
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
    >
      {transition.content}
    </div>
  );
}
