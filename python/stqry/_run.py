import os
import subprocess
import sys


def main():
    binary = os.path.join(os.path.dirname(os.path.abspath(__file__)), "bin", "stqry")
    if sys.platform == "win32":
        binary += ".exe"
    sys.exit(subprocess.call([binary] + sys.argv[1:]))
