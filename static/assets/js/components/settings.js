PetiteVue.createApp({
    //$delimiters: ['${', '}'],  // https://github.com/vuejs/petite-vue/pull/100
    activeTab: defaultTab,
    selectedTimezone: userTimeZone,
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
            formRegenerate.submit()
        }
    },
    confirmWakatimeImport() {
        if (confirm('Are you sure? The import can not be undone.')) {
            formImportWakatime.submit()
        }
    },
    confirmDeleteAccount() {
        if (confirm('Are you sure? This can not be undone!')) {
            formDelete.submit()
        }
    },
    mounted() {
        this.updateTab()
        window.addEventListener('hashchange', () => this.updateTab())
    }
}).mount('#settings-page')