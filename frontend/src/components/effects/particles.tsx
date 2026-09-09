import { useMemo, type CSSProperties } from "react";

const PARTICLE_COUNT = 100;

/**
 * Mirrors the Sass `random($limit)` helper the effect was written with:
 * an integer between 1 and `limit`.
 */
function random(limit: number) {
  return Math.floor(Math.random() * limit) + 1;
}

function createParticleStyle(): CSSProperties {
  const startPositionY = random(10) + 100;

  return {
    "--particle-size": `${random(8)}px`,
    "--particle-duration": `${28000 + random(9000)}ms`,
    "--particle-delay": `${random(37000)}ms`,
    "--particle-circle-delay": `${random(4000)}ms`,
    "--particle-x-from": `${random(100)}vw`,
    "--particle-y-from": `${startPositionY}vh`,
    "--particle-x-to": `${random(100)}vw`,
    "--particle-y-to": `${-startPositionY - random(30)}vh`,
  } as CSSProperties;
}

export function Particles() {
  const particles = useMemo(
    () =>
      Array.from({ length: PARTICLE_COUNT }, (_, index) => ({
        id: index,
        style: createParticleStyle(),
      })),
    [],
  );

  return (
    <>
      {particles.map((particle) => (
        <div
          key={particle.id}
          className="circle-container"
          style={particle.style}
        >
          <div className="circle"></div>
        </div>
      ))}
    </>
  );
}
