"""Market and independent supplier account checks for a disposable deployment."""
import json
import hashlib
import time
from urllib.parse import urlencode, quote
import os
import shlex
import subprocess
from urllib.request import Request, urlopen
from urllib.error import HTTPError


def verify_provider_market(call, h5_base, admin_token, player_token):
    compose = shlex.split(os.getenv('CASINO_COMPOSE', 'docker compose'))

    def sql(statement):
        command = compose + ['exec', '-T', 'mysql', 'sh', '-c',
                             'MYSQL_PWD="$MYSQL_PASSWORD" mysql --default-character-set=utf8mb4 -u casino game_casino --batch --skip-column-names']
        result = subprocess.run(command, input=statement, text=True,
                                stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        assert result.returncode == 0, result.stderr
        return result.stdout

    def request(path, body=None, token=player_token, method=None):
        req = Request(h5_base + path, data=json.dumps(body).encode() if body is not None else None,
                      method=method or ('POST' if body is not None else 'GET'),
                      headers={'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token})
        try:
            with urlopen(req, timeout=15) as response:
                return response.status, json.load(response)
        except HTTPError as error:
            return error.code, json.load(error)

    forbidden_status, forbidden = request('/admin/game-provider-connections')
    assert forbidden_status in (401, 403) or forbidden['code'] in (401, 403)
    assert 'list' not in (forbidden.get('data') or {})
    before = call('/admin/game-provider-connections')
    assert before['market'] == {'countryCode': 'PK', 'currencyCode': 'PKR'}
    assert before['countries'] and len(before['adapters']) == 2
    initial_policy = call('/app/config', base=h5_base)['resourceVersion']
    connections = {}
    for code, country, currency in [('accept-pk', 'PK', 'PKR'), ('accept-ca', 'CA', 'CAD')]:
        body = dict(providerCode=code, adapterType='softaggregator', name='Acceptance ' + country,
                    apiBaseUrl='https://supplier-test.invalid', environment='sandbox', enabled=True,
                    supportedCountries=country, supportedCurrencies=currency, callbackIPAllowlist=[],
                    version=0, credentials={'apiLogin': 'accept-login-' + code,
                                            'apiPassword': 'accept-password-' + code,
                                            'saltKey': 'accept-signing-' + code}, reason='disposable market acceptance')
        result = call('/admin/game-provider-connections/' + code, body, method='PUT')
        assert result['version'] == 1 and result['credentialsConfigured']
        assert result['marketAllowed'] == (code == 'accept-pk')
        assert all(value not in json.dumps(result) for value in body['credentials'].values())
        connections[code] = body
    assert len(call('/admin/game-provider-connections')['list']) >= 4
    assert call('/app/config', base=h5_base)['resourceVersion'] != initial_policy
    assert call('/bootstrap?country=CA', base=h5_base)['country']['code'] == 'PK'
    home = call('/home', base=h5_base)
    assert [row['providerCode'] for row in home['providers']] == ['accept-pk']
    ciphertext = sql("SELECT credentials_ciphertext FROM game_provider_connections WHERE provider_code='accept-pk';")
    assert 'accept-password-' not in ciphertext and 'accept-signing-' not in ciphertext
    # Blank password fields retain existing encrypted values, and versions prevent lost edits.
    update = dict(connections['accept-pk'], version=1, credentials={'apiPassword': ''})
    saved = call('/admin/game-provider-connections/accept-pk', update, method='PUT')
    assert saved['version'] == 2 and saved['credentialsConfigured']
    _, conflict = request('/admin/game-provider-connections/accept-pk', update,
                          token=admin_token, method='PUT')
    assert conflict['code'] != 0 and conflict['data']['errorCode'] == 'PROVIDER_CONFIG_CONFLICT'
    for game_id, code, country, currency in [(950001, 'accept-pk', 'PK', 'PKR'),
                                            (950002, 'accept-ca', 'PK', 'PKR'),
                                            (950003, 'accept-pk', 'CA', 'PKR'),
                                            (950004, 'accept-pk', 'PK', 'CAD')]:
        upstream_code = ('studio/' if game_id == 950004 else '') + f'accept-{game_id}'
        sql(f"""INSERT INTO game_games(id,site_id,game_code,provider_id,provider_game_code,name,slug,game_type,
                 provider_enabled,platform_enabled,status,supported_countries,supported_currencies,
                 supports_demo,name_i18n,raw_metadata)
                 SELECT {game_id},1,'accept-{game_id}',id,'{upstream_code}','Acceptance Game','accept-{game_id}','SLOT',
                 1,1,1,'{country}','{currency}',1,'{{}}','{{}}'
                 FROM game_game_providers WHERE provider_code='{code}' AND site_id=0 AND deleted_at IS NULL;""")
    public = call('/games?provider=accept-pk', base=h5_base)
    assert [row['id'] for row in public['list']] == [950001]
    forged = call('/games?provider=accept-pk&country=CA&currency=CAD', base=h5_base,
                  headers={'X-Country-Code': 'CA', 'X-Currency-Code': 'CAD'})
    assert [row['id'] for row in forged['list']] == [950001]
    assert not call('/games?provider=accept-ca&country=CA&currency=CAD', base=h5_base)['list']
    assert not call('/game/providers/accept-ca/gamelist?enabled=false', base=h5_base)['list']
    lobby = call('/lobby?country=CA&currency=CAD', base=h5_base)
    assert 'accept-ca' not in json.dumps(lobby) and 'accept-pk' in json.dumps(lobby)
    for game_id in [950002, 950003, 950004]:
        _, detail = request('/games/' + str(game_id))
        assert detail['code'] != 0
        _, legacy_launch = request('/games/' + str(game_id) + '/launch', {})
        assert legacy_launch['code'] != 0
        code = 'accept-ca' if game_id == 950002 else 'accept-pk'
        upstream_code = ('studio/' if game_id == 950004 else '') + f'accept-{game_id}'
        _, launch = request(f'/games/{code}/{quote(upstream_code, safe="")}/launch/demo', {'country': 'PK', 'currency': 'PKR'})
        assert launch['code'] != 0 and launch['data']['errorCode'] == 'MARKET_UNAVAILABLE'
    # Even an eligible game cannot use a caller-selected country or currency.
    _, spoofed = request('/games/accept-pk/accept-950001/launch/demo', {'country': 'CA', 'currency': 'CAD'})
    assert spoofed['code'] != 0 and spoofed['data']['errorCode'] == 'MARKET_UNAVAILABLE'
    disabled = dict(connections['accept-pk'], version=2, enabled=False, credentials={})
    call('/admin/game-provider-connections/accept-pk', disabled, method='PUT')
    assert not call('/games?provider=accept-pk', base=h5_base)['list']
    # Callback parsing remains registered after stopping new games, but unsigned callbacks fail.
    status, callback = request('/game/providers/accept-pk/callback', {}, token='')
    assert status != 404 and callback['error'] != 0
    player_id = int(sql("SELECT id FROM game_users ORDER BY id DESC LIMIT 1;").strip())
    sql(f"INSERT INTO game_provider_players(site_id,provider_code,user_id,provider_player_id,provider_player_password,currency,status) VALUES(1,'accept-pk',{player_id},'accept-player','unused-test-password','PKR',1);")
    timestamp = str(int(time.time()))
    signed = urlencode({'action': 'balance', 'username': 'accept-player', 'currency': 'PKR',
                        'timestamp': timestamp, 'key': hashlib.md5((timestamp + 'accept-signing-accept-pk').encode()).hexdigest()})
    _, callback = request('/game/providers/accept-pk/callback?' + signed, token='')
    assert callback['error'] == 0 and callback['balance'] >= 0
    print('Supplier account encryption, live reload, fixed-market visibility, launch rejection and callback retention passed.')
