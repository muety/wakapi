export interface PKCEResult {
  code_verifier: string;
  code_challenge: string;
  challenge_method: string;
}

export class PKCEGenerator {
  /**
   * Generates a random string for use as a PKCE code verifier
   * @param {number} length - Length of the code verifier (default: 43 characters)
   * @returns {string} Code verifier
   */
  generateCodeVerifier(length: number = 43): string {
    const randomBytes = new Uint8Array(32);
    window.crypto.getRandomValues(randomBytes);
    const verifier = this.base64urlEncode(randomBytes);
    return verifier.slice(0, length);
  }

  /**
   * Creates a code challenge from a code verifier using SHA256
   * @param {string} verifier - The code verifier
   * @returns {Promise<string>} Code challenge (as a Promise)
   */
  async generateCodeChallenge(verifier: string): Promise<string> {
    const msgBuffer = new TextEncoder().encode(verifier);
    const hashBuffer = await crypto.subtle.digest("SHA-256", msgBuffer);
    return this.base64urlEncode(new Uint8Array(hashBuffer));
  }

  /**
   * Base64url encoding function
   * @param {Uint8Array} a - Array of bytes to encode
   * @returns {string} Base64url encoded string
   */
  base64urlEncode(a: Uint8Array): string {
    let base64 = btoa(String.fromCharCode.apply(null, a as any));
    base64 = base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
    return base64;
  }

  /**
   * Generates both code verifier and code challenge
   * @returns {Promise<{codeVerifier: string, codeChallenge: string}>}
   */
  async generatePKCE(): Promise<PKCEResult> {
    const code_verifier = this.generateCodeVerifier();
    const code_challenge = await this.generateCodeChallenge(code_verifier);
    return { code_verifier, code_challenge, challenge_method: "S256" };
  }
}
