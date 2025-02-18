// class PKCEGenerator {
//   /**
//    * Generates a random string for use as a PKCE code verifier
//    * @param {number} length - Length of the code verifier (default: 43 characters)
//    * @returns {string} Code verifier
//    */
//   generateCodeVerifier(length = 43) {
//     const randomBytes = new Uint8Array(32);
//     window.crypto.getRandomValues(randomBytes);
//     const verifier = this.base64urlEncode(randomBytes);
//     return verifier.slice(0, length);
//   }

//   /**
//    * Creates a code challenge from a code verifier using SHA256
//    * @param {string} verifier - The code verifier
//    * @returns {Promise<string>} Code challenge (as a Promise)
//    */
//   async generateCodeChallenge(verifier: string) {
//     const msgBuffer = new TextEncoder().encode(verifier);
//     const hashBuffer = await crypto.subtle.digest("SHA-256", msgBuffer);
//     return this.base64urlEncode(new Uint8Array(hashBuffer));
//   }

//   /**
//    * Base64url encoding function
//    * @param {Uint8Array} a - Array of bytes to encode
//    * @returns {string} Base64url encoded string
//    */
//   base64urlEncode(a) {
//     let base64 = btoa(String.fromCharCode.apply(null, a));
//     base64 = base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
//     return base64;
//   }

//   /**
//    * Generates both code verifier and code challenge
//    * @returns {Promise<{codeVerifier: string, codeChallenge: string}>}
//    */
//   async generatePKCE() {
//     const codeVerifier = this.generateCodeVerifier();
//     const codeChallenge = await this.generateCodeChallenge(codeVerifier);
//     return { codeVerifier, codeChallenge };
//   }
// }

// // Example usage:
// const pkceGen = new PKCEGenerator();

// async function demo() {
//   // Must be in an async function
//   const pkce = await pkceGen.generatePKCE();
//   console.log("Code Verifier:", pkce.codeVerifier);
//   console.log("Code Challenge:", pkce.codeChallenge);

//   // Example of using the class methods individually:
//   const verifier = pkceGen.generateCodeVerifier();
//   const challenge = await pkceGen.generateCodeChallenge(verifier);
//   console.log("Verifier (individual):", verifier);
//   console.log("Challenge (individual):", challenge);
// }

// demo(); // Call the async demo function
