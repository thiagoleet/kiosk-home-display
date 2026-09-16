import { milkpiProfile } from "./milkpi";
import { snespiProfile } from "./snespi";

import { kioskProfileId } from "../config/profile-config";

import type { KioskProfile } from "../types/kiosk-profile";

const profiles: Record<string, KioskProfile> = {
  [milkpiProfile.id]: milkpiProfile,
  [snespiProfile.id]: snespiProfile,
};

// An unset or unknown VITE_KIOSK_PROFILE falls back to MILKPI, which is what
// the build did before profiles could be selected, so `pnpm dev` and a plain
// `pnpm build` keep working. A typo in an environment file would otherwise
// deploy the wrong box silently, so it is called out.
function resolveProfile(id: string): KioskProfile {
  if (!id) {
    return milkpiProfile;
  }

  const profile = profiles[id];

  if (!profile) {
    console.error(
      `Unknown kiosk profile "${id}". Falling back to ${milkpiProfile.id}.`,
    );

    return milkpiProfile;
  }

  return profile;
}

export const kioskProfile = resolveProfile(kioskProfileId);
