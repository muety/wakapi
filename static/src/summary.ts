import Chart, {type LegendItem, type TooltipOptions} from 'chart.js/auto';
// eslint-disable-next-line import/no-unassigned-import
import 'typed-query-selector';
import seedrandom from 'seedrandom';
import type {DeepPartial} from 'node_modules/chart.js/dist/types/utils';

type Data = Array<{key: string; total: number}>;
type Colors = Record<string, string>;

declare const wakapiData: {
	projects: Data;
	operatingSystems: Data;
	editors: Data;
	languages: Data;
	machines: Data;
	labels: Data;
	branches: Data;
	entities: Data;
};

declare const languageColors: Colors;
declare const editorColors: Colors;
declare const osColors: Colors;

// Dirty hack to vertically align legends across multiple charts
// however, without monospace font, it's still not perfectly aligned
// waiting for https://github.com/chartjs/Chart.js/discussions/9890
const LEGEND_CHARACTERS = 20;

// https://hihayk.github.io/scale/#4/6/50/80/-51/67/20/14/276749/39/103/73/white
const baseColors = ['#112836', '#163B43', '#1C4F4D', '#215B4C', '#276749', '#437C57', '#5F9167', '#7DA67C', '#9FBA98', '#BFCEB5', '#DCE2D3'];

const projectsCanvas = document.querySelector('canvas#chart-projects')!;
const osCanvas = document.querySelector('canvas#chart-os')!;
const editorsCanvas = document.querySelector('canvas#chart-editor')!;
const languagesCanvas = document.querySelector('canvas#chart-language')!;
const machinesCanvas = document.querySelector('canvas#chart-machine')!;
const labelsCanvas = document.querySelector('canvas#chart-label')!;
const branchesCanvas = document.querySelector('canvas#chart-branches')!;
const entitiesCanvas = document.querySelector('canvas#chart-entities')!;

const projectContainer = document.querySelector('div#project-container')!;
const osContainer = document.querySelector('div#os-container')!;
const editorContainer = document.querySelector('div#editor-container')!;
const languageContainer = document.querySelector('div#language-container')!;
const machineContainer = document.querySelector('div#machine-container')!;
const labelContainer = document.querySelector('div#label-container')!;
const branchContainer = document.querySelector('div#branch-container')!;
const entityContainer = document.querySelector('div#entity-container')!;

const containers = [projectContainer, osContainer, editorContainer, languageContainer, machineContainer, labelContainer, branchContainer, entityContainer];
const canvases = [projectsCanvas, osCanvas, editorsCanvas, languagesCanvas, machinesCanvas, labelsCanvas, branchesCanvas, entitiesCanvas];
const data = [wakapiData.projects, wakapiData.operatingSystems, wakapiData.editors, wakapiData.languages, wakapiData.machines, wakapiData.labels, wakapiData.branches, wakapiData.entities];

const topNPickers = [...document.querySelectorAll('input.top-picker')];
topNPickers.sort(
	(a, b) =>
		Number.parseInt(a.dataset.entity!, 10)
		- Number.parseInt(b.dataset.entity!, 10),
);

for (const element of topNPickers) {
	const index = Number.parseInt(element.dataset.entity!, 10);
	element.max = String(data[index].length);
	element.value = String(Math.min(Number.parseInt(element.max, 10), 9));
}

const charts: Chart[] = [];
let showTopN: number[] = [];

Chart.defaults.color = '#E2E8F0';
Chart.defaults.borderColor = '#242b3a';
Chart.defaults.font.family = 'Source Sans 3, Roboto, Helvetica Neue, Arial, sens-serif';

function toHHMMSS(number: string) {
	const secondNumber = Number.parseInt(number, 10);
	let hours: number | string = Math.floor(secondNumber / 3600);
	let minutes: number | string = Math.floor((secondNumber - (hours * 3600)) / 60);
	let seconds: number | string = secondNumber - (hours * 3600) - (minutes * 60);

	if (hours < 10) {
		hours = '0' + hours;
	}

	if (minutes < 10) {
		minutes = '0' + minutes;
	}

	if (seconds < 10) {
		seconds = '0' + seconds;
	}

	return `${hours}:${minutes}:${seconds}`;
}

// eslint-disable-next-line complexity
function draw(subselection?: number[]) {
	function getTooltipOptions(key: keyof typeof wakapiData): DeepPartial<TooltipOptions> {
		return {
			callbacks: {
				label(item) {
					const d = wakapiData[key][item.dataIndex];
					return ` ${d.key}: ${toHHMMSS(d.total.toString())}`;
				},
				title: () => 'Total Time',
				footer: () => key === 'projects' ? 'Click for details' : undefined,
			},
		};
	}

	function filterLegendItem(item: LegendItem) {
		item.text = item.text.length > LEGEND_CHARACTERS ? item.text.slice(0, LEGEND_CHARACTERS - 3).padEnd(LEGEND_CHARACTERS, '.') : item.text;
		item.text = item.text.padEnd(LEGEND_CHARACTERS + 3);
		return true;
	}

	function shouldUpdate(index: number) {
		return !subselection || (subselection.includes(index) && data[index].length >= showTopN[index]);
	}

	for (const c of charts
		.filter((c, i) => shouldUpdate(i))) {
		c.destroy();
	}

	const vibrantColors = Boolean(JSON.parse(window.localStorage.getItem('wakapi_vibrant_colors') ?? 'false'));

	const projectChart = projectsCanvas && !projectsCanvas.classList.contains('hidden') && shouldUpdate(0)
		? new Chart(projectsCanvas.getContext('2d')!, {
			// Type: 'horizontalBar',
			type: 'bar',
			data: {
				datasets: [{
					data: wakapiData.projects
						.slice(0, Math.min(showTopN[0], wakapiData.projects.length))
						.map(p => p.total),
					backgroundColor: wakapiData.projects.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.projects.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
				}],
				labels: wakapiData.projects
					.slice(0, Math.min(showTopN[0], wakapiData.projects.length))
					.map(p => p.key),
			},
			options: {
				indexAxis: 'y',
				scales: {
					xAxes: {
						title: {
							display: true,
							text: 'Duration (hh:mm:ss)',
						},
						ticks: {
							callback: label => toHHMMSS(label.toString()),
						},
					},
				},
				plugins: {
					legend: {
						display: false,
					},
					tooltip: getTooltipOptions('projects'),
				},
				maintainAspectRatio: false,
				onClick(event, data) {
					const url = new URL(window.location.href);
					const name = wakapiData.projects[data[0].index].key;
					url.searchParams.set('project', name === 'unknown' ? '-' : name); // Will break if the project is actually named "unknown"
					window.location.href = url.href;
				},
				onHover(event, element) {
					const target = event.native!.target as HTMLElement;
					target.style.cursor = element[0] ? 'pointer' : 'default';
				},
			},
		})
		: null;

	const osChart = osCanvas && !osCanvas.classList.contains('hidden') && shouldUpdate(1)
		? new Chart(osCanvas.getContext('2d')!, {
			type: 'pie',
			data: {
				datasets: [{
					data: wakapiData.operatingSystems
						.slice(0, Math.min(showTopN[1], wakapiData.operatingSystems.length))
						.map(p => p.total),
					backgroundColor: wakapiData.operatingSystems.map((p: Data[number], i: number) => {
						const c = hexToRgb(vibrantColors ? (osColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.operatingSystems.map((p, i) => {
						const c = hexToRgb(vibrantColors ? (osColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
					borderWidth: 0,
				}],
				labels: wakapiData.operatingSystems
					.slice(0, Math.min(showTopN[1], wakapiData.operatingSystems.length))
					.map(p => p.key),
			},
			options: {
				plugins: {
					tooltip: getTooltipOptions('operatingSystems'),
					legend: {
						position: 'right',
						labels: {
							filter: filterLegendItem,
						},
					},
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	const editorChart = editorsCanvas && !editorsCanvas.classList.contains('hidden') && shouldUpdate(2)
		? new Chart(editorsCanvas.getContext('2d')!, {
			type: 'pie',
			data: {
				datasets: [{
					data: wakapiData.editors
						.slice(0, Math.min(showTopN[2], wakapiData.editors.length))
						.map(p => p.total),
					backgroundColor: wakapiData.editors.map((p, i) => {
						const c = hexToRgb(vibrantColors ? (editorColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.editors.map((p, i) => {
						const c = hexToRgb(vibrantColors ? (editorColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
					borderWidth: 0,
				}],
				labels: wakapiData.editors
					.slice(0, Math.min(showTopN[2], wakapiData.editors.length))
					.map(p => p.key),
			},
			options: {
				plugins: {
					tooltip: getTooltipOptions('editors'),
					legend: {
						position: 'right',
						labels: {
							filter: filterLegendItem,
						},
					},
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	const languageChart = languagesCanvas && !languagesCanvas.classList.contains('hidden') && shouldUpdate(3)
		? new Chart(languagesCanvas.getContext('2d')!, {
			type: 'pie',
			data: {
				datasets: [{
					data: wakapiData.languages
						.slice(0, Math.min(showTopN[3], wakapiData.languages.length))
						.map(p => p.total),
					backgroundColor: wakapiData.languages.map(p => {
						const c = hexToRgb(languageColors[p.key.toLowerCase()] || getRandomColor(p.key))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.languages.map(p => {
						const c = hexToRgb(languageColors[p.key.toLowerCase()] || getRandomColor(p.key))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
					borderWidth: 0,
				}],
				labels: wakapiData.languages
					.slice(0, Math.min(showTopN[3], wakapiData.languages.length))
					.map(p => p.key),
			},
			options: {
				plugins: {
					tooltip: getTooltipOptions('languages'),
					legend: {
						position: 'right',
						labels: {
							filter: filterLegendItem,
						},
						title: {
							display: true,
						},
					},
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	const machineChart = machinesCanvas && !machinesCanvas.classList.contains('hidden') && shouldUpdate(4)
		? new Chart(machinesCanvas.getContext('2d')!, {
			type: 'pie',
			data: {
				datasets: [{
					data: wakapiData.machines
						.slice(0, Math.min(showTopN[4], wakapiData.machines.length))
						.map(p => p.total),
					backgroundColor: wakapiData.machines.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.machines.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
					borderWidth: 0,
				}],
				labels: wakapiData.machines
					.slice(0, Math.min(showTopN[4], wakapiData.machines.length))
					.map(p => p.key),
			},
			options: {
				plugins: {
					tooltip: getTooltipOptions('machines'),
					legend: {
						position: 'right',
						labels: {
							filter: filterLegendItem,
						},
					},
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	const labelChart = labelsCanvas && !labelsCanvas.classList.contains('hidden') && shouldUpdate(5)
		? new Chart(labelsCanvas.getContext('2d')!, {
			type: 'pie',
			data: {
				datasets: [{
					data: wakapiData.labels
						.slice(0, Math.min(showTopN[5], wakapiData.labels.length))
						.map(p => p.total),
					backgroundColor: wakapiData.labels.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.labels.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
					borderWidth: 0,
				}],
				labels: wakapiData.labels
					.slice(0, Math.min(showTopN[5], wakapiData.labels.length))
					.map(p => p.key),
			},
			options: {
				plugins: {
					tooltip: getTooltipOptions('labels'),
					legend: {
						position: 'right',
						labels: {
							filter: filterLegendItem,
						},
					},
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	const branchChart = branchesCanvas && !branchesCanvas.classList.contains('hidden') && shouldUpdate(6)
		? new Chart(branchesCanvas.getContext('2d')!, {
			type: 'bar',
			data: {
				datasets: [{
					data: wakapiData.branches
						.slice(0, Math.min(showTopN[6], wakapiData.branches.length))
						.map(p => p.total),
					backgroundColor: wakapiData.branches.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.branches.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
				}],
				labels: wakapiData.branches
					.slice(0, Math.min(showTopN[6], wakapiData.branches.length))
					.map(p => p.key),
			},
			options: {
				indexAxis: 'y',
				scales: {
					xAxes: {
						title: {
							display: true,
							text: 'Duration (hh:mm:ss)',
						},
						ticks: {
							callback: label => toHHMMSS(label.toString()),
						},
					},
				},
				plugins: {
					legend: {
						display: false,
					},
					tooltip: getTooltipOptions('branches'),
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	const entityChart = entitiesCanvas && !entitiesCanvas.classList.contains('hidden') && shouldUpdate(7)
		? new Chart(entitiesCanvas.getContext('2d')!, {
			// Type: 'horizontalBar',
			type: 'bar',
			data: {
				datasets: [{
					data: wakapiData.entities
						.slice(0, Math.min(showTopN[7], wakapiData.entities.length))
						.map(p => p.total),
					backgroundColor: wakapiData.entities.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`;
					}),
					hoverBackgroundColor: wakapiData.entities.map((p, i) => {
						const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))!;
						return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`;
					}),
				}],
				labels: wakapiData.entities
					.slice(0, Math.min(showTopN[7], wakapiData.entities.length))
					.map(p => extractFile(p.key)),
			},
			options: {
				indexAxis: 'y',
				scales: {
					xAxes: {
						title: {
							display: true,
							text: 'Duration (hh:mm:ss)',
						},
						ticks: {
							callback: label => toHHMMSS(label.toString()),
						},
					},
				},
				plugins: {
					legend: {
						display: false,
					},
					tooltip: getTooltipOptions('entities'),
				},
				maintainAspectRatio: false,
			},
		})
		: null;

	charts[0] = projectChart ?? charts[0];
	charts[1] = osChart ?? charts[1];
	charts[2] = editorChart ?? charts[2];
	charts[3] = languageChart ?? charts[3];
	charts[4] = machineChart ?? charts[4];
	charts[5] = labelChart ?? charts[5];
	charts[6] = branchChart ?? charts[6];
	charts[7] = entityChart ?? charts[7];
}

function parseTopN() {
	showTopN = topNPickers.map(element => Number.parseInt(element.value, 10));
}

function togglePlaceholders(mask: boolean[]) {
	const placeholderElements = containers.map(c => c ? c.querySelector('.placeholder-container') : null);

	for (const [i, element] of mask.entries()) {
		const placeholder = placeholderElements[i];
		if (!placeholder) {
			continue;
		}

		if (element) {
			canvases[i].classList.remove('hidden');
			placeholder.classList.add('hidden');
		} else {
			canvases[i].classList.add('hidden');
			placeholder.classList.remove('hidden');
		}
	}
}

function getPresentDataMask() {
	return data.map(
		list => (list ? list.reduce(
			(
				accumulator,
				item,
			) => accumulator + item.total,
			0,
		) : 0) > 0);
}

function getColor(seed: string, index: number) {
	if (index < baseColors.length) {
		return baseColors[(index + 5) % baseColors.length];
	}

	return getRandomColor(seed);
}

function getRandomColor(seed: string) {
	seed ??= '1234567';
	const rng = seedrandom(seed);
	const letters = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'] as const;
	let color = '#';
	for (let i = 0; i < 6; i++) {
		color += letters[Math.floor(rng() * 16)];
	}

	return color;
}

// https://stackoverflow.com/a/5624139/3112139
function hexToRgb(hex: string) {
	const shorthandRegex = /^#?([a-f\d])([a-f\d])([a-f\d])$/i;
	hex = hex.replace(shorthandRegex, (m, r: string, g: string, b: string) => r + r + g + g + b + b);
	const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
	return result ? {
		r: Number.parseInt(result[1], 16),
		g: Number.parseInt(result[2], 16),
		b: Number.parseInt(result[3], 16),
	} : null;
}

function swapCharts(showEntity: string, hideEntity: string) {
	document.querySelector(`#${showEntity}-container`)!.parentElement!.classList.remove('hidden');
	document.querySelector(`#${hideEntity}-container`)!.parentElement!.classList.add('hidden');
}

function extractFile(filePath: string) {
	const delimiter = filePath.includes('\\') ? '\\' : '/'; // Windows style path?
	return filePath.split(delimiter).at(-1);
}

function updateNumberTotal() {
	for (const [i, datum] of data.entries()) {
		document.querySelector(`span[data-entity='${i}']`)!.textContent = datum.length.toString();
	}
}

window.addEventListener('load', () => {
	for (const element of topNPickers) {
		element.addEventListener('change', () => {
			parseTopN();
			draw([Number.parseInt(element.dataset.entity!, 10)]);
		});
	}

	parseTopN();
	togglePlaceholders(getPresentDataMask());
	draw();
	updateNumberTotal();
});
