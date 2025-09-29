# Rendering Nagare diagrams inside PR comments

When the `/test` endpoint produces HTML/SVG, we can surface that visual feedback in pull requests without relying on blocked `data:` URLs or unauthenticated `raw.githubusercontent.com` links. Below are a few options that work inside private repositories.

## 1. Post the raw `<svg>` markup inside the comment

GitHub's Markdown renderer allows inline SVG elements as long as the markup does not contain disallowed tags (e.g., `<script>`). You can wrap the generated SVG in a collapsible block so the comment stays short:

```markdown
<details>
  <summary>Test diagram preview</summary>

  <svg xmlns="http://www.w3.org/2000/svg" viewBox="...">
    <!-- paste the SVG from `/test` here -->
  </svg>
</details>
```

If you already capture the `/test` response in CI, pipe it through a tiny sanitizer (strip the outer HTML shell and keep only the `<svg>` element) and post the snippet in a comment. No external hosting is needed and the preview works for anyone with repo access.

## 2. Upload the PNG/SVG through `uploadCommentAttachment`

GitHub exposes a GraphQL mutation called `uploadCommentAttachment` that mirrors the drag-and-drop behaviour in the web UI. It stores the file on the `user-images.githubusercontent.com` CDN and returns a URL that you can embed in the comment body:

```bash
ATTACHMENT_URL=$(gh api graphql 
  -H 'GraphQL-Features: comment-attachment
  -F commentId="$COMMENT_NODE_ID" \
  -F name="diagram.png" \
  -F contentType="image/png" \
  -F file=@diagram.png \
  -f query='mutation($commentId: ID!, $name: String!, $contentType: String!, $file: Upload!) {
    uploadCommentAttachment(input: {commentId: $commentId, name: $name, contentType: $contentType, file: $file}) {
      attachment { downloadUrl }
    }
  }' \
  --jq '.data.uploadCommentAttachment.attachment.downloadUrl')

gh api repos/:owner/:repo/issues/:pr_number/comments \
  -f body="![diagram preview]($ATTACH
You can call the mutation from a GitHub Action (the `GITHUB_TOKEN` already has permission on the pull request) and avoid committing the artifacts anywhere in the repo. The extra `GraphQL-Features` header enables the preview API that unlocks comment attachments for CLI usage.


## 3. Publish diagrams as workflow artifacts

When a full image upload is unnecessary, add the diagram as a build artifact and include a link to the artifact in the comment. Team members can download the archive straight from the PR checks panel. This does not render inline, but it keeps the pipeline simple and avoids storing binary assets in the repository.

Choose whichever path best matches your security and automation requirements. Options (1) and (2) keep the review conversation self-contained, while (3) can act as a fallback when you simply need to provide access to the generated files.
