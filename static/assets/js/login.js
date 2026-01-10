document.addEventListener('DOMContentLoaded', () => {
  const webAuthnButton = document.getElementById("webauthn_login");
  webAuthnButton.onclick = () => {
    fetch("/webauthn_options/", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    })
      .then(response => response.text())
      .then(text => {
        const parsed = JSON.parse(text);
        if (parsed.error) {
          alert(parsed.error);
          throw new Error('json_error:' + parsed.error);
        }
        return parsed;
      })
      .then(optionsJSON => SimpleWebAuthnBrowser.startAuthentication({ optionsJSON: optionsJSON.publicKey }))
      .then(JSON.stringify)
      .then(assertionJSON => {
        const assertionInput = document.getElementById("webauthn_assertion_json");
        assertionInput.value = assertionJSON;
        const form = assertionInput.form;
        form.submit();
      })
      .catch(error => {
        if (!error.message.startsWith('json_error:')) {
          const errorDetail = (error && error.message) ? " Error: " + error.message : "";  
          alert("WebAuthn authentication failed. Please try again or use your password." + errorDetail);  
        }
      });
  };
});