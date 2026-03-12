#!/usr/bin/env python3
"""
Read potentially malformed JSON from stdin (aichat output), extract title and
body, and print them as plain text: title on first line, blank line, then body.
Exits with 1 on failure (no output).
"""

import sys
import json
import re

text = sys.stdin.read()

m = re.search(r'\{.*\}', text, re.DOTALL)
if not m:
    sys.exit(1)

s = m.group()
obj = None

try:
    obj = json.loads(s)
except Exception:
    s2 = re.sub(r'(?<!\\)\n', r'\\n', s)
    try:
        obj = json.loads(s2)
    except Exception:
        sys.exit(1)

title = obj.get('title', '').strip()
body  = obj.get('body', '').strip()

if not title or not body:
    sys.exit(1)

print(f"{title}\n\n{body}")
