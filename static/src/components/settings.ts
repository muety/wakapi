import {createApp} from 'petite-vue';
// eslint-disable-next-line import/no-unassigned-import
import 'typed-query-selector';
import type Timezones from '../timezones';

type Data = {value: string; text: string};

declare const defaultTab: string;
declare const userTimeZone: typeof Timezones[number];
declare const defaultTzOption: Data;
declare const tzs: typeof Timezones;

function Settings() {
	return {
		// $delimiters: ['${', '}'],  // https://github.com/vuejs/petite-vue/pull/100
		activeTab: defaultTab,
		selectedTimezone: userTimeZone,
		vibrantColorsEnabled: JSON.parse(
			localStorage.getItem('wakapi_vibrant_colors') ?? 'false',
		) as boolean,
		// eslint-disable-next-line @typescript-eslint/consistent-type-assertions
		labels: {} as Record<string, boolean>,
		get tzOptions(): Data[] {
			return [
				defaultTzOption,
				...tzs.toSorted().map(tz => ({value: tz, text: tz})),
			];
		},
		updateTab() {
			this.activeTab = window.location.hash.slice(1) || defaultTab;
		},
		isActive(tab: string) {
			return this.activeTab === tab;
		},
		/* eslint-disable no-alert */
		confirmRegenerate() {
			if (confirm('Are you sure?')) {
				document.querySelector('form#form-regenerate-summaries')!.submit();
			}
		},
		confirmWakatimeImport() {
			if (confirm('Are you sure? The import can not be undone.')) {
				// Weird hack to sync the "legacy importer" form field from the wakatime connection form to the (invisible) import form
				document.querySelector('input#use_legacy_importer')!.value = document.querySelector('input#use_legacy_importer_tmp')!.checked.toString();
				document.querySelector('form#form-import-wakatime')!.submit();
			}
		},
		confirmClearData() {
			if (confirm('Are you sure? This can not be undone!')) {
				document.querySelector('form#form-clear-data')!.submit();
			}
		},
		confirmDeleteAccount() {
			if (confirm('Are you sure? This can not be undone!')) {
				document.querySelector('form#form-delete-user')!.submit();
			}
		},
		/* eslint-enable no-alert */
		onToggleVibrantColors() {
			localStorage.setItem('wakapi_vibrant_colors', JSON.stringify(this.vibrantColorsEnabled));
		},
		showProjectAddButton(index: string) {
			this.labels[index] = true;
		},
		mounted() {
			this.updateTab();
			window.addEventListener('hashchange', () => {
				this.updateTab();
			});
		},
	};
}

createApp().mount('#settings-page');
