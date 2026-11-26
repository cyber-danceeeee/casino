#!/usr/bin/env python3
"""Create a private deployment .env once, without overwriting existing credentials."""
import base64
import os
from pathlib import Path
import secrets

root = Path(__file__).resolve().parent.parent
content = (root / '.env.example').read_text()
values = {
    'MYSQL_ROOT_PASSWORD': secrets.token_hex(24),
    'MYSQL_PASSWORD': secrets.token_hex(24),
    'RABBITMQ_PASSWORD': secrets.token_hex(24),
    'CASINO_MFA_ENCRYPTION_KEY': base64.b64encode(secrets.token_bytes(32)).decode(),
    'CASINO_BOOTSTRAP_ADMIN_PASSWORD': secrets.token_urlsafe(24),
}
for key, value in values.items():
    content = content.replace(f'{key}=\n', f'{key}={value}\n')
try:
    fd = os.open(root / '.env', os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
except FileExistsError:
    raise SystemExit('.env already exists; credentials were left unchanged.')
with os.fdopen(fd, 'w') as output:
    output.write(content)
print('Created private .env. Initial administrator: system-admin.')
print('Read CASINO_BOOTSTRAP_ADMIN_PASSWORD from .env to sign in, then bind MFA.')
