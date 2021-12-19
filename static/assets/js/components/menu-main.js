PetiteVue.createApp({
    $delimiters: ['${', '}'],
    state: {
        showDropdownResources: false,
        showDropdownUser: false,
        showApiKey: false
    },
    mounted() {
        window.addEventListener('click', (e) => {
            const skip = findParentAttribute(e.target, 'data-trigger-for')?.value
            Object.keys(this.state).filter(k => k !== skip).forEach(k => this.state[k] = false)
        })
    }
}).mount('#main-menu')