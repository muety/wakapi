PetiteVue.createApp({
    //$delimiters: ['${', '}'],  // https://github.com/vuejs/petite-vue/pull/100
    activeTab: defaultTab,
    isActive(tab) {
        return this.activeTab === tab
    },
    updateTab() {
        this.activeTab = window.location.hash.slice(1) || defaultTab
    },
    mounted() {
        this.updateTab()
        window.addEventListener('hashchange', () => this.updateTab())
    }
}).mount('#leaderboard-page')
