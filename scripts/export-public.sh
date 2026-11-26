#!/bin/sh
# Make a fresh copy of this repository for publishing: the files of the
# current commit and nothing of its history, as one commit under the name and
# address you give. It pushes nothing and changes nothing here.
#
# usage: scripts/export-public.sh <new directory> "<author name>" <author email>
#   for example a GitHub no-reply address: <id>+<login>@users.noreply.github.com
set -eu

if [ $# -ne 3 ]; then
	sed -n '2,8p' "$0" | sed 's/^# \{0,1\}//'
	exit 2
fi
dst=$1 name=$2 email=$3

cd "$(git rev-parse --show-toplevel)"
if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
	echo "there are uncommitted changes to tracked files; commit them first" >&2
	exit 1
fi
if [ -e "$dst" ] && [ -n "$(ls -A "$dst" 2>/dev/null)" ]; then
	echo "$dst exists and is not empty" >&2
	exit 1
fi

# Things that must not be published. Each check names files only.
fail=0
check() { # check <what> <extended regex>
	found=$(git grep -lIE -e "$2" -- . ':!scripts/export-public.sh' || true)
	if [ -n "$found" ]; then
		echo "refusing to export: $1 in:" >&2
		echo "$found" | sed 's/^/  /' >&2
		fail=1
	fi
}
check "a home directory path" '/(Users|home)/[A-Za-z0-9._-]{2,}/'
check "an email address other than a placeholder" '[A-Za-z0-9._%+-]+@([A-Za-z0-9-]+\.)+(edu|com|org|net|io|dev)([^A-Za-z]|$)'
check "something shaped like a key or token" '(sk-[A-Za-z0-9_-]{24,}|AKIA[0-9A-Z]{16}|gh[pous]_[A-Za-z0-9]{30,}|AIza[0-9A-Za-z_-]{30,}|BEGIN [A-Z ]*PRIVATE KEY)'
[ $fail -eq 0 ] || exit 1
for f in LICENSE NOTICE README.md SECURITY.md; do
	[ -f "$f" ] || { echo "refusing to export: $f is missing" >&2; exit 1; }
done

mkdir -p "$dst"
git archive HEAD | tar -x -C "$dst"
cd "$dst"
git init -q -b main
git add -A
git -c user.name="$name" -c user.email="$email" commit -q -m "Initial public release"
echo "exported $(git ls-files | wc -l | tr -d ' ') files to $dst as one commit by $name <$email>"
echo "nothing was pushed. To publish: create an empty repository on GitHub, then"
echo "  git -C $dst remote add origin <its url> && git -C $dst push -u origin main"
