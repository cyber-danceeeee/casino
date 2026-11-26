let audioContext: AudioContext | null = null

// Prime the top-level audio session while the game-card click still has user activation.
export function primeGameAudio() {
  if (typeof window === 'undefined') return
  const AudioContextClass = window.AudioContext
    || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!AudioContextClass) return
  try {
    audioContext ||= new AudioContextClass()
    void audioContext.resume()
    const source = audioContext.createBufferSource()
    source.buffer = audioContext.createBuffer(1, 1, 22050)
    const gain = audioContext.createGain()
    gain.gain.value = 0
    source.connect(gain).connect(audioContext.destination)
    source.start(0)
  } catch {
    // The provider iframe also receives an explicit cross-origin autoplay permission.
  }
}
