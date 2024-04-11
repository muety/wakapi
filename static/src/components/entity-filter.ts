import type {nextTick} from 'petite-vue/scheduler';

function EntityFilter({type, options, selection}: {
	type: string;
	options: string[];
	// eslint-disable-next-line @typescript-eslint/ban-types
	selection: string | null;
}) {
	return {
		$template: '#entity-filter-template',
		$delimiters: ['${', '}'],
		type,
		options,
		selection,
		display() {
			return this.type.toUpperCase();
		},
		onSelectionUpdated(event: Event) {
			this.selection = (event.target as HTMLSelectElement).value;
			const $nt = (this as unknown as {$nextTick: typeof nextTick}).$nextTick;
			void $nt(() => {
				const query = new URLSearchParams(window.location.search);
				const value = this.selection === 'unknown' ? '-' : this.selection; // Will break if the project is actually named "unknown"
				if (value && this.selection) {
					query.set(type, value);
				} else {
					query.delete(type);
				}

				window.location.search = query.toString();
			});
		},
		mounted() {
			const query = new URLSearchParams(window.location.search);
			if (query.has(type)) {
				const value = query.get(type) === '-' ? 'unknown' : query.get(type);
				if (value && !this.options.includes(value)) {
					this.options = [value, ...this.options];
				}

				this.selection = value;
			}
		},
	};
}

Object.assign(window, {
	EntityFilter,
});
