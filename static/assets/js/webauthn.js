/**
 * WebAuthn Utility Functions
 * Provides helper functions for WebAuthn credential registration and authentication
 */

class WebAuthnUtil {
    /**
     * Base64URL decode helper function
     */
    static base64urlDecode(str) {
        if (!str) {
            return new Uint8Array();
        }
        
        try {
            // Convert base64url to base64
            let base64 = str.replace(/-/g, '+').replace(/_/g, '/');
            
            // Add padding if needed
            const padding = 4 - (base64.length % 4);
            if (padding !== 4) {
                base64 += '='.repeat(padding);
            }
            
            // Decode and convert to Uint8Array
            const binaryString = atob(base64);
            const bytes = new Uint8Array(binaryString.length);
            for (let i = 0; i < binaryString.length; i++) {
                bytes[i] = binaryString.charCodeAt(i);
            }
            return bytes;
        } catch (error) {
            console.error('Base64URL decode error:', error, 'Input:', str);
            throw new Error(`Failed to decode base64url string: ${str}`);
        }
    }

    /**
     * Base64URL encode helper function
     */
    static base64urlEncode(buffer) {
        return btoa(String.fromCharCode(...new Uint8Array(buffer)))
            .replace(/\+/g, '-')
            .replace(/\//g, '_')
            .replace(/=/g, '');
    }

    /**
     * Convert server credential creation options to browser format
     */
    static parseCredentialCreationOptions(options) {
        if (options.challenge) {
            options.challenge = this.base64urlDecode(options.challenge);
        }
        if (options.user && options.user.id) {
            options.user.id = this.base64urlDecode(options.user.id);
        }
        
        if (options.excludeCredentials) {
            for (let cred of options.excludeCredentials) {
                if (cred.id) {
                    cred.id = this.base64urlDecode(cred.id);
                }
            }
        }
        
        return options;
    }

    /**
     * Convert server credential request options to browser format
     */
    static parseCredentialRequestOptions(options) {
        if (options.challenge) {
            options.challenge = this.base64urlDecode(options.challenge);
        }
        
        if (options.allowCredentials) {
            for (let cred of options.allowCredentials) {
                if (cred.id) {
                    cred.id = this.base64urlDecode(cred.id);
                }
            }
        }
        
        return options;
    }

    /**
     * Convert browser credential to server format
     */
    static formatCredential(credential) {
        return {
            id: credential.id,
            rawId: this.base64urlEncode(credential.rawId),
            type: credential.type,
            response: {
                authenticatorData: this.base64urlEncode(credential.response.authenticatorData),
                clientDataJSON: this.base64urlEncode(credential.response.clientDataJSON),
                attestationObject: credential.response.attestationObject ? 
                    this.base64urlEncode(credential.response.attestationObject) : undefined,
                signature: credential.response.signature ? 
                    this.base64urlEncode(credential.response.signature) : undefined,
                userHandle: credential.response.userHandle ? 
                    this.base64urlEncode(credential.response.userHandle) : undefined
            }
        };
    }

    /**
     * Register a new WebAuthn credential
     */
    static async register(username, credentialName) {
        try {
            // Prepare request body
            const requestBody = { username };
            if (credentialName && credentialName.trim()) {
                requestBody.credentialName = credentialName.trim();
            }

            // Begin registration
            const beginResponse = await fetch('/api/auth/webauthn/register/begin', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestBody)
            });

            if (!beginResponse.ok) {
                throw new Error(`Registration begin failed: ${beginResponse.status}`);
            }

            const response = await beginResponse.json();
            const credentialCreationOptions = this.parseCredentialCreationOptions(response.options.publicKey);

            // Create credential
            const credential = await navigator.credentials.create({
                publicKey: credentialCreationOptions
            });

            if (!credential) {
                throw new Error('Failed to create credential');
            }

            // Format credential for server
            const formattedCredential = this.formatCredential(credential);

            // Finish registration - send both credential and session data
            const finishResponse = await fetch('/api/auth/webauthn/register/finish', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    sessionData: response.sessionData,
                    response: formattedCredential
                })
            });

            if (!finishResponse.ok) {
                const error = await finishResponse.text();
                throw new Error(`Registration finish failed: ${error}`);
            }

            return await finishResponse.json();
        } catch (error) {
            console.error('WebAuthn registration error:', error);
            throw error;
        }
    }

    /**
     * Authenticate with WebAuthn
     */
    static async authenticate(username) {
        try {
            // Prepare request body - only include username if provided
            const requestBody = {};
            if (username && username.trim()) {
                requestBody.username = username.trim();
            }

            // Begin authentication
            const beginResponse = await fetch('/api/auth/webauthn/login/begin', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestBody)
            });

            if (!beginResponse.ok) {
                if (beginResponse.status === 400) {
                    throw new Error('No passkey registered for this username');
                } else if (beginResponse.status === 404) {
                    throw new Error('User not found');
                } else if (beginResponse.status === 429) {
                    throw new Error('Too many requests. Please wait a moment and try again');
                } else {
                    throw new Error(`Authentication begin failed: ${beginResponse.status}`);
                }
            }

            const response = await beginResponse.json();
            const credentialRequestOptions = this.parseCredentialRequestOptions(response.options.publicKey);

            // Get credential
            const assertion = await navigator.credentials.get({
                publicKey: credentialRequestOptions
            });

            if (!assertion) {
                throw new Error('Failed to get assertion');
            }

            // Format assertion for server
            const formattedAssertion = this.formatCredential(assertion);

            // Prepare finish request - include session data and optionally username
            const finishRequestBody = {
                sessionData: response.sessionData,
                response: formattedAssertion
            };
            
            if (username && username.trim()) {
                finishRequestBody.username = username.trim();
            }

            // Finish authentication
            const finishResponse = await fetch('/api/auth/webauthn/login/finish', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(finishRequestBody)
            });

            if (!finishResponse.ok) {
                const error = await finishResponse.text();
                throw new Error(`Authentication finish failed: ${error}`);
            }

            return await finishResponse.json();
        } catch (error) {
            console.error('WebAuthn authentication error:', error);
            throw error;
        }
    }

    /**
     * Check if WebAuthn is supported
     */
    static isSupported() {
        return 'credentials' in navigator && 
               'create' in navigator.credentials && 
               'get' in navigator.credentials &&
               window.PublicKeyCredential !== undefined;
    }

    /**
     * Show error message in UI
     */
    static showError(message, targetId = 'webauthn-error') {
        const errorDiv = document.getElementById(targetId);
        if (errorDiv) {
            errorDiv.textContent = message;
            errorDiv.style.display = 'block';
            setTimeout(() => {
                errorDiv.style.display = 'none';
            }, 5000);
        }
    }

    /**
     * Show success message in UI
     */
    static showSuccess(message, targetId = 'webauthn-success') {
        const successDiv = document.getElementById(targetId);
        if (successDiv) {
            successDiv.textContent = message;
            successDiv.style.display = 'block';
            setTimeout(() => {
                successDiv.style.display = 'none';
            }, 3000);
        }
    }
}

// Make WebAuthnUtil globally available
window.WebAuthnUtil = WebAuthnUtil;
