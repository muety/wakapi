const SHOW_TOP_N = 10
const CHART_TARGET_SIZE = 200

const projectsCanvas = document.getElementById('chart-projects')
const osCanvas = document.getElementById('chart-os')
const editorsCanvas = document.getElementById('chart-editor')
const languagesCanvas = document.getElementById('chart-language')
const machinesCanvas = document.getElementById('chart-machine')

const projectContainer = document.getElementById('project-container')
const osContainer = document.getElementById('os-container')
const editorContainer = document.getElementById('editor-container')
const languageContainer = document.getElementById('language-container')
const machineContainer = document.getElementById('machine-container')

const containers = [projectContainer, osContainer, editorContainer, languageContainer, machineContainer]
const canvases = [projectsCanvas, osCanvas, editorsCanvas, languagesCanvas, machinesCanvas]
const data = [wakapiData.projects, wakapiData.operatingSystems, wakapiData.editors, wakapiData.languages, wakapiData.machines]

let charts = []
let resizeCount = 0

String.prototype.toHHMMSS = function () {
    var sec_num = parseInt(this, 10)
    var hours = Math.floor(sec_num / 3600)
    var minutes = Math.floor((sec_num - (hours * 3600)) / 60)
    var seconds = sec_num - (hours * 3600) - (minutes * 60)

    if (hours < 10) {
        hours = '0' + hours
    }
    if (minutes < 10) {
        minutes = '0' + minutes
    }
    if (seconds < 10) {
        seconds = '0' + seconds
    }
    return hours + ':' + minutes + ':' + seconds
}

function draw() {
    function getTooltipOptions(key, type) {
        return {
            mode: 'single',
            callbacks: {
                label: (item) => {
                    let idx = type === 'pie' ? item.index : item.datasetIndex
                    let d = wakapiData[key][idx]
                    return `${d.key}: ${d.total.toString().toHHMMSS()}`
                }
            }
        }
    }

    charts.forEach(c => c.destroy())

    let projectChart = !projectsCanvas.classList.contains('hidden')
        ? new Chart(projectsCanvas.getContext('2d'), {
            type: 'horizontalBar',
            data: {
                datasets: wakapiData.projects
                    .slice(0, Math.min(SHOW_TOP_N, wakapiData.projects.length))
                    .map(p => {
                        return {
                            label: p.key,
                            data: [parseInt(p.total) / 60],
                            backgroundColor: getRandomColor(p.key)
                        }
                    })
            },
            options: {
                tooltips: getTooltipOptions('projects', 'bar'),
                legend: {
                    display: false
                },
                scales: {
                    xAxes: [{
                        scaleLabel: {
                            display: true,
                            labelString: 'Minutes'
                        }
                    }]
                },
                maintainAspectRatio: false,
                onResize: onChartResize
            }
        })
        : null

    let osChart = !osCanvas.classList.contains('hidden')
        ? new Chart(osCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.operatingSystems
                        .slice(0, Math.min(SHOW_TOP_N, wakapiData.operatingSystems.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.operatingSystems.map(p => getRandomColor(p.key))
                }],
                labels: wakapiData.operatingSystems
                    .slice(0, Math.min(SHOW_TOP_N, wakapiData.operatingSystems.length))
                    .map(p => p.key)
            },
            options: {
                tooltips: getTooltipOptions('operatingSystems', 'pie'),
                maintainAspectRatio: false,
                onResize: onChartResize
            }
        })
        : null

    let editorChart = !editorsCanvas.classList.contains('hidden')
        ? new Chart(editorsCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.editors
                        .slice(0, Math.min(SHOW_TOP_N, wakapiData.editors.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.editors.map(p => getRandomColor(p.key))
                }],
                labels: wakapiData.editors
                    .slice(0, Math.min(SHOW_TOP_N, wakapiData.editors.length))
                    .map(p => p.key)
            },
            options: {
                tooltips: getTooltipOptions('editors', 'pie'),
                maintainAspectRatio: false,
                onResize: onChartResize
            }
        })
        : null

    let languageChart = !languagesCanvas.classList.contains('hidden')
        ? new Chart(languagesCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.languages
                        .slice(0, Math.min(SHOW_TOP_N, wakapiData.languages.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.languages.map(p => languageColors[p.key.toLowerCase()] || getRandomColor(p.key))
                }],
                labels: wakapiData.languages
                    .slice(0, Math.min(SHOW_TOP_N, wakapiData.languages.length))
                    .map(p => p.key)
            },
            options: {
                tooltips: getTooltipOptions('languages', 'pie'),
                maintainAspectRatio: false,
                onResize: onChartResize
            }
        })
        : null

    let machineChart = !machinesCanvas.classList.contains('hidden')
        ? new Chart(machinesCanvas.getContext('2d'), {
            type: 'pie',
            data: {
                datasets: [{
                    data: wakapiData.machines
                        .slice(0, Math.min(SHOW_TOP_N, wakapiData.machines.length))
                        .map(p => parseInt(p.total)),
                    backgroundColor: wakapiData.machines.map(p => getRandomColor(p.key))
                }],
                labels: wakapiData.machines
                    .slice(0, Math.min(SHOW_TOP_N, wakapiData.machines.length))
                    .map(p => p.key)
            },
            options: {
                tooltips: getTooltipOptions('machines', 'pie'),
                maintainAspectRatio: false,
                onResize: onChartResize
            }
        })
        : null

    getTotal(wakapiData.operatingSystems)

    charts = [projectChart, osChart, editorChart, languageChart, machineChart].filter(c => !!c)

    charts.forEach(c => c.options.onResize(c.chart))
    equalizeHeights()
}

function setTopLabels() {
    [...document.getElementsByClassName('top-label')]
        .forEach(e => e.innerText = `(top ${SHOW_TOP_N})`)
}

function togglePlaceholders(mask) {
    const placeholderElements = containers.map(c => c.querySelector('.placeholder-container'))

    for (let i = 0; i < mask.length; i++) {
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
    return data.map(list => list ? list.reduce((acc, e) => acc + e.total, 0) : 0 > 0)
}

function getContainer(chart) {
    return chart.canvas.parentNode
}

function onChartResize(chart) {
    let container = getContainer(chart)
    let targetHeight = Math.min(chart.width, CHART_TARGET_SIZE)
    let actualHeight = chart.height - chart.chartArea.top
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
    let total = items.reduce((acc, d) => acc + d.total, 0)
    document.getElementById('total-span').innerText = total.toString().toHHMMSS()
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

// https://koddsson.com/posts/emoji-favicon/
const favicon = document.querySelector('link[rel=icon]')
if (favicon) {
    const emoji = favicon.getAttribute('data-emoji')
    if (emoji) {
        const canvas = document.createElement('canvas')
        canvas.height = 64
        canvas.width = 64
        const ctx = canvas.getContext('2d')
        ctx.font = '64px serif'
        ctx.fillText(emoji, 0, 64)
        favicon.href = canvas.toDataURL()
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
    setTopLabels()
    togglePlaceholders(getPresentDataMask())
    draw()
})
