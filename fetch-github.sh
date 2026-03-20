#!/bin/sh
# Fetch all GitHub events (paginated, up to 10 pages) into data/github_events.json
# If the API fails, keeps the existing cached file.

USER="azaidrahman"
OUT="data/github_events.json"
TMPDIR_PAGES=$(mktemp -d)

mkdir -p data

page=1
max_pages=10

while [ "$page" -le "$max_pages" ]; do
  pagefile="$TMPDIR_PAGES/page${page}.json"
  status=$(curl -s -o "$pagefile" -w "%{http_code}" \
    "https://api.github.com/users/${USER}/events?per_page=100&page=${page}")

  if [ "$status" != "200" ]; then
    rm -f "$pagefile"
    if [ "$page" -eq 1 ]; then
      rm -rf "$TMPDIR_PAGES"
      if [ -f "$OUT" ]; then
        echo "API returned $status — using cached data."
      else
        echo "[]" > "$OUT"
        echo "API returned $status — no cache, wrote empty array."
      fi
      exit 0
    fi
    break
  fi

  count=$(python3 -c "import json; print(len(json.load(open('$pagefile'))))")
  if [ "$count" -lt 100 ]; then
    page=$((page + 1))
    break
  fi

  page=$((page + 1))
done

# Merge all page files into one array
python3 -c "
import json, glob, os
pages = sorted(glob.glob('$TMPDIR_PAGES/page*.json'))
all_events = []
for p in pages:
    all_events.extend(json.load(open(p)))
json.dump(all_events, open('$OUT', 'w'))
print(f'GitHub events updated — {len(all_events)} events across {len(pages)} page(s).')
"

rm -rf "$TMPDIR_PAGES"
