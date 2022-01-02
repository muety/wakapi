function TimePicker({ fromDate, toDate, timeSelection }) {
    return {
        $template: '#time-picker-template',
        $delimiters: ['${', '}'],
        state: {
            showDropdownTimepicker: false
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
    }
}