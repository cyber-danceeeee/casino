#!/usr/bin/env python3
"""Check both deployed portals, SPA routing, bundled media and API proxying."""
import json
import os
from pathlib import Path
import re
from urllib.error import HTTPError
from urllib.request import urlopen

for port, deep_link in [(os.getenv('H5_PORT', '80'), '/wallet'),
                        (os.getenv('ADMIN_PORT', '8080'), '/platform/login')]:
    base = f'http://127.0.0.1:{port}'
    for path in ['/', deep_link, '/health/ready']:
        with urlopen(base + path, timeout=15) as response:
            body = response.read()
            if path == '/health/ready':
                assert json.loads(body)['status'] == 'READY'
            else:
                assert b'<html' in body.lower() and b'<div id="app"' in body
                asset = re.search(rb'(?:src|href)="(/assets/[^" ]+\.js)"', body)
                assert asset, f'No built JavaScript in {base + path}'
                with urlopen(base + asset.group(1).decode(), timeout=15) as js:
                    assert 'javascript' in js.headers['Content-Type']
    # Missing assets must return 404 rather than an HTML SPA fallback.
    try:
        urlopen(base + '/assets/does-not-exist.js', timeout=15)
        raise AssertionError('Missing asset returned success')
    except HTTPError as error:
        assert error.code == 404
    print(f'Portal {port}: index, deep link, JavaScript and API readiness passed')

root = Path(__file__).resolve().parent.parent
asset = next(p for p in (root / 'apps/h5/public/static').rglob('*') if p.is_file())
path = '/static/' + asset.relative_to(root / 'apps/h5/public/static').as_posix()
for port in [os.getenv('H5_PORT', '80'), os.getenv('ADMIN_PORT', '8080')]:
    with urlopen(f'http://127.0.0.1:{port}' + path, timeout=15) as response:
        assert response.read() == asset.read_bytes(), 'Bundled media did not survive the API proxy'
print('Shared static media passed for both portals')
