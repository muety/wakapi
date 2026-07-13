function ComboboxFilter({type, options, selection, remote, minChars, debounceMs, project}) {
    return {
        $template: '#combobox-filter-template',
        $delimiters: ['${', '}'],
        type: type,
        options: options || [],
        selection: selection || null,
        remote: !!remote,
        minChars: minChars || 1,
        debounceMs: debounceMs || 300,
        project: project || '',
        query: '',
        open: false,
        loading: false,
        hint: '',
        visibleOptions: [],
        _debounceTimer: null,
        _boundOutside: null,
        onInput() {
            this.open = true
            if (this.remote) {
                this.onRemoteInput()
            } else {
                this.filterLocal()
            }
        },
        onFocus() {
            this.open = true
            if (!this.remote && !this.visibleOptions.length) this.filterLocal()
        },
        filterLocal() {
            const q = this.query.trim().toLowerCase()
            this.visibleOptions = q
                ? this.options.filter(o => String(o).toLowerCase().includes(q))
                : this.options.slice()
            this.hint = ''
        },
        onRemoteInput() {
            if (this._debounceTimer) {
                clearTimeout(this._debounceTimer)
                this._debounceTimer = null
            }
            const q = this.query.trim()
            if (q.length < this.minChars) {
                this.visibleOptions = []
                this.loading = false
                this.hint = `Type at least ${this.minChars} characters …`
                return
            }
            this.hint = ''
            this.loading = true
            this._debounceTimer = setTimeout(() => this.fetchRemote(q), this.debounceMs)
        },
        fetchRemote(q) {
            const params = new URLSearchParams()
            if (this.project) params.set('project', this.project)
            params.set('q', q)
            fetch(`api/branches?${params.toString()}`, {headers: {'Accept': 'application/json'}})
                .then(res => {
                    if (!res.ok) throw new Error(`request failed with status ${res.status}`)
                    return res.json()
                })
                .then(data => {
                    this.visibleOptions = Array.isArray(data) ? data : []
                    this.loading = false
                    this.hint = this.visibleOptions.length ? '' : 'No matches'
                })
                .catch(err => {
                    console.error('branch search failed', err)
                    this.visibleOptions = []
                    this.loading = false
                    this.hint = 'Search failed'
                })
        },
        select(option) {
            this.selection = option
            this.query = option
            this.open = false
            this.$nextTick(() => {
                const query = new URLSearchParams(window.location.search)
                const val = this.selection === 'unknown' ? '-' : this.selection  // will break if the value is actually named "unknown"
                if (this.selection) query.set(this.type, val)
                else query.delete(this.type)
                window.location.search = query.toString()
            })
        },
        onKeydown(e) {
            if (e.key === 'Escape') {
                this.open = false
            }
        },
        onClickOutside(e) {
            if (this.$el && !this.$el.contains(e.target)) this.open = false
        },
        mounted() {
            const query = new URLSearchParams(window.location.search)
            if (query.has(this.type)) {
                const val = query.get(this.type) === '-' ? 'unknown' : query.get(this.type)
                this.selection = val
                this.query = val
            }
            if (!this.remote) this.filterLocal()
            this._boundOutside = e => this.onClickOutside(e)
            document.addEventListener('click', this._boundOutside)
        },
    }
}
