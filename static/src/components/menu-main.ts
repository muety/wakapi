import {createApp} from 'petite-vue';
import {findParentAttribute} from '../base';

function MainMenu() {
	return {
		$delimiters: ['${', '}'],
		state: {
			showDropdownResources: false,
			showDropdownUser: false,
			showApiKey: false,
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
		},
	};
}

createApp(MainMenu).mount('#main-menu');
