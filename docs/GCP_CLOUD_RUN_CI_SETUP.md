# GCP Cloud Run CI Setup (GitHub Actions + Workload Identity Federation)

This guide fixes the CI error:

> `Permission 'iam.serviceAccounts.getAccessToken' denied`

That error means the GitHub OIDC principal is authenticated, but it is **not allowed to impersonate** the target service account to mint an access token.

---

## Why this happens

`google-github-actions/auth@v2` with `token_format: access_token` calls IAM Credentials API to generate an OAuth2 token for your deployer service account.

For that to work, the GitHub principal from your Workload Identity Provider must have:

- `roles/iam.workloadIdentityUser` **on the service account** (required)
- Workload Identity Provider configured with matching attribute mapping/condition (required)

In some org setups, adding `roles/iam.serviceAccountTokenCreator` to the same principal can also be required for access-token minting flows.

Without the first item, CI fails with `iam.serviceAccounts.getAccessToken` denied.

---

## Prerequisites

- A Google Cloud project with billing enabled
- Repo admin access (to set GitHub Actions secrets)
- `gcloud` installed and authenticated locally (for CLI path)

Set these variables for CLI commands:

```bash
export PROJECT_ID="<your-project-id>"
export PROJECT_NUMBER="$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')"
export REGION="us-central1"
export REPO_OWNER="<github-owner>"
export REPO_NAME="nagare"
export REPO_FULL="${REPO_OWNER}/${REPO_NAME}"
export POOL_ID="github"
export PROVIDER_ID="github"
export SA_NAME="nagare-deployer"
export SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
```

---

## Option A: Configure everything with gcloud CLI

### 1) Enable required APIs

```bash
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  iamcredentials.googleapis.com \
  iam.googleapis.com \
  sts.googleapis.com \
  --project "$PROJECT_ID"
```

### 2) Create Artifact Registry repository

```bash
gcloud artifacts repositories create nagare \
  --repository-format=docker \
  --location="$REGION" \
  --project "$PROJECT_ID"
```

If it already exists, this command can fail safely.

### 3) Create deployer service account

```bash
gcloud iam service-accounts create "$SA_NAME" \
  --project "$PROJECT_ID" \
  --display-name="Nagare GitHub deployer"
```

### 4) Grant project roles to deployer service account

```bash
for role in roles/run.admin roles/artifactregistry.writer roles/iam.serviceAccountUser; do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="$role"
done
```

### 5) Create Workload Identity Pool

```bash
gcloud iam workload-identity-pools create "$POOL_ID" \
  --project="$PROJECT_ID" \
  --location="global" \
  --display-name="GitHub Actions Pool"
```

### 6) Create Workload Identity Provider (GitHub OIDC)

```bash
gcloud iam workload-identity-pools providers create-oidc "$PROVIDER_ID" \
  --project="$PROJECT_ID" \
  --location="global" \
  --workload-identity-pool="$POOL_ID" \
  --display-name="GitHub Provider" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.ref=assertion.ref,attribute.actor=assertion.actor" \
  --attribute-condition="assertion.repository=='${REPO_FULL}'"
```

### 7) **Critical fix**: allow WIF principal to impersonate the service account

This is the step that fixes your current CI error.

```bash
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/attribute.repository/${REPO_FULL}"

# Optional fallback if your org/policy still blocks access token minting:
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$PROJECT_ID" \
  --role="roles/iam.serviceAccountTokenCreator" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/attribute.repository/${REPO_FULL}"
```

Optional stricter binding for `main` only:

```bash
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --project="$PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/attribute.ref/refs/heads/main"
```

### 8) Collect values for GitHub secrets

```bash
WIP_RESOURCE="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/providers/${PROVIDER_ID}"
echo "$WIP_RESOURCE"
echo "$SA_EMAIL"
```

Set these repo secrets:

- `GCP_PROJECT_ID` = your project ID
- `GCP_WORKLOAD_IDENTITY_PROVIDER` = output of `WIP_RESOURCE`
- `GCP_SERVICE_ACCOUNT` = `SA_EMAIL`

### 9) Validate before pushing

```bash
gcloud iam service-accounts get-iam-policy "$SA_EMAIL" \
  --project="$PROJECT_ID" \
  --format=json
```

Confirm you see `roles/iam.workloadIdentityUser` with your `principalSet://...` member.

Also verify the project number in the `principalSet://...` member is the **numeric project number** that owns the Workload Identity Pool (not `PROJECT_ID`). A mismatched project number is the most common reason this still fails after setup.

---

## Option B: Configure with GCP Console (UI)

### 1) Enable APIs

- Go to **APIs & Services → Library**.
- Enable:
  - Cloud Run API
  - Artifact Registry API
  - IAM Service Account Credentials API
  - Security Token Service API
  - IAM API

### 2) Create Artifact Registry repo

- Go to **Artifact Registry → Repositories → Create Repository**.
- Name: `nagare`
- Format: `Docker`
- Region: `us-central1` (or your workflow region)

### 3) Create deployer service account

- Go to **IAM & Admin → Service Accounts → Create Service Account**.
- Name: `nagare-deployer`

### 4) Grant project roles to the service account

In **IAM & Admin → IAM**, grant to `nagare-deployer@...`:

- `Cloud Run Admin` (`roles/run.admin`)
- `Artifact Registry Writer` (`roles/artifactregistry.writer`)
- `Service Account User` (`roles/iam.serviceAccountUser`)

### 5) Create Workload Identity Pool + Provider

- Go to **IAM & Admin → Workload Identity Federation**.
- Create Pool: e.g. `github`
- Create Provider in that pool:
  - Provider type: **OIDC**
  - Issuer URL: `https://token.actions.githubusercontent.com`
  - Audience: default (Google)
  - Attribute mapping:
    - `google.subject` -> `assertion.sub`
    - `attribute.repository` -> `assertion.repository`
    - `attribute.ref` -> `assertion.ref`
  - Attribute condition:
    - `assertion.repository=='<OWNER>/<REPO>'`

### 6) **Critical fix in UI**: grant Workload Identity User on SA

- Open **IAM & Admin → Service Accounts → nagare-deployer → Permissions**.
- Click **Grant access**.
- Principal:
  - `principalSet://iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github/attribute.repository/<OWNER>/<REPO>`
- Role:
  - `Workload Identity User` (`roles/iam.workloadIdentityUser`)

This is the UI equivalent of the CLI fix command.

If CI still fails, also grant `Service Account Token Creator` (`roles/iam.serviceAccountTokenCreator`) to the same `principalSet://...` principal on the same service account.

### 7) Configure GitHub secrets

In GitHub: **Repo → Settings → Secrets and variables → Actions → New repository secret**

- `GCP_PROJECT_ID`
- `GCP_WORKLOAD_IDENTITY_PROVIDER` (from provider details page, full resource name)
- `GCP_SERVICE_ACCOUNT` (service account email)

---

## Quick troubleshooting checklist

If deploy still fails:

1. Verify workflow has `permissions: id-token: write` (required for OIDC).
2. Verify `GCP_WORKLOAD_IDENTITY_PROVIDER` matches exact provider resource path.
3. Verify `GCP_SERVICE_ACCOUNT` exists and is correct.
4. Verify SA IAM policy includes `roles/iam.workloadIdentityUser` for the expected `principalSet://...`.
5. Verify the `principalSet://.../projects/<PROJECT_NUMBER>/...` uses the correct **pool project number**.
6. Verify provider condition matches repo and (if used) branch.
7. If still denied, grant `roles/iam.serviceAccountTokenCreator` on the service account to the same principal.
8. Verify required APIs are enabled.

---

## Notes for this repository

The workflow file is `.github/workflows/deploy-gcp.yml` and already uses Workload Identity Federation. If you see the access token permission error, the missing piece is almost always step **"grant Workload Identity User on the service account"**.
