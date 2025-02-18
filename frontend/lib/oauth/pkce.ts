// import * as crypto from "crypto";

// export class PKCE {
//   private codeVerifier: string;
//   private codeChallenge: string;
//   private verifierBase64: string;
//   private codeChallengeMethod: string = "S256";

//   constructor(length: number = 43) {
//     this.codeVerifier = this.generateCodeVerifier(length);
//     this.codeChallenge = this.generateCodeChallenge(this.codeVerifier);
//     this.verifierBase64 = this.toBase64(this.codeVerifier);
//   }

//   /**
//    * Generates a random string for use as a PKCE code verifier
//    * @param {number} length - Length of the code verifier (default: 43 characters)
//    * @returns {string} Code verifier
//    */
//   private generateCodeVerifier(length: number): string {
//     // Generate random bytes
//     const buffer = crypto.randomBytes(32);

//     // Convert to base64URL
//     const verifier = buffer.toString("base64url");

//     // Trim to desired length and ensure it meets PKCE requirements
//     return verifier.slice(0, length);
//   }

//   /**
//    * Creates a code challenge from a code verifier using SHA256
//    * @param {string} verifier - The code verifier
//    * @returns {string} Code challenge
//    */
//   private generateCodeChallenge(verifier: string): string {
//     // Create SHA256 hash of verifier
//     const hash = crypto.createHash("sha256").update(verifier).digest();

//     // Convert to base64URL
//     return hash.toString("base64url");
//   }

//   /**
//    * Converts a string to base64
//    * @param {string} str - String to convert
//    * @returns {string} Base64 encoded string
//    */
//   private toBase64(str: string): string {
//     return Buffer.from(str).toString("base64");
//   }

//   /**
//    * Returns the generated PKCE values
//    * @returns {object} Object containing codeVerifier, codeChallenge, and verifierBase64
//    */
//   public getPKCE(): {
//     codeVerifier: string;
//     codeChallenge: string;
//     verifierBase64: string;
//     codeChallengeMethod: string;
//   } {
//     return {
//       codeVerifier: this.codeVerifier,
//       codeChallenge: this.codeChallenge,
//       verifierBase64: this.verifierBase64,
//       codeChallengeMethod: this.codeChallengeMethod,
//     };
//   }
// }

// // Example usage
// const pkce = new PKCE();
// const { codeVerifier, codeChallenge, verifierBase64 } = pkce.getPKCE();

// console.log("Code Verifier:", codeVerifier);
// console.log("Code Challenge:", codeChallenge);
// console.log("Verifier Base64:", verifierBase64);

export interface PKCEResult {
  codeVerifier: string;
  codeChallenge: string;
  codeChallengeMethod: string;
  verifierBase64: string;
}

export class PKCEGenerator {
  private readonly codeChallengeMethod: string;

  constructor() {
    // According to spec, S256 is the recommended and most secure method
    this.codeChallengeMethod = "S256";
  }

  /**
   * Generates random bytes and converts to base64URL string
   * @param {number} length - Length of the code verifier (default: 43 characters)
   * @returns {Promise<string>} Code verifier
   */
  async generateCodeVerifier(length: number = 43): Promise<string> {
    // Generate random bytes
    const buffer = new Uint8Array(32);
    crypto.getRandomValues(buffer);

    // Convert to base64URL
    const base64 = btoa(String.fromCharCode.apply(null, Array.from(buffer)))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=/g, "");

    // Trim to desired length and ensure it meets PKCE requirements
    return base64.slice(0, length);
  }

  /**
   * Creates a code challenge from a code verifier using SHA256
   * @param {string} verifier - The code verifier
   * @returns {Promise<string>} Code challenge
   */
  async generateCodeChallenge(verifier: string): Promise<string> {
    // Convert verifier to UTF-8 encoded buffer
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);

    // Create SHA256 hash of verifier
    const hashBuffer = await crypto.subtle.digest("SHA-256", data);

    // Convert to base64URL
    const base64 = btoa(
      //   String.fromCharCode.apply(null, new Uint8Array(hashBuffer))
      String.fromCharCode(...Array.from(new Uint8Array(hashBuffer)))
    )
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=/g, "");

    return base64;
  }

  /**
   * Converts a string to base64
   * @param {string} str - String to convert
   * @returns {string} Base64 encoded string
   */
  private toBase64(str: string): string {
    return btoa(str);
  }

  /**
   * Generates complete PKCE parameters
   * @returns {Promise<PKCEResult>} Object containing codeVerifier, codeChallenge, and verifierBase64
   */
  async generatePKCE(): Promise<PKCEResult> {
    const codeVerifier = await this.generateCodeVerifier();
    const codeChallenge = await this.generateCodeChallenge(codeVerifier);
    const verifierBase64 = this.toBase64(codeVerifier);

    return {
      codeVerifier,
      codeChallenge,
      codeChallengeMethod: this.codeChallengeMethod,
      verifierBase64,
    };
  }
}

// Example usage:
// const pkceGenerator = new PKCEGenerator();
// pkceGenerator.generatePKCE().then((pkce: PKCEResult) => {
//   console.log("Code Verifier:", pkce.codeVerifier);
//   console.log("Code Challenge:", pkce.codeChallenge);
//   console.log("Challenge Method:", pkce.codeChallengeMethod);
//   console.log("Verifier Base64:", pkce.verifierBase64);
// });
// Example usage:
// const pkceGenerator = new PKCEGenerator();
// pkceGenerator.generatePKCE().then(pkce => {
//   console.log("Code Verifier:", pkce.codeVerifier);
//   console.log("Code Challenge:", pkce.codeChallenge);
//   console.log("Challenge Method:", pkce.codeChallengeMethod);
//   console.log("Verifier Base64:", pkce.verifierBase64);
// });
