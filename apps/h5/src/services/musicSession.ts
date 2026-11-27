import type { H5MusicConfig, H5MusicTrack } from "./h5Api";
import { prefetchMusicURLs } from "./assetCache";
import { backgroundMusic } from "../config/media";

export interface H5MusicSession {
  config: H5MusicConfig;
  playlist: H5MusicTrack[];
}
let sessionPromise: Promise<H5MusicSession> | null = null;
export const loadMusicSession = () => {
  if (sessionPromise) return sessionPromise;
  sessionPromise = Promise.resolve().then(() => {
    const config = backgroundMusic;
    const playlist = config.tracks.filter((track) => track.audioUrl);
    void prefetchMusicURLs(playlist.map((track) => track.audioUrl));
    return { config, playlist };
  });
  return sessionPromise;
};
export const loadMusicSessionWithin = async (timeoutMs: number) => {
  let timer = 0;
  try {
    return await Promise.race([
      loadMusicSession(),
      new Promise<null>((resolve) => {
        timer = window.setTimeout(() => resolve(null), timeoutMs);
      }),
    ]);
  } finally {
    if (timer) window.clearTimeout(timer);
  }
};
