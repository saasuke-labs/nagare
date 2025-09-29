#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <diagram-path>" >&2
  exit 1
fi

diagram_path=$1
if [[ ! -f "$diagram_path" ]]; then
  echo "diagram file '$diagram_path' not found" >&2
  exit 1
fi

if [[ ! -s "$diagram_path" ]]; then
  echo "diagram file '$diagram_path' is empty" >&2
  exit 1
fi

diagram_name=$(basename "$diagram_path")

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

repo=${REPOSITORY:-${GITHUB_REPOSITORY:-}}
pr_number=${PR_NUMBER:-}

if [[ -z "${GH_TOKEN:-}" ]]; then
  if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    GH_TOKEN=$GITHUB_TOKEN
  else
    echo "GH_TOKEN is required" >&2
    exit 1
  fi
fi

declare -a missing=()
if [[ -z "$repo" ]]; then
  missing+=("REPOSITORY")
fi
if [[ -z "$pr_number" ]]; then
  missing+=("PR_NUMBER")
fi
if [[ ${#missing[@]} -gt 0 ]]; then
  printf 'missing required environment variables: %s\n' "${missing[*]}" >&2
  exit 1
fi

marker="<!-- nagare-test-diagram-preview -->"

existing_comment=$(gh api "repos/$repo/issues/$pr_number/comments" --paginate \
  --jq "map(select(.user.login == \"github-actions[bot]\" and (.body | contains(\"$marker\")))) | first")

comment_id=""
comment_node_id=""
if [[ "$existing_comment" != "null" && -n "$existing_comment" ]]; then
  comment_id=$(echo "$existing_comment" | jq -r '.id')
  comment_node_id=$(echo "$existing_comment" | jq -r '.node_id')
else
  placeholder_body=$'### Nagare /test diagram preview\n'$marker$'\n\nUploading preview…'
  new_comment=$(gh api "repos/$repo/issues/$pr_number/comments" -f body="$placeholder_body")
  comment_id=$(echo "$new_comment" | jq -r '.id')
  comment_node_id=$(echo "$new_comment" | jq -r '.node_id')
fi

if [[ -z "$comment_id" || -z "$comment_node_id" ]]; then
  echo "failed to determine comment metadata" >&2
  exit 1
fi

query=$(cat <<'Q'
mutation($commentId: ID!, $name: String!, $contentType: String!, $file: Upload!) {
  uploadCommentAttachment(input: {commentId: $commentId, name: $name, contentType: $contentType, file: $file}) {
    attachment {
      displayUrl
      downloadUrl
    }
  }
}
Q
)

response_file=$(mktemp)
comment_file=$(mktemp)
trap 'rm -f "$comment_file" "$response_file"' EXIT

operations_json=$(jq -cn --arg query "$query" --arg commentId "$comment_node_id" --arg name "$diagram_name" '{
  query: $query,
  variables: {
    commentId: $commentId,
    name: $name,
    contentType: "image/svg+xml",
    file: null
  }
}')

map_json='{"0":["variables.file"]}'

http_status=$(curl -sS \
  -H "Authorization: bearer $GH_TOKEN" \
  -H "GraphQL-Features: comment-attachments" \
  -H "Accept: application/vnd.github+json" \
  --form-string "operations=$operations_json" \
  --form-string "map=$map_json" \
  -F "0=@${diagram_path};type=image/svg+xml" \
  -w '%{http_code}' \
  -o "$response_file" \
  https://api.github.com/graphql)

if [[ "$http_status" -ge 400 ]]; then
  echo "failed to upload diagram attachment (status $http_status)" >&2
  cat "$response_file" >&2
  exit 1
fi

display_url=$(jq -e -r '.data.uploadCommentAttachment.attachment.displayUrl // empty' "$response_file")
download_url=$(jq -e -r '.data.uploadCommentAttachment.attachment.downloadUrl // empty' "$response_file")

if [[ -z "$display_url" && -z "$download_url" ]]; then
  echo "failed to parse attachment URLs" >&2
  cat "$response_file" >&2
  exit 1
fi

attachment_url=${display_url:-$download_url}

cat >"$comment_file" <<EOF_COMMENT
### Nagare /test diagram preview
$marker

<details>
<summary>View diagram</summary>

![Nagare /test preview]($attachment_url)

</details>
EOF_COMMENT

gh api "repos/$repo/issues/comments/$comment_id" -X PATCH --raw-field body="$(<"$comment_file")"
