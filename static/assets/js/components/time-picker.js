function TimePicker({fromDate, toDate, timeSelection}) {
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
        // Step to the previous (-1) or next (+1) period. `toDate` is exclusive (midnight after the range).
        // Named periods snap to the full previous / next period of their nominal length, anchored at their start, so a partial "this week" steps to the whole week before.
        // Custom ranges spanning whole calendar months step by months, all others by their length in days.
        shiftedRange(direction) {
            const parse = (s) => {
                const [y, m, d] = s.split('-').map(Number)
                return new Date(Date.UTC(y, m - 1, d))
            }
            const format = (date) => date.toISOString().slice(0, 10)
            const add = (date, unit, n) => {
                const d = new Date(date)
                if (unit === 'months') d.setUTCMonth(d.getUTCMonth() + n)
                else d.setUTCDate(d.getUTCDate() + n)
                return d
            }

            const from = parse(this.fromDate)
            const to = parse(this.toDate)
            if (isNaN(from) || isNaN(to) || to <= from) return null

            // keys mirror the aliases in models/interval.go
            const namedSteps = {
                'today': ['days', 1],
                'day': ['days', 1],
                'yesterday': ['days', 1],
                '24_hours': ['days', 1],
                'last_24_hours': ['days', 1],
                'last_day': ['days', 1],
                'week': ['days', 7],
                'last_week': ['days', 7],
                '7_days': ['days', 7],
                'last_7_days': ['days', 7],
                '14_days': ['days', 14],
                'last_14_days': ['days', 14],
                '30_days': ['days', 30],
                'last_30_days': ['days', 30],
                'month': ['months', 1],
                'last_month': ['months', 1],
                '6_months': ['months', 6],
                'last_6_months': ['months', 6],
                'year': ['months', 12],
                '12_months': ['months', 12],
                'last_12_months': ['months', 12],
                'last_year': ['months', 12],
            }
            const params = new URLSearchParams(window.location.search)
            const interval = params.get('interval') || params.get('start')  // start is a server-side alias for interval (see helpers.ParseSummaryParams)
            if (['any', 'all_time', 'All Time'].includes(interval)) return null  // no shifting for all time interval

            let step = namedSteps[interval]
            if (!step) {
                if (from.getUTCDate() === 1 && to.getUTCDate() === 1) {
                    step = ['months', (to.getUTCFullYear() - from.getUTCFullYear()) * 12 + (to.getUTCMonth() - from.getUTCMonth())]
                } else {
                    step = ['days', Math.round((to - from) / 86400000)]
                }
            }

            const [unit, size] = step
            const newFrom = add(from, unit, direction * size)
            return {from: format(newFrom), to: format(add(newFrom, unit, size))}
        },
        canShift(direction) {
            const range = this.shiftedRange(direction)
            if (!range) return false
            if (direction < 0) return true
            const now = new Date()
            const today = new Date(Date.UTC(now.getFullYear(), now.getMonth(), now.getDate())).toISOString().slice(0, 10)
            return range.from <= today // don't navigate into ranges that start in the future
        },
        shiftLink(direction) {
            const range = this.shiftedRange(direction)
            if (!range) return null
            const query = new URLSearchParams(window.location.search)
            query.delete('interval')
            query.delete('start')
            query.set('from', range.from)
            query.set('to', range.to)
            return `summary?${query.toString()}`
        },
        onSubmit() {
            const query = new URLSearchParams(window.location.search)
            query.delete('interval')
            query.delete('start')
            query.set('from', this.fromDate)
            query.set('to', this.toDate)
            this.state.showDropdownTimepicker = false
            window.location.search = query.toString()
        },
        mounted() {
            window.addEventListener('click', (e) => {
                const skip = findParentAttribute(e.target, 'data-trigger-for')?.value
                Object.keys(this.state).filter(k => k !== skip).forEach(k => this.state[k] = false)
            })

            const query = new URLSearchParams(window.location.search)
            if (query.has('interval')) {
                const refEl = document.getElementById(`time-option-${query.get('interval')}`)
                this.timeSelection = refEl ? refEl.childNodes[0].textContent : 'Unknown'
            }

            // Time picker keyboard shortcuts
            document.addEventListener('keyup', e => {
                if (e.target !== document.body) return
                const optionEl = document.querySelector(`a[data-hotkey="${e.key}"]`)
                if (optionEl) optionEl.click()
            })
        }
    }
}
