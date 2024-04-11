import {createApp} from 'petite-vue';

function Summary() {
	return {
		$delimiters: ['${', '}'],
		activityChartSvg: '',
		get currentInterval() {
			const urlParameters = new URLSearchParams(window.location.search);
			if (urlParameters.has('interval')) {
				return urlParameters.get('interval');
			}

			if (!urlParameters.has('from') && !urlParameters.has('to')) {
				return 'today';
			}

			return null;
		},
		mounted({userId}: {userId: `${number}`}) {
			void fetch(`api/activity/chart/${userId}.svg?dark&noattr`)
				.then(async response => response.text())
				.then(data => {
					this.activityChartSvg = data;
				});
		},
	};
}

createApp(Summary).mount('#summary-page');
