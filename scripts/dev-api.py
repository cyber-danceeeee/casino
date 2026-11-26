#!/usr/bin/env python3
"""Start the local Go API using the private root .env without invoking a shell."""
import argparse
import os
from pathlib import Path
import re
import shlex

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check-config', action='store_true',
                    help='Validate configuration without connecting to business services.')
args = parser.parse_args()
root = Path(__file__).resolve().parent.parent
environment = os.environ.copy()
env_file = root / '.env'
if env_file.exists():
    for number, line in enumerate(env_file.read_text().splitlines(), 1):
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        key, separator, value = line.partition('=')
        key, value = key.strip(), value.strip()
        if not separator or not re.fullmatch(r'[A-Za-z_][A-Za-z_0-9]*', key):
            raise SystemExit(f'Invalid .env assignment on line {number}; contents omitted.')
        if value.startswith(('"', "'")):
            try:
                parts = shlex.split(value, comments=True, posix=True)
                if len(parts) != 1:
                    raise ValueError()
                value = parts[0]
            except ValueError:
                raise SystemExit(f'Invalid quoted .env value on line {number}; contents omitted.')
        else:
            value = value.split(' #', 1)[0].rstrip()
        # Explicit process environment wins, including empty optional values.
        environment.setdefault(key, value)
environment.setdefault('GOTOOLCHAIN', 'go1.24.13')
target = './cmd/check-config' if args.check_config else './cmd/casino'
os.execvpe('go', ['go', '-C', str(root / 'server'), 'run', target,
                 '-c', 'configs/casino.yml'], environment)
