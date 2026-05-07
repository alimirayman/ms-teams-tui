# Azure AD App Registration Setup

This guide explains how to create your own Azure AD app registration so that teams-tui-go can authenticate with the Microsoft Graph API on your behalf.

> **Note**: The app ships with a built-in fallback client ID. You only need to follow this guide if you want to use your own app registration (recommended for production use).

---

## Steps

### 1. Go to Azure Portal

Navigate to [https://portal.azure.com](https://portal.azure.com) and sign in with your Microsoft account.

### 2. Create an App Registration

1. Search for **"App registrations"** in the top search bar and click it.
2. Click **"New registration"**.
3. Fill in:
   - **Name**: `teams-tui-go` (or any name you like)
   - **Supported account types**: Select **"Accounts in any organizational directory (Any Microsoft Entra ID tenant - Multitenant) and personal Microsoft accounts (e.g. Skype, Xbox)"**
   - **Redirect URI**: Leave blank (not needed for device code flow)
4. Click **Register**.

### 3. Enable Public Client Flow

1. In your app registration, go to **Authentication** (left sidebar).
2. Scroll down to **Advanced settings**.
3. Set **"Allow public client flows"** to **Yes**.
4. Click **Save**.

### 4. Add API Permissions

1. Go to **API permissions** (left sidebar).
2. Click **"Add a permission"** → **Microsoft Graph** → **Delegated permissions**.
3. Search for and add each of the following:
   - `User.Read`
   - `Chat.Read`
   - `Chat.ReadWrite`
   - `offline_access`
4. Click **Add permissions**.
5. (Optional) Click **"Grant admin consent"** if you have admin rights.

### 5. Copy Your Client ID

1. Go to **Overview** (left sidebar).
2. Copy the **Application (client) ID** — this is your `CLIENT_ID`.

### 6. Configure teams-tui-go

Set your client ID using either method:

**Method A — `.env` file** (in the project directory):
```bash
cp .env.example .env
# Edit .env:
CLIENT_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

**Method B — config file** (`~/.config/teams-tui-go/config.json`):
```json
{
  "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

---

## Required Permissions Summary

| Permission | Type | Purpose |
|------------|------|---------|
| `User.Read` | Delegated | Read your profile (display name, ID) |
| `Chat.Read` | Delegated | Read your Teams chats and messages |
| `Chat.ReadWrite` | Delegated | Send messages and mark chats as read |
| `offline_access` | Delegated | Get refresh tokens for silent sign-in |

---

## Troubleshooting

**"AADSTS50020: User account from identity provider does not exist in tenant"**
→ Make sure you selected **"Multitenant and personal Microsoft accounts"** in step 2, not "Single tenant".

**"AADSTS7000218: The request body must contain the following parameter: 'client_assertion' or 'client_secret'"**
→ Make sure **"Allow public client flows"** is set to **Yes** in the Authentication settings (step 3).

**Permissions not working**
→ Try signing out (delete `~/.cache/teams-tui-go/token.json`) and re-authenticating so that the new permissions take effect.
