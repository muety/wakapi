PetiteVue.createApp({
    //$delimiters: ['${', '}'],  // https://github.com/vuejs/petite-vue/pull/100
    activeTab: defaultTab,
    selectedTimezone: userTimeZone,
    vibrantColorEnabled: JSON.parse(localStorage.getItem('vibrant-color') || false),
    get tzOptions() {
        return [defaultTzOption, ...tzs.sort().map(tz => ({ value: tz, text: tz }))]
    },
    updateTab() {
        this.activeTab = window.location.hash.slice(1) || defaultTab
    },
    isActive(tab) {
        return this.activeTab === tab
    },
    confirmRegenerate() {
        if (confirm('Are you sure?')) {
            document.querySelector('#form-regenerate-summaries').submit()
        }
    },
    confirmWakatimeImport() {
        if (confirm('Are you sure? The import can not be undone.')) {
            document.querySelector('#form-import-wakatime').submit()
        }
    },
    confirmClearData() {
        if (confirm('Are you sure? This can not be undone!')) {
            document.querySelector('#form-clear-data').submit()
        }
    },
    confirmDeleteAccount() {
        if (confirm('Are you sure? This can not be undone!')) {
            document.querySelector('#form-delete-user').submit()
        }
    },
    toggleVibrantColor() {
        let key = 'vibrant-color';
        let value = !(JSON.parse(localStorage.getItem(key) || false));
        localStorage.setItem(key, value);
        this.vibrantColorEnabled = value;
        return value;
    },
    mounted() {
        this.updateTab()
        window.addEventListener('hashchange', () => this.updateTab())
    }
}).mount('#settings-page')
