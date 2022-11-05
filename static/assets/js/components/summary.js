PetiteVue.createApp({
    $delimiters: ['${', '}'],
    get currentInterval() {
        const urlParams = new URLSearchParams(window.location.search)
        const cookies = new URLSearchParams(document.cookie.replaceAll('; ', '&'))

        if (urlParams.has('interval'))
            return urlParams.get('interval')
        if (cookies.has('wakapi_summary_interval'))
            return cookies.get('wakapi_summary_interval')
        if (!urlParams.has('from') && !urlParams.has('to'))
            return 'today'

        return null
    }
}).mount('#summary-page')