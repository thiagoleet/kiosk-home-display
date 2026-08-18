import type { DisplayState } from "../types/display";

type DisplayStatusProps = {
  display: DisplayState;
};

export function DisplayStatus({ display }: DisplayStatusProps) {
  return (
    <section>
      <h2>Display</h2>

      <p>
        Power: <strong>{display.power}</strong>
      </p>

      <p>
        Brightness: <strong>{display.brightness}%</strong>
      </p>
    </section>
  );
}
