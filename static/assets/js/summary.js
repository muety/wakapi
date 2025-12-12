// dirty hack to vertically align legends across multiple charts
// however, without monospace font, it's still not perfectly aligned
// waiting for https://github.com/chartjs/Chart.js/discussions/9890
const LEGEND_CHARACTERS = 20

const projectsCanvas = document.getElementById('chart-projects')
const osCanvas = document.getElementById('chart-os')
const editorsCanvas = document.getElementById('chart-editor')
const languagesCanvas = document.getElementById('chart-language')
const machinesCanvas = document.getElementById('chart-machine')
const labelsCanvas = document.getElementById('chart-label')
const branchesCanvas = document.getElementById('chart-branches')
const entitiesCanvas = document.getElementById('chart-entities')
const categoriesCanvas = document.getElementById('chart-categories')
const timelineCanvas = document.getElementById('chart-timeline')
const hourlyCanvas = document.getElementById('chart-hourly')

const projectContainer = document.getElementById('project-container')
const osContainer = document.getElementById('os-container')
const editorContainer = document.getElementById('editor-container')
const languageContainer = document.getElementById('language-container')
const machineContainer = document.getElementById('machine-container')
const labelContainer = document.getElementById('label-container')
const branchContainer = document.getElementById('branch-container')
const entityContainer = document.getElementById('entity-container')
const categoryContainer = document.getElementById('category-container')
const timelineContainer = document.getElementById('timeline-container')
const hourlyContainer = document.getElementById('hourly-container')

const containers = [projectContainer, osContainer, editorContainer, languageContainer, machineContainer, labelContainer, branchContainer, entityContainer, categoryContainer, timelineContainer, hourlyContainer]
const canvases = [projectsCanvas, osCanvas, editorsCanvas, languagesCanvas, machinesCanvas, labelsCanvas, branchesCanvas, entitiesCanvas, categoriesCanvas, timelineCanvas, hourlyCanvas]
const data = [wakapiData.projects, wakapiData.operatingSystems, wakapiData.editors, wakapiData.languages, wakapiData.machines, wakapiData.labels, wakapiData.branches, wakapiData.entities, wakapiData.categories, wakapiData.timelineStats, wakapiData.hourlyBreakdown]

let topNPickers = [...document.getElementsByClassName('top-picker')]
topNPickers.sort(((a, b) => parseInt(a.attributes['data-entity'].value) - parseInt(b.attributes['data-entity'].value)))
topNPickers.forEach(e => {
    const idx = parseInt(e.attributes['data-entity'].value)
    e.max = data[idx].length
    e.value = Math.min(e.max, 9)
})

let charts = []
let showTopN = []

Chart.defaults.font.family = 'Source Sans 3, Roboto, Helvetica Neue, Arial, sens-serif'

String.prototype.toHHMMSS = function () {
    const sec_num = parseInt(this, 10)
    let hours = Math.floor(sec_num / 3600)
    let minutes = Math.floor((sec_num - (hours * 3600)) / 60)
    let seconds = sec_num - (hours * 3600) - (minutes * 60)

    if (hours < 10) {
        hours = '0' + hours
    }
    if (minutes < 10) {
        minutes = '0' + minutes
    }
    if (seconds < 10) {
        seconds = '0' + seconds
    }
    return `${hours}:${minutes}:${seconds}`
}

function filterLegendItem(item) {
    if (!item || !item.text) return false;
    item.text = item.text.length > LEGEND_CHARACTERS ? item.text.slice(0, LEGEND_CHARACTERS - 3).padEnd(LEGEND_CHARACTERS, '.') : item.text
    item.text = item.text.padEnd(LEGEND_CHARACTERS + 3)
    return true
}

function draw(subselection) {
    function getTooltipOptions(key, stacked) {
        return {
            callbacks: {
                label: (item) => {
                    const raw = item.chart.data.datasets[item.datasetIndex].data[item.dataIndex]
                    const val = (typeof raw === 'object' && raw !== null) ? raw.x : raw
                    const lbl = (typeof raw === 'object' && raw !== null && raw.details) ? raw.details : item.chart.data.labels[item.dataIndex]
                    const d = stacked
                        ? [val, item.chart.data.datasets[item.datasetIndex].label]
                        : [val, lbl]
                    return ` ${d[1]}: ${d[0].toString().toHHMMSS()}`
                },
                title: () => 'Total Time',
                footer: () => key === 'projects' ? 'Click for details' : null
            }
        }
    }

    function shouldUpdate(index) {
        return !subselection || (subselection.includes(index) && data[index].length >= showTopN[index])
    }

    charts
        .filter((c, i) => shouldUpdate(i))
        .forEach(c => c.destroy())

    const vibrantColors = JSON.parse(window.localStorage.getItem('wakapi_vibrant_colors') || false);

    let projectChart = projectsCanvas && !projectsCanvas.classList.contains('hidden') && shouldUpdate(0)
        ? new Chart(projectsCanvas.getContext('2d'), {
            //type: 'horizontalBar',
            type: "bar",
            data: {
                datasets: [{
                    data: wakapiData.projects
                        .slice(0, Math.min(showTopN[0], wakapiData.projects.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.projects.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.projects.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                }],
                labels: wakapiData.projects
                    .slice(0, Math.min(showTopN[0], wakapiData.projects.length))
                    .map(p => p.key)
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
                            callback: (label) => label.toString().toHHMMSS(),
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false,
                    },
                    tooltip: getTooltipOptions('projects'),
                },
                maintainAspectRatio: false,
                onClick: (event, data) => {
                    const url = new URL(window.location.href)
                    const name = wakapiData.projects[data[0].index].key
                    url.searchParams.set('project', name === 'unknown' ? '-' : name)  // will break if the project is actually named "unknown"
                    window.location.href = url.href
                },
                onHover: (event, elem) => {
                    event.native.target.style.cursor = elem[0] ? 'pointer' : 'default'
                }
            }
        })
        : null

    let osChart = osCanvas && !osCanvas.classList.contains('hidden') && shouldUpdate(1)
        ? new Chart(osCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.operatingSystems
                        .slice(0, Math.min(showTopN[1], wakapiData.operatingSystems.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.operatingSystems.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? (osColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.operatingSystems.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? (osColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderWidth: 0
                }],
                labels: wakapiData.operatingSystems
                    .slice(0, Math.min(showTopN[1], wakapiData.operatingSystems.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('operatingSystems'),
                    legend: {
                        position: 'right',
                        labels: {
                            filter: filterLegendItem
                        },
                    },
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let editorChart = editorsCanvas && !editorsCanvas.classList.contains('hidden') && shouldUpdate(2)
        ? new Chart(editorsCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.editors
                        .slice(0, Math.min(showTopN[2], wakapiData.editors.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.editors.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? (editorColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.editors.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? (editorColors[p.key.toLowerCase()] || getRandomColor(p.key)) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderWidth: 0
                }],
                labels: wakapiData.editors
                    .slice(0, Math.min(showTopN[2], wakapiData.editors.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('editors'),
                    legend: {
                        position: 'right',
                        labels: {
                            filter: filterLegendItem
                        },
                    },
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let languageChart = languagesCanvas && !languagesCanvas.classList.contains('hidden') && shouldUpdate(3)
        ? new Chart(languagesCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.languages
                        .slice(0, Math.min(showTopN[3], wakapiData.languages.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.languages.map(p => {
                        const c = hexToRgb(languageColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.languages.map(p => {
                        const c = hexToRgb(languageColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderWidth: 0
                }],
                labels: wakapiData.languages
                    .slice(0, Math.min(showTopN[3], wakapiData.languages.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('languages'),
                    legend: {
                        position: 'right',
                        labels: {
                            filter: filterLegendItem
                        },
                        title: {
                            display: true,
                        }
                    },
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let machineChart = machinesCanvas && !machinesCanvas.classList.contains('hidden') && shouldUpdate(4)
        ? new Chart(machinesCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.machines
                        .slice(0, Math.min(showTopN[4], wakapiData.machines.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.machines.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.machines.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderWidth: 0
                }],
                labels: wakapiData.machines
                    .slice(0, Math.min(showTopN[4], wakapiData.machines.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('machines'),
                    legend: {
                        position: 'right',
                        labels: {
                            filter: filterLegendItem
                        },
                    },
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let labelChart = labelsCanvas && !labelsCanvas.classList.contains('hidden') && shouldUpdate(5)
        ? new Chart(labelsCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.labels
                        .slice(0, Math.min(showTopN[5], wakapiData.labels.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.labels.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.labels.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderWidth: 0
                }],
                labels: wakapiData.labels
                    .slice(0, Math.min(showTopN[5], wakapiData.labels.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('labels'),
                    legend: {
                        position: 'right',
                        labels: {
                            filter: filterLegendItem
                        },
                    },
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let branchChart = branchesCanvas && !branchesCanvas.classList.contains('hidden') && shouldUpdate(6)
        ? new Chart(branchesCanvas.getContext('2d'), {
            type: "bar",
            data: {
                datasets: [{
                    data: wakapiData.branches
                        .slice(0, Math.min(showTopN[6], wakapiData.branches.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.branches.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.branches.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                }],
                labels: wakapiData.branches
                    .slice(0, Math.min(showTopN[6], wakapiData.branches.length))
                    .map(p => p.key)
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
                            callback: (label) => label.toString().toHHMMSS(),
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false,
                    },
                    tooltip: getTooltipOptions('branches'),
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let entityChart = entitiesCanvas && !entitiesCanvas.classList.contains('hidden') && shouldUpdate(7)
        ? new Chart(entitiesCanvas.getContext('2d'), {
            //type: 'horizontalBar',
            type: "bar",
            data: {
                datasets: [{
                    data: wakapiData.entities
                        .slice(0, Math.min(showTopN[7], wakapiData.entities.length))
                        .map(p => ({
                            x: parseInt(p.total),
                            y: extractFile(p.key),
                            details: p.key
                        })),
                    backgroundColor: wakapiData.entities.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.entities.map((p, i) => {
                        const c = hexToRgb(vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                }],
                labels: wakapiData.entities
                    .slice(0, Math.min(showTopN[7], wakapiData.entities.length))
                    .map(p => extractFile(p.key))
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
                            callback: (label) => label.toString().toHHMMSS(),
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false,
                    },
                    tooltip: getTooltipOptions('entities'),
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let categoryChart = categoriesCanvas && !categoriesCanvas.classList.contains('hidden') && shouldUpdate(8)
        ? new Chart(categoriesCanvas.getContext('2d'), {
            type: "bar",
            data: {
                labels: ['Categories'],
                datasets: wakapiData.categories
                    .slice(0, Math.min(showTopN[8], wakapiData.categories.length))
                    .map((p, i) => ({
                        label: p.key,
                        data: [parseInt(p.total)],
                        backgroundColor: vibrantColors ? getRandomColor(p.key) : getColor(p.key, i % baseColors.length),
                        barPercentage: 1.0
                    })),
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
                            callback: (label) => label.toString().toHHMMSS(),
                        },
                        stacked: true,
                        max: wakapiData.categories.map(c => c.total).reduce((a, b) => a + b, 0)
                    },
                    y: {
                        stacked: true,
                        display: false,
                    }
                },
                plugins: {
                    tooltip: getTooltipOptions('categories', true),
                },
                maintainAspectRatio: false,
            }
        })
        : null

    let timelineChart = timelineCanvas && !timelineCanvas.classList.contains('hidden') && shouldUpdate(9)
        ? new Chart(timelineCanvas.getContext('2d'), {
            type: 'bar',
            data: {
                labels: wakapiData.timelineStats.map(day => new Date(day.date).toLocaleDateString()),
                datasets: wakapiData.timelineStats
                    .flatMap(day => day.projects.map(project => project.name))
                    .sort()
                    .filter((value, index, self) => self.indexOf(value) === index)
                    .map((project, i) => ({
                        label: project,
                        data: wakapiData.timelineStats.map(day => day.projects.reduce((acc, p) => p.name === project ? acc + p.duration : acc, 0)),
                        backgroundColor: vibrantColors ? getRandomColor(project) : getColor(project, i % baseColors.length),
                        barPercentage: 1.0
                    }))
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        stacked: true,
                        title: {
                            display: true,
                            text: 'Date'
                        }
                    },
                    y: {
                        stacked: true,
                        title: {
                            display: true,
                            text: 'Duration (hh:mm:ss)'
                        },
                        ticks: {
                            callback: value => value.toString().toHHMMSS()
                        }
                    }
                },
                plugins: {
                    tooltip: {
                        callbacks: {
                            label: (context) => {
                                return `${context.dataset.label}: ${context.raw.toString().toHHMMSS()}`
                            }
                        }
                    },
                    legend: {
                        position: 'right',
                        labels: {
                            filter: filterLegendItem
                        }
                    }
                }
            }
        })
        : null

    // https://stackoverflow.com/a/73071513
    let hourlyBreakdownChart = hourlyCanvas && !hourlyCanvas.classList.contains('hidden') && shouldUpdate(10)
        ? new Chart(hourlyCanvas.getContext('2d'), {
            type: 'bar',
            data: {
                labels: wakapiData.hourlyBreakdown.map(i => i.project),
                datasets: wakapiData.hourlyBreakdown.flatMap((cur, i) => {
                    let project = cur.project
                    let pre = 0;
                    return cur.items.map((cur, _) => {
                        let data = wakapiData.hourlyBreakdown.map(() => null)
                        let fromTime = new Date(cur.from_time)
                        let toTime = new Date(fromTime.getTime() + (cur.duration / 1e9 * 1e3))

                        // "The values for the first bar of a stack are absolute values, all following values of the same stack must be relative to the end of the previous bar"
                        data[i] = [+fromTime - pre, +toTime - pre, `${fromTime.toLocaleTimeString()} - ${toTime.toLocaleTimeString()} (${(cur.duration / 1e9).toString().toHHMMSS()})`]
                        pre = +toTime
                        return {
                            data,
                            backgroundColor: vibrantColors ? getRandomColor(project) : getColor(project, i % baseColors.length),
                            label: cur.entity,
                            stack: project,
                            skipNull: true,
                        }
                    })
                })
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                indexAxis: 'y',
                scales: {
                    x: {
                        stacked: true,
                        min: +new Date(wakapiData.hourlyBreakdownFromTime),
                        max: +new Date(wakapiData.hourlyBreakdownToTime),
                        ticks: {
                            stepSize: 1000 * 60 * 60, // pre hour
                            callback: (value) => {
                                return new Date(value).toLocaleString([], {
                                    dateStyle: 'short',
                                    timeStyle: 'short',
                                })
                            },
                        },
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    },
                    y: {
                        stacked: true,
                        title: {
                            display: true,
                            text: 'Projects'
                        }
                    }
                },
                plugins: {
                    tooltip: {
                        callbacks: {
                            label: (context) => {
                                return `${context.dataset.label}: ${context.raw[2]}`
                            }
                        }
                    },
                    legend: {
                        display: false
                    },
                    zoom: {
                        pan: {
                            enabled: true,
                        },
                        zoom: {
                            wheel: {
                                enabled: true,
                                modifierKey: 'ctrl',
                            },
                            pinch: {
                                enabled: true,
                            },
                            // disabled drag-to-zoom (that is, select a range to zoom there) beacause it conflicts with the pan
                            // drag: {
                            //     enabled: true,
                            // },
                            mode: 'x',
                        },
                        limits: {
                            x: {
                                min: "original",
                                max: "original",
                            },
                        },
                    }
                }
            }
        })
        : null

    charts[0] = projectChart ? projectChart : charts[0]
    charts[1] = osChart ? osChart : charts[1]
    charts[2] = editorChart ? editorChart : charts[2]
    charts[3] = languageChart ? languageChart : charts[3]
    charts[4] = machineChart ? machineChart : charts[4]
    charts[5] = labelChart ? labelChart : charts[5]
    charts[6] = branchChart ? branchChart : charts[6]
    charts[7] = entityChart ? entityChart : charts[7]
    charts[8] = categoryChart ? categoryChart : charts[8]
    charts[9] = timelineChart ? timelineChart : charts[9]
    charts[10] = hourlyBreakdownChart ? hourlyBreakdownChart : charts[10]
}

function parseTopN() {
    showTopN = topNPickers.map(e => parseInt(e.value))
}

function togglePlaceholders(mask) {
    const placeholderElements = containers.map(c => c ? c.querySelector('.placeholder-container') : null)

    for (let i = 0; i < mask.length; i++) {
        if (placeholderElements[i] === null) {
            continue;
        }
        if (!mask[i]) {
            canvases[i].classList.add('hidden')
            placeholderElements[i].classList.remove('hidden')
        } else {
            canvases[i].classList.remove('hidden')
            placeholderElements[i].classList.add('hidden')
        }
    }
}

function getPresentDataMask() {
    return data.map(list => (list ? list.reduce((acc, e) => acc + (e.total ? e.total : ((e.projects || e.items) ? (e.projects || e.items).reduce((acc, f) => acc + f.duration, 0) : 0)), 0) : 0) > 0)
}

function getColor(seed, index) {
    if (index < baseColors.length) return baseColors[(index + 5) % baseColors.length]
    return getRandomColor(seed)
}

function getRandomColor(seed) {
    seed = seed ? seed : '1234567'
    Math.seedrandom(seed)
    var letters = '0123456789ABCDEF'.split('')
    var color = '#'
    for (var i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)]
    }
    return color
}

// https://stackoverflow.com/a/5624139/3112139
function hexToRgb(hex) {
    var shorthandRegex = /^#?([a-f\d])([a-f\d])([a-f\d])$/i;
    hex = hex.replace(shorthandRegex, function (m, r, g, b) {
        return r + r + g + g + b + b;
    });
    var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
    } : null;
}

function swapCharts(showEntity, hideEntity) {
    document.getElementById(`${showEntity}-container`).parentElement.classList.remove('hidden')
    document.getElementById(`${hideEntity}-container`).parentElement.classList.add('hidden')
}

function extractFile(filePath) {
    const delimiter = filePath.includes('\\') ? '\\' : '/'  // windows style path?
    return filePath.split(delimiter).at(-1)
}

function updateNumTotal() {
    // Why length - 2:
    //  We don't have a 'topN' for the DailyProjectStats & Timeline
    //  So there isn't a input for it.
    for (let i = 0; i < data.length - 2; i++) {
        document.querySelector(`span[data-entity='${i}']`).innerText = data[i].length.toString()
    }
}

window.addEventListener('load', function () {
    topNPickers.forEach(e => e.addEventListener('change', () => {
        parseTopN()
        draw([parseInt(e.attributes['data-entity'].value)])
    }))

    parseTopN()
    togglePlaceholders(getPresentDataMask())
    draw()
    updateNumTotal()
})