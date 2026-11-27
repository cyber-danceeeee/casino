import type { H5MusicConfig, H5MusicTrack } from "../services/h5Api";

// Only versioned background tracks enter this playlist. Notification sounds are separate.
const files = import.meta.glob<string>("../assets/music/*.{mp3,ogg,wav}", {
  eager: true,
  query: "?url",
  import: "default",
});
const tracks: H5MusicTrack[] = Object.entries(files)
  .sort(([a], [b]) => a.localeCompare(b))
  .map(([path, audioUrl], index) => ({
    id: index + 1,
    code:
      path
        .split("/")
        .pop()
        ?.replace(/\.[^.]+$/, "") || `track-${index + 1}`,
    nameZh: "大厅背景音乐",
    nameEn: "Lobby Background Music",
    audioUrl,
    loop: true,
    volumePercent: 35,
  }));
export const backgroundMusic: H5MusicConfig = {
  mode: "SPECIFIED",
  specifiedMusicID: tracks[0]?.id || 0,
  tracks,
};
