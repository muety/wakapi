const LEGEND_MAX_ENTRIES = 9
// dirty hack to vertically align legends across multiple charts
// however, without monospace font, it's still not perfectly aligned
// waiting for https://github.com/chartjs/Chart.js/discussions/9890
const LEGEND_CHARACTERS = 20

// https://hihayk.github.io/scale/#4/6/50/80/-51/67/20/14/276749/39/103/73/white
const baseColors = [ '#112836', '#163B43', '#1C4F4D', '#215B4C', '#276749', '#437C57', '#5F9167', '#7DA67C', '#9FBA98', '#BFCEB5', '#DCE2D3' ]

const projectsCanvas = document.getElementById('chart-projects')
const osCanvas = document.getElementById('chart-os')
const editorsCanvas = document.getElementById('chart-editor')
const languagesCanvas = document.getElementById('chart-language')
const machinesCanvas = document.getElementById('chart-machine')
const labelsCanvas = document.getElementById('chart-label')

const projectContainer = document.getElementById('project-container')
const osContainer = document.getElementById('os-container')
const editorContainer = document.getElementById('editor-container')
const languageContainer = document.getElementById('language-container')
const machineContainer = document.getElementById('machine-container')
const labelContainer = document.getElementById('label-container')

const containers = [projectContainer, osContainer, editorContainer, languageContainer, machineContainer, labelContainer]
const canvases = [projectsCanvas, osCanvas, editorsCanvas, languagesCanvas, machinesCanvas, labelsCanvas]
const data = [wakapiData.projects, wakapiData.operatingSystems, wakapiData.editors, wakapiData.languages, wakapiData.machines, wakapiData.labels]

let topNPickers = [...document.getElementsByClassName('top-picker')]
topNPickers.sort(((a, b) => parseInt(a.attributes['data-entity'].value) - parseInt(b.attributes['data-entity'].value)))
topNPickers.forEach(e => {
    const idx = parseInt(e.attributes['data-entity'].value)
    e.max = Math.min(data[idx].length, 10)
    e.value = e.max
})

let charts = []
let showTopN = []
let resizeCount = 0

Chart.defaults.color = "#E2E8F0"
Chart.defaults.borderColor = "#242b3a"

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

function draw(subselection) {
    function getTooltipOptions(key) {
        return {
            callbacks: {
                label: (item) => {
                    const d = wakapiData[key][item.dataIndex]
                    return `${d.key}: ${d.total.toString().toHHMMSS()}`
                },
                title: () => 'Total Time'
            }
        }
    }

    function filterLegendItem(item) {
        item.text = item.text.length > LEGEND_CHARACTERS ? item.text.slice(0, LEGEND_CHARACTERS - 3).padEnd(LEGEND_CHARACTERS, '.') : item.text
        item.text = item.text.padEnd(LEGEND_CHARACTERS + 3)
        return item.index < LEGEND_MAX_ENTRIES
    }

    function shouldUpdate(index) {
        return !subselection || (subselection.includes(index) && data[index].length >= showTopN[index])
    }

    charts
        .filter((c, i) => shouldUpdate(i))
        .forEach(c => c.destroy())

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
                        const c = hexToRgb(getColor(p.key, i % baseColors.length))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.projects.map((p, i) => {
                        const c = hexToRgb(getColor(p.key, i % baseColors.length))
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
                        const c = hexToRgb(getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.operatingSystems.map((p, i) => {
                        const c = hexToRgb(getColor(p.key, i))
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
                        const c = hexToRgb(getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.editors.map((p, i) => {
                        const c = hexToRgb(getColor(p.key, i))
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
                        const c = hexToRgb(getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.machines.map((p, i) => {
                        const c = hexToRgb(getColor(p.key, i))
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
                        const c = hexToRgb(getColor(p.key, i))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    hoverBackgroundColor: wakapiData.labels.map((p, i) => {
                        const c = hexToRgb(getColor(p.key, i))
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

    charts[0] = projectChart ? projectChart : charts[0]
    charts[1] = osChart ? osChart : charts[1]
    charts[2] = editorChart ? editorChart : charts[2]
    charts[3] = languageChart ? languageChart : charts[3]
    charts[4] = machineChart ? machineChart : charts[4]
    charts[5] = labelChart ? labelChart : charts[5]
}

function parseTopN() {
    showTopN = topNPickers.map(e => parseInt(e.value))
}

function togglePlaceholders(mask) {
    const placeholderElements = containers.map(c => c ? c.querySelector('.placeholder-container'): null)

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
    return data.map(list => (list ? list.reduce((acc, e) => acc + e.total, 0) : 0) > 0)
}

function getContainer(chart) {
    // See https://github.com/muety/wakapi/issues/235#issuecomment-907762100
    return chart.canvas.parentNode
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
    hex = hex.replace(shorthandRegex, function(m, r, g, b) {
        return r + r + g + g + b + b;
    });
    var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
    } : null;
}

function showUserMenuPopup(event) {
    const el = document.getElementById('user-menu-popup')
    el.classList.remove('hidden')
    el.classList.add('block')
    event.stopPropagation()
}

function hideUserMenuPopup(event) {
    const el = document.getElementById('user-menu-popup')
    el.classList.remove('block')
    el.classList.add('hidden')
    event.stopPropagation()
}

function toggleTimePickerPopup(event) {
    const el = document.getElementById('time-picker-popup')
    if (el.classList.contains('hidden')) {
        el.classList.remove('hidden')
        el.classList.add('block')
    } else {
        el.classList.add('hidden')
        el.classList.remove('block')
    }
    event.stopPropagation()
}

function showApiKeyPopup(event) {
    const el = document.getElementById('api-key-popup')
    el.classList.remove('hidden')
    el.classList.add('block')
    event.stopPropagation()
}

function copyApiKey(event) {
    const el = document.getElementById('api-key-container')
    el.select()
    el.setSelectionRange(0, 9999)
    document.execCommand('copy')
    event.stopPropagation()
}

function submitTimePicker(event) {
    const el = document.getElementById('time-picker-form')
    el.submit()
}

function swapCharts(showEntity, hideEntity) {
    document.getElementById(`${showEntity}-container`).parentElement.classList.remove('hidden')
    document.getElementById(`${hideEntity}-container`).parentElement.classList.add('hidden')
}

function updateTimeSelection() {
    const query = new URLSearchParams(window.location.search)
    if (query.has('interval')) {
        const targetEl = document.getElementById('current-time-selection')
        const refEl = document.getElementById(`time-option-${query.get('interval')}`)
        targetEl.innerText = refEl ? refEl.innerText : 'Unknown'
    }
}



// Click outside
window.addEventListener('click', function (event) {
    if (event.target.classList.contains('popup')) {
        return
    }
    document.querySelectorAll('.popup').forEach(el => {
        el.classList.remove('block')
        el.classList.add('hidden')
    })
})

window.addEventListener('load', function () {
    updateTimeSelection()

    topNPickers.forEach(e => e.addEventListener('change', () => {
        parseTopN()
        draw([parseInt(e.attributes['data-entity'].value)])
    }))

    parseTopN()
    togglePlaceholders(getPresentDataMask())
    draw()
})

