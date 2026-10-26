#!/bin/bash
TOKEN="$1"
if [ -z "$TOKEN" ]; then
  echo "Usage: bash test-concurrent.sh TOKEN"
  exit 1
fi

echo "Mengirim 10 request serentak ke POST /portfolio/projects..."
for i in {1..10}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -X POST "http://localhost:3000/api/v1/students/1/portfolio/projects" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Serentak","technologies":["go"],"year":2026}' &
done
wait
echo "Selesai. Harus ada: 2 x 201, 8 x 409"
