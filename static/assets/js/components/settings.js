PetiteVue.createApp({
	//$delimiters: ['${', '}'],  // https://github.com/vuejs/petite-vue/pull/100
	activeTab: defaultTab,
	selectedTimezone: userTimeZone,
	vibrantColorsEnabled: JSON.parse(
		localStorage.getItem("wakapi_vibrant_colors"),
	),
	labels: {},
	get tzOptions() {
		return [
			defaultTzOption,
			...tzs.sort().map((tz) => ({value: tz, text: tz})),
		];
	},
	updateTab() {
		this.activeTab = window.location.hash.slice(1) || defaultTab;
	},
	isActive(tab) {
		return this.activeTab === tab;
	},
	confirmRegenerate() {
		if (confirm("Are you sure?")) {
			document.querySelector("#form-regenerate-summaries").submit();
		}
	},
	confirmWakatimeImport() {
		if (confirm("Are you sure? The import can not be undone.")) {
			// weird hack to sync the "legacy importer" form field from the wakatime connection form to the (invisible) import form
			document.getElementById('use_legacy_importer').value = document.getElementById('use_legacy_importer_tmp').checked.toString()
			document.querySelector("#form-import-wakatime").submit();
		}
	},
	confirmClearData() {
		if (confirm("Are you sure? This can not be undone!")) {
			document.querySelector("#form-clear-data").submit();
		}
	},
	confirmDeleteAccount() {
		if (confirm("Are you sure? This can not be undone!")) {
			document.querySelector("#form-delete-user").submit();
		}
	},
	onToggleVibrantColors() {
		localStorage.setItem("wakapi_vibrant_colors", this.vibrantColorsEnabled);
	},
	showProjectAddButton(index) {
		this.labels[index] = true;
	},
	mounted() {
		this.updateTab();
		window.addEventListener("hashchange", () => this.updateTab());
	},
}).mount("#settings-page");
