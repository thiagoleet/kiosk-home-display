import { useEffect, useRef, useState, type ReactNode } from "react";

type TransitionPhase = "idle" | "exiting" | "entering";

type TransitionDirection = "forward" | "backward" | "notification";

type TransitionState = {
  mode: string;
  content: ReactNode;
  phase: TransitionPhase;
  direction: TransitionDirection;
};

type KioskTransitionProps = {
  mode: string;
  transitionKey: string;
  children: ReactNode;
};

const EXIT_DURATION = 200;
const ENTER_DURATION = 350;
const NOTIFICATION_EXIT_DURATION = 120;

export function KioskTransition({
  mode,
  transitionKey,
  children,
}: KioskTransitionProps) {
  const [transition, setTransition] = useState<TransitionState>({
    mode,
    content: children,
    phase: "idle",
    direction: "forward",
  });

  const pendingContentRef = useRef<ReactNode>(children);

  const previousModeRef = useRef(mode);

  const previousKeyRef = useRef(transitionKey);

  /**
   * Always keep the latest content available
   * without causing a render.
   *
   * This is important because the content may
   * change while the transition is in progress.
   */
  useEffect(() => {
    pendingContentRef.current = children;
  }, [children]);

  /**
   * Detect changes to the screen or notification.
   */
  useEffect(() => {
    if (previousKeyRef.current === transitionKey) {
      return;
    }

    const previousMode = previousModeRef.current;

    const isNotificationChange =
      previousMode === "notification" && mode === "notification";

    previousModeRef.current = mode;
    previousKeyRef.current = transitionKey;

    const direction: TransitionDirection = isNotificationChange
      ? "notification"
      : previousMode === "home"
        ? "forward"
        : "backward";

    const exitDuration = isNotificationChange
      ? NOTIFICATION_EXIT_DURATION
      : EXIT_DURATION;

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
    }, exitDuration);

    return () => {
      window.clearTimeout(exitTimer);
    };
  }, [mode, transitionKey]);

  /**
   * Finish the entering phase.
   */
  useEffect(() => {
    if (transition.phase !== "entering") {
      return;
    }

    const enterTimer = window.setTimeout(
      () => {
        setTransition((current) => ({
          ...current,
          phase: "idle",
        }));
      },
      transition.direction === "notification" ? 220 : ENTER_DURATION,
    );

    return () => {
      window.clearTimeout(enterTimer);
    };
  }, [transition.phase, transition.direction]);

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
