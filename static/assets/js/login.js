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
        const form = document.createElement("form");
        form.method = "POST";
        form.action = "/login_webauthn";

        const assertionInput = document.createElement("input");
        assertionInput.type = "hidden";
        assertionInput.name = "assertion_json";
        assertionInput.value = assertionJSON;
        form.appendChild(assertionInput);

        document.body.appendChild(form);
        form.submit();
      })
      .catch(error => {
        if (!error.message.startsWith('json_error:')) {
          alert("WebAuthn authentication failed. Please try again or use your password.", error);
        }
      });
  };
});