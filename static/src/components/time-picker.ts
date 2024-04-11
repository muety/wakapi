// eslint-disable-next-line import/no-unassigned-import
import 'typed-query-selector';
import {findParentAttribute} from '../base';

function TimePicker({fromDate, toDate, timeSelection}: {
	fromDate: string;
	toDate: string;
	timeSelection: string;
}) {
	return {
		$template: '#time-picker-template',
		$delimiters: ['${', '}'],
		state: {
			showDropdownTimepicker: false,
		},
		fromDate,
		toDate,
		timeSelection,
		intervalLink(interval: string) {
			const queryParameters = new URLSearchParams(window.location.search);
			queryParameters.set('interval', interval);
			return `summary?${queryParameters.toString()}`;
		},
		onDateUpdated() {
			document.querySelector('form#time-picker-form')?.submit();
		},
		mounted() {
			window.addEventListener('click', event => {
				const skip = findParentAttribute(event.target as Element, 'data-trigger-for');
				for (const _k in this.state) {
					if (!Object.hasOwn(this.state, _k)) {
						continue;
					}

					const k = _k as keyof typeof this.state;
					if (k !== skip) {
						continue;
					}

					this.state[k] = false;
				}
			});

			const query = new URLSearchParams(window.location.search);
			if (query.has('interval')) {
				const referenceElement = document.querySelector(`#time-option-${query.get('interval')}`);
				this.timeSelection = referenceElement?.textContent ?? 'Unknown';
			}
		},
	};
}

Object.assign(window, {
	TimePicker,
});
