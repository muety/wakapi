function EntityFilter({type, options, selection}) {
    return {
        $template: '#entity-filter-template',
        $delimiters: ['${', '}'],
        type: type,
        options: options,
        selection: selection,
        display() {
            return this.type.capitalize()
        },
        onSelectionUpdated(e) {
            this.selection = e.target.value == 'null' ? null : e.target.value
            this.$nextTick(() => {
                const query = new URLSearchParams(window.location.search)
                const val = this.selection === 'unknown' ? '-' : this.selection  // will break if the project is actually named "unknown"
                if (this.selection) query.set(type, val)
                else query.delete(type)
                window.location.search = query.toString()
            })
        },
        mounted() {
            const query = new URLSearchParams(window.location.search)
            if (query.has(type)) {
                const val = query.get(type) === '-' ? 'unknown' : query.get(type)
                if (!this.options.includes(val)) {
                    this.options = [val, ...this.options]
                }
                this.selection = val
            }
        }
    }
}
