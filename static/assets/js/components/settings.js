PetiteVue.createApp({
    //$delimiters: ['${', '}'],  // https://github.com/vuejs/petite-vue/pull/100
    activeTab: defaultTab,
    selectedTimezone: userTimeZone,
    vibrantColorsEnabled: JSON.parse(
        localStorage.getItem("wakapi_vibrant_colors"),
    ) || false,
    labels: {},
    get tzOptions() {
        return [
            defaultTzOption,
            ...tzs.sort().map((tz) => ({ value: tz, text: tz })),
        ];
    },
    updateTab() {
        this.activeTab = window.location.hash.slice(1) || defaultTab;
    },
    isActive(tab) {
        return this.activeTab === tab;
    },
    confirmChangeUsername() {
        if (confirm("Are you sure? This cannot be undone.")) {
            document.querySelector("#form-change-username").submit();
        }
    },
    confirmRegenerate() {
        if (confirm("Are you sure?")) {
            document.querySelector("#form-regenerate-summaries").submit();
        }
    },
    confirmWakatimeImport() {
        if (confirm("Are you sure? The import can not be undone.")) {
            // weird hack to sync the "legacy importer" form field from the wakatime connection form to the (invisible) import form
            document.getElementById('use_legacy_importer').value = document.getElementById('use_legacy_importer_tmp').checked.toString()
            document.querySelector("#form-import-wakatime").submit();
        }
    },
    confirmClearData() {
        if (confirm("Are you sure? This can not be undone!")) {
            document.querySelector("#form-clear-data").submit();
        }
    },
    confirmDeleteAccount() {
        if (confirm("Are you sure? This can not be undone!")) {
            document.querySelector("#form-delete-user").submit();
        }
    },
    onToggleVibrantColors() {
        localStorage.setItem("wakapi_vibrant_colors", this.vibrantColorsEnabled);
    },
    showProjectAddButton(index) {
        this.labels[index] = true;
    },
    async webAuthnAdd() {
        const options = await fetch("/settings/webauthn/options/").then(res => res.json());
        console.log("Starting WebAuthn registration with options:", options);
        SimpleWebAuthnBrowser.startRegistration({ optionsJSON: options.publicKey })
            .then((credential) => {
                console.log("WebAuthn registration successful:", credential);
                document.getElementById(
                    "credential_json",
                ).value = JSON.stringify(credential);
                document.getElementById("form-webauthn-add").submit();
            })
            .catch((error) => {
                console.error(error);
                alert(`Error: ${error.message}`);
            });
    },
    mounted() {
        this.updateTab();
        window.addEventListener("hashchange", () => this.updateTab());
    },
}).mount("#settings-page");
