// Footer artwork is versioned with H5, independently of the playable catalog.
const images = import.meta.glob<string>('../assets/images/providers/provider-*.webp', {
  eager: true, import: 'default', query: '?url',
})
export const footerBrandImages = Array.from({ length: 11 }, (_, index) =>
  images[`../assets/images/providers/provider-${index + 1}.webp`],
).filter((image): image is string => Boolean(image))
