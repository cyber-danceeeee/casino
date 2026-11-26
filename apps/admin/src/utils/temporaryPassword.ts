function secureRandomIndex(length: number) {
  const limit = 256 - (256 % length)
  const bytes = new Uint8Array(1)
  do globalThis.crypto.getRandomValues(bytes); while (bytes[0] >= limit)
  return bytes[0] % length
}

export function generateTemporaryPassword() {
  const groups = ['ABCDEFGHJKLMNPQRSTUVWXYZ', 'abcdefghijkmnopqrstuvwxyz', '23456789', '!@#$%^&*']
  const all = groups.join('')
  const characters = groups.map((group) => group[secureRandomIndex(group.length)])
  while (characters.length < 20) characters.push(all[secureRandomIndex(all.length)])
  for (let index = characters.length - 1; index > 0; index -= 1) {
    const target = secureRandomIndex(index + 1)
    ;[characters[index], characters[target]] = [characters[target], characters[index]]
  }
  return characters.join('')
}
