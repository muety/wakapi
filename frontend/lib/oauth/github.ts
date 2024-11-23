function makeState(payload: Record<string, any>): string {
  return Buffer.from(JSON.stringify(payload)).toString("base64");
}

export function startGithubLoginFlow() {
  const { clientId, redirectUri, scope } = getGithubConfig();

  const state_payload = {
    redirectUri,
    scope,
    clientId,
  };
  const state = makeState(state_payload);
  const url = `https://github.com/login/oauth/authorize?client_id=${clientId}&redirect_uri=${redirectUri}&scope=${scope}&state=${state}`;
  return url;
}

export function getGithubConfig() {
  const clientId = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID;
  const redirectUri = process.env.NEXT_PUBLIC_GITHUB_REDIRECT_URI;
  const scope = "user:email read:user";
  return { clientId, redirectUri, scope };
}
