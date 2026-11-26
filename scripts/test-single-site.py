#!/usr/bin/env python3
"""Exercise a fresh disposable deployment, including MFA and H5 publication."""
import argparse
import base64
import hashlib
import hmac
import json
import os
from pathlib import Path
import secrets
import struct
import time
from urllib.error import HTTPError
from urllib.request import Request, urlopen

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--disposable', action='store_true', required=True,
                    help='Confirm this is a fresh test database; its account and H5 settings will change.')
parser.parse_args()

admin_port = os.getenv('ADMIN_PORT', '18081')
h5_port = os.getenv('H5_PORT', '18080')
admin_base = f'http://127.0.0.1:{admin_port}/api/v1'
h5_base = f'http://127.0.0.1:{h5_port}/api/v1'
root = Path(__file__).resolve().parent.parent
settings = dict(line.split('=', 1) for line in (root / '.env').read_text().splitlines()
                if line and not line.startswith('#'))
token = ''


def call(path, body=None, method=None, base=admin_base, headers=None):
    request_headers = {'Content-Type': 'application/json'}
    if token:
        request_headers['Authorization'] = 'Bearer ' + token
    request_headers.update(headers or {})
    request = Request(base + path, data=json.dumps(body).encode() if body is not None else None,
                      headers=request_headers, method=method or ('POST' if body is not None else 'GET'))
    with urlopen(request, timeout=30) as response:
        result = json.load(response)
    assert result['code'] == 0, f'{path}: {result.get("msg", "API error")}'
    return result['data']


config = call('/app/config', base=h5_base)
assert config['siteId'] > 0 and config['currency']['code']
other_scope = call('/app/config', base=h5_base,
                   headers={'X-Tenant-ID': '99999', 'X-Site-ID': '99999'})
assert other_scope['siteId'] == config['siteId']

login = call('/admin/auth/login', {
    'username': 'system-admin', 'password': settings['CASINO_BOOTSTRAP_ADMIN_PASSWORD'],
})
assert login['status'] == 'MFA_SETUP_REQUIRED', 'This test requires a fresh disposable deployment.'
assert login['administrator']['adminType'] == 'SITE'
secret = base64.b32decode(login['secret'] + '=' * ((-len(login['secret'])) % 8))
digest = hmac.new(secret, struct.pack('>Q', int(time.time()) // 30), hashlib.sha1).digest()
offset = digest[-1] & 15
code = f'{(struct.unpack(">I", digest[offset:offset + 4])[0] & 0x7fffffff) % 1000000:06d}'
session = call('/admin/auth/mfa/setup/confirm', {
    'challengeToken': login['challengeToken'], 'code': code,
})
token = session['token']
password = 'SiteControl9!' + secrets.token_hex(10)
call('/admin/account/password', {
    'currentPassword': settings['CASINO_BOOTSTRAP_ADMIN_PASSWORD'],
    'newPassword': password, 'confirmPassword': password,
})
actor = call('/admin/auth/me')
assert actor['adminType'] == 'SITE' and not actor['platform']
assert actor['siteID'] == config['siteId']

codes = {item['code'] for item in call('/admin/system/navigation')['list']}
assert {'brand-config', 'user-list', 'game-sync', 'system-admins', 'system-roles'}.issubset(codes)
assert not {'tenants', 'sites', 'tenant-packages', 'create-brand'}.intersection(codes)
for path in ['/admin/system/tenants', '/admin/system/tenant-packages',
             '/admin/system/platform/dashboard']:
    try:
        call(path)
        raise AssertionError('A SaaS management API is still reachable')
    except HTTPError as error:
        assert error.code == 404
for path in ['/admin/system/admins', '/admin/system/roles', '/admin/system/countries',
             '/admin/game-providers', '/admin/games']:
    call(path)
assert call('/admin/system/h5/markets')['list']

site = call('/admin/system/brand/overview')
payload = dict(site['draftConfig'])
payload.update(siteName='Single H5 Acceptance', h5Domain=f'http://127.0.0.1:{h5_port}',
               reason='single-site integration test')
payload['brandConfig'] = dict(payload['brandConfig'], appName='Single H5 Acceptance')
payload['themeConfig'] = dict(payload['themeConfig'], primary='#12ab34', button='#12ab34')
payload['seoConfig'] = dict(payload.get('seoConfig') or {}, title='Single H5 Acceptance',
                            description='Single-site deployment verification')
call('/admin/system/h5/settings', payload, method='PUT')
assert call('/app/config', base=h5_base)['branding']['theme'].get('primaryColor') != '#12ab34'
call(f'/admin/system/sites/{site["siteID"]}/config/publish', {'reason': 'single-site integration test'})
published = call('/app/config', base=h5_base)
assert published['siteId'] == site['siteID']
assert published['branding']['theme']['primaryColor'] == '#12ab34'
assert published['branding']['appName'] == 'Single H5 Acceptance'
assert published['seo']['title'] == 'Single H5 Acceptance'
print('Single-site initialization, login/MFA, permissions, SaaS API removal, draft isolation and H5 publication passed.')
