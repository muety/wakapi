PetiteVue.createApp({
    $delimiters: ['${', '}'],
    get currentInterval() {
        const urlParams = new URLSearchParams(window.location.search)
        return urlParams.has('interval')
            ? urlParams.get('interval')
            : null
    }
}).mount('#summary-page')