const crypto = require("crypto");

/**
 * Generates a random string for use as a PKCE code verifier
 * @param {number} length - Length of the code verifier (default: 43 characters)
 * @returns {string} Code verifier
 */
function generateCodeVerifier(length = 43) {
  // Generate random bytes
  const buffer = crypto.randomBytes(32);

  // Convert to base64URL
  const verifier = buffer.toString("base64url");

  // Trim to desired length and ensure it meets PKCE requirements
  return verifier.slice(0, length);
}

/**
 * Creates a code challenge from a code verifier using SHA256
 * @param {string} verifier - The code verifier
 * @returns {string} Code challenge
 */
function generateCodeChallenge(verifier) {
  // Create SHA256 hash of verifier
  const hash = crypto.createHash("sha256").update(verifier).digest();

  // Convert to base64URL
  return hash.toString("base64url");
}

/**
 * Converts a string to base64
 * @param {string} str - String to convert
 * @returns {string} Base64 encoded string
 */
function toBase64(str) {
  return Buffer.from(str).toString("base64");
}

// Example usage
function generatePKCE() {
  const codeVerifier = generateCodeVerifier();
  const codeChallenge = generateCodeChallenge(codeVerifier);
  const verifierBase64 = toBase64(codeVerifier);

  return {
    codeVerifier,
    codeChallenge,
    verifierBase64,
  };
}

// Demo
const pkce = generatePKCE();
console.log("Code Verifier:", pkce.codeVerifier);
console.log("Code Challenge:", pkce.codeChallenge);
console.log("Verifier Base64:", pkce.verifierBase64);
