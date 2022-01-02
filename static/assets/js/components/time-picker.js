// To use this, do the following:
// 1. Include petite-vue
// 2. Define variables timeSelection, fromDate, toDate in an inline script below previous
// 3. Include this script file below previous
// 4. Include time-picker.tpl.html template partial in body

PetiteVue.createApp({
    $delimiters: ['${', '}'],
    state: {
        showDropdownTimepicker: false,
    },
    fromDate: fromDate,
    toDate: toDate,
    timeSelection: timeSelection,
    onDateUpdated() {
        document.getElementById('time-picker-form').submit()
    },
    mounted() {
        window.addEventListener('click', (e) => {
            const skip = findParentAttribute(e.target, 'data-trigger-for')?.value
            Object.keys(this.state).filter(k => k !== skip).forEach(k => this.state[k] = false)
        })

        const query = new URLSearchParams(window.location.search)
        if (query.has('interval')) {
            const refEl = document.getElementById(`time-option-${query.get('interval')}`)
            this.timeSelection = refEl ? refEl.innerText : 'Unknown'
        }
    }
}).mount('#time-picker-container')