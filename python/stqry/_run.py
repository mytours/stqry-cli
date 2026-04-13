import os
import subprocess
import sys


def main():
    binary = os.path.join(os.path.dirname(os.path.abspath(__file__)), "bin", "stqry")
    if sys.platform == "win32":
        binary += ".exe"
    if not os.path.isfile(binary):
        print(f"stqry: binary not found at {binary}", file=sys.stderr)
        sys.exit(1)
    sys.exit(subprocess.call([binary] + sys.argv[1:]))
