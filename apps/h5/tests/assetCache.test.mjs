import assert from 'node:assert/strict'
import test from 'node:test'
import { MessageChannel } from 'node:worker_threads'

let scenarioID = 0

async function browserScenario({ secure = false, navigator = {}, messageChannel = MessageChannel }, run) {
  const globals = {
    window: {
      isSecureContext: secure,
      setTimeout,
      clearTimeout,
      location: { origin: 'http://app.test', reload() {} },
    },
    navigator,
    document: { querySelectorAll: () => [] },
    MessageChannel: messageChannel,
  }
  const previous = new Map(Object.keys(globals).map(key => [key, Object.getOwnPropertyDescriptor(globalThis, key)]))
  try {
    for (const [key, value] of Object.entries(globals)) {
      Object.defineProperty(globalThis, key, { configurable: true, value })
    }
    const cache = await import(`../src/services/assetCache.ts?scenario=${++scenarioID}`)
    await run(cache)
  } finally {
    for (const [key, descriptor] of previous) {
      if (descriptor) Object.defineProperty(globalThis, key, descriptor)
      else delete globalThis[key]
    }
  }
}

async function assertCacheUnavailable(cache) {
  assert.equal(await cache.registerAssetCache(), null)
  assert.equal(await cache.configureAssetCache('v1'), false)
  assert.equal(await cache.cacheApplicationShell(), false)
  assert.equal(await cache.prefetchAssetURLs(['/image.webp']), false)
  assert.equal(await cache.prefetchMusicURLs(['/music.mp3']), false)
}

test('HTTP origins without Service Worker support can complete all startup cache calls', async () => {
  await browserScenario({}, assertCacheUnavailable)
})

test('HTTP origins do not access the restricted Service Worker API', async () => {
  const navigator = Object.defineProperty({}, 'serviceWorker', {
    get() { throw new Error('The API must not be accessed in an insecure context') },
  })
  await browserScenario({ navigator }, assertCacheUnavailable)
})

test('secure browsers without Service Worker support or with access disabled remain usable', async () => {
  await browserScenario({ secure: true }, assertCacheUnavailable)
  const navigator = Object.defineProperty({}, 'serviceWorker', {
    get() { throw new Error('Service Worker access is disabled') },
  })
  await browserScenario({ secure: true, navigator }, assertCacheUnavailable)
})

test('a rejected registration leaves caching optional', async t => {
  t.mock.method(console, 'warn', () => {})
  const navigator = {
    serviceWorker: {
      controller: null,
      addEventListener() {},
      register: async () => { throw new Error('Registration rejected by browser policy') },
    },
  }
  await browserScenario({ secure: true, navigator }, assertCacheUnavailable)
})

function workingServiceWorker(postMessage) {
  const worker = { postMessage }
  const registration = { active: worker, update: async () => {} }
  return {
    serviceWorker: {
      controller: worker,
      addEventListener() {},
      ready: Promise.resolve(registration),
      register: async () => registration,
    },
  }
}

test('supported secure browsers still configure and prefetch through the worker', async () => {
  const messages = []
  const navigator = workingServiceWorker((payload, ports) => {
    messages.push(payload)
    ports[0].postMessage({ ok: true })
    ports[0].close()
  })
  await browserScenario({ secure: true, navigator }, async cache => {
    assert.equal(await cache.configureAssetCache('published-v2'), true)
    assert.equal(await cache.cacheApplicationShell(), true)
    assert.equal(await cache.prefetchAssetURLs(['/image.webp', '/image.webp', '']), true)
    assert.equal(await cache.prefetchMusicURLs(['/music.mp3']), true)
  })
  assert.deepEqual(messages.map(message => message.type), [
    'SET_ASSET_VERSION', 'CACHE_APP_SHELL', 'PREFETCH_ASSETS', 'PREFETCH_MUSIC',
  ])
  assert.equal(messages[0].version, 'published-v2')
  assert.deepEqual(messages[2].urls, ['/image.webp'])
})

test('a worker that cannot accept messages does not block application startup', async () => {
  const navigator = workingServiceWorker(() => { throw new Error('Worker is no longer available') })
  await browserScenario({ secure: true, navigator }, async cache => {
    assert.equal(await cache.configureAssetCache('v1'), false)
    assert.equal(await cache.cacheApplicationShell(), false)
  })
})

test('browsers without MessageChannel skip worker messaging', async () => {
  const navigator = workingServiceWorker(() => assert.fail('Messaging must be skipped'))
  await browserScenario({ secure: true, navigator }, async cache => {
    Object.defineProperty(globalThis, 'MessageChannel', { configurable: true, value: undefined })
    assert.equal(await cache.configureAssetCache('v1'), false)
  })
})
