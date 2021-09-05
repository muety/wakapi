const CHART_TARGET_SIZE = 200

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

charts.color = "#E2E8F0"
charts.borderColor = "#E2E8F0"
charts.backgroundColor = "#E2E8F0"

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

String.prototype.toHHMM = function () {
    const sec_num = parseInt(this, 10)
    const hours = Math.floor(sec_num / 3600)
    const minutes = Math.floor((sec_num - (hours * 3600)) / 60)
    return `${hours}:${minutes}`
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
                    backgroundColor: wakapiData.projects.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.6)`
                    }),
                    hoverBackgroundColor: wakapiData.projects.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderColor: wakapiData.projects.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                    borderWidth: 2
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
                            text: 'Duration (hh:mm:ss)'
                        },
                        ticks: {
                            callback: (label) => label.toString().toHHMMSS(),
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: getTooltipOptions('projects'),
                },
                maintainAspectRatio: false,
                onResize: onChartResize
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
                    backgroundColor: wakapiData.operatingSystems.map(p => {
                        const c = hexToRgb(osColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.6)`
                    }),
                    hoverBackgroundColor: wakapiData.operatingSystems.map(p => {
                        const c = hexToRgb(osColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderColor: wakapiData.operatingSystems.map(p => {
                        const c = hexToRgb(osColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                }],
                labels: wakapiData.operatingSystems
                    .slice(0, Math.min(showTopN[1], wakapiData.operatingSystems.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('operatingSystems'),
                },
                maintainAspectRatio: false,
                onResize: onChartResize
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
                    backgroundColor: wakapiData.editors.map(p => {
                        const c = hexToRgb(editorColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.6)`
                    }),
                    hoverBackgroundColor: wakapiData.editors.map(p => {
                        const c = hexToRgb(editorColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderColor: wakapiData.editors.map(p => {
                        const c = hexToRgb(editorColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                }],
                labels: wakapiData.editors
                    .slice(0, Math.min(showTopN[2], wakapiData.editors.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('editors'),
                },
                maintainAspectRatio: false,
                onResize: onChartResize
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
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.6)`
                    }),
                    hoverBackgroundColor: wakapiData.languages.map(p => {
                        const c = hexToRgb(languageColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderColor: wakapiData.languages.map(p => {
                        const c = hexToRgb(languageColors[p.key.toLowerCase()] || getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                }],
                labels: wakapiData.languages
                    .slice(0, Math.min(showTopN[3], wakapiData.languages.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('languages'),
                },
                maintainAspectRatio: false,
                onResize: onChartResize
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
                    backgroundColor: wakapiData.machines.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.6)`
                    }),
                    hoverBackgroundColor: wakapiData.machines.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderColor: wakapiData.machines.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                }],
                labels: wakapiData.machines
                    .slice(0, Math.min(showTopN[4], wakapiData.machines.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('machines'),
                },
                maintainAspectRatio: false,
                onResize: onChartResize
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
                    backgroundColor: wakapiData.labels.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.6)`
                    }),
                    hoverBackgroundColor: wakapiData.labels.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 0.8)`
                    }),
                    borderColor: wakapiData.labels.map(p => {
                        const c = hexToRgb(getRandomColor(p.key))
                        return `rgba(${c.r}, ${c.g}, ${c.b}, 1)`
                    }),
                }],
                labels: wakapiData.labels
                    .slice(0, Math.min(showTopN[5], wakapiData.labels.length))
                    .map(p => p.key)
            },
            options: {
                plugins: {
                    tooltip: getTooltipOptions('labels'),
                },
                maintainAspectRatio: false,
                onResize: onChartResize
            }
        })
        : null

    getTotal(wakapiData.operatingSystems)

    charts[0] = projectChart ? projectChart : charts[0]
    charts[1] = osChart ? osChart : charts[1]
    charts[2] = editorChart ? editorChart : charts[2]
    charts[3] = languageChart ? languageChart : charts[3]
    charts[4] = machineChart ? machineChart : charts[4]
    charts[5] = labelChart ? labelChart : charts[5]

    if (!subselection) {
        charts.forEach(c => c.options.onResize(c))
        equalizeHeights()
    }
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

function onChartResize(chart) {
    let container = getContainer(chart)
    let targetHeight = Math.min(chart.width, CHART_TARGET_SIZE)
//    let actualHeight = chart.height - chart.chartArea.top
    let actualHeight = chart.height - chart.top
    let containerTargetHeight = container.clientHeight += (targetHeight - actualHeight)
    container.style.height = parseInt(containerTargetHeight) + 'px'

    resizeCount++
    watchEqualize()
}

function watchEqualize() {
    if (resizeCount === charts.length) {
        equalizeHeights()
        resizeCount = 0
    }
}

function equalizeHeights() {
    let maxHeight = 0
    charts.forEach(c => {
        let container = getContainer(c)
        if (maxHeight < container.clientHeight) {
            maxHeight = container.clientHeight
        }
    })
    charts.forEach(c => {
        let container = getContainer(c)
        container.style.height = parseInt(maxHeight) + 'px'
    })
}

function getTotal(items) {
    const el = document.getElementById('total-span')
    if (!el) return
    const total = items.reduce((acc, d) => acc + d.total, 0)
    const formatted = total.toString().toHHMM()
    el.innerText = `${formatted.split(':')[0]} hours, ${formatted.split(':')[1]} minutes`
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
    topNPickers.forEach(e => e.addEventListener('change', () => {
        parseTopN()
        draw([parseInt(e.attributes['data-entity'].value)])
    }))

    parseTopN()
    togglePlaceholders(getPresentDataMask())
    draw()
})

