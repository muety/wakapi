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
        intervalLink(interval) {
            const queryParams = new URLSearchParams(window.location.search)
            queryParams.set('interval', interval)
            return `summary?${queryParams.toString()}`
        },
        onDateUpdated() {
            document.getElementById('time-picker-form').submit()
        },
        mounted() {
            window.addEventListener('click', (e) => {
                const skip = findParentAttribute(e.target, 'data-trigger-for')?.value
                Object.keys(this.state).filter(k => k !== skip).forEach(k => this.state[k] = false)
            })

            const query = new URLSearchParams(window.location.search)
            const cookies = new URLSearchParams(document.cookie.replaceAll('; ', '&'))

            let interval = undefined;
            if (query.has('interval'))
                interval = query.get('interval')
            else if (cookies.has('wakapi_summary_interval'))
                interval = cookies.get('wakapi_summary_interval')

            if (interval) {
                const refEl = document.getElementById(`time-option-${interval}`)
                this.timeSelection = refEl ? refEl.innerText : 'Unknown'
            }
        }
    }
}