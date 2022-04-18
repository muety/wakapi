PetiteVue.createApp({
    $delimiters: ['${', '}'],
    get currentInterval() {
        const urlParams = new URLSearchParams(window.location.search)
        if (urlParams.has('interval')) return urlParams.get('interval')
        if (!urlParams.has('from') && !urlParams.has('to')) return 'today'
        return null
    }
}).mount('#summary-page')