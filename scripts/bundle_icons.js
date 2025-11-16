#!/usr/bin/env node

'use strict'

// Usage:
// yarn add -D @iconify/json-tools @iconify/json
// node bundle_icons.js
// https://iconify.design/docs/icon-bundles/

const fs = require('fs')
const path = require('path')
const { Collection } = require('@iconify/json-tools')
const { locate } = require("@iconify/json");

let icons = [
    'fxemoji:key',
    'fxemoji:rocket',
    'fxemoji:satelliteantenna',
    'fxemoji:lockandkey',
    'fxemoji:clipboard',
    'flat-color-icons:donate',
    'flat-color-icons:clock',
    'codicon:github-inverted',
    'ant-design:check-square-filled',
    'emojione-v1:white-heavy-check-mark',
    'emojione-v1:alarm-clock',
    'emojione-v1:warning',
    'emojione-v1:backhand-index-pointing-right',
    'twemoji:light-bulb',
    'noto:play-button',
    'noto:stop-button',
    'noto:lock',
    'twemoji:gear',
    'eva:corner-right-down-fill',
    'bi:heart-fill',
    'fxemoji:running',
    'ic:round-person',
    'bx:bxs-bar-chart-alt-2',
    'bi:people-fill',
    'fluent:data-bar-horizontal-24-filled',
    'ic:round-dashboard',
    'ci:settings-filled',
    'akar-icons:chevron-down',
    'ls:logout',
    'fluent:key-32-filled',
    'majesticons:clipboard-copy',
    'fa-regular:calendar-alt',
    'ph:books-bold',
    'fa-solid:external-link-alt',
    'bx:bx-code-curly',
    'simple-icons:wakatime',
    'bx:bxs-heart',
    'heroicons-solid:light-bulb',
    'ion:rocket',
    'heroicons-solid:server',
    'eva:checkmark-circle-2-fill',
    'fluent:key-24-filled',
    'mdi:language-c',
    'mdi:language-cpp',
    'mdi:language-go',
    'mdi:language-haskell',
    'mdi:language-html5',
    'mdi:language-java',
    'mdi:language-javascript',
    'mdi:language-kotlin',
    'mdi:language-lua',
    'mdi:language-php',
    'mdi:language-python',
    'mdi:language-r',
    'mdi:language-ruby',
    'mdi:language-rust',
    'mdi:language-swift',
    'mdi:language-typescript',
    'mdi:language-markdown',
    'mdi:vuejs',
    'mdi:react',
    'mdi:code-json',
    'mdi:bash',
    'mdi:nix',
    'twemoji:frowning-face',
    'ci:dot-03-m',
    'jam:crown-f',
    'octicon:project-16',
    'octicon:share-16',
    'mdi:filter',
    'mdi:invite',
    'octicon:info-16',
    'devicon-plain:codeberg',
    'devicon-plain:google',
    'devicon-plain:gitlab',
    'devicon-plain:okta',
    'devicon-plain:facebook',
    'mdi:microsoft',
    'twemoji:flag-germany',
    'ic:round-download',
    'ic:outline-integration-instructions',
]

const output = path.normalize(path.join(__dirname, '../static/assets/js/icons.dist.js'))
const pretty = false

// Sort icons by collections: filtered[prefix][array of icons]
let filtered = {}
icons.forEach(icon => {
    let parts = icon.split(':'),
        prefix

    if (parts.length > 1) {
        prefix = parts.shift()
        icon = parts.join(':')
    } else {
        parts = icon.split('-')
        prefix = parts.shift()
        icon = parts.join('-')
    }
    if (filtered[prefix] === void 0) {
        filtered[prefix] = []
    }
    if (filtered[prefix].indexOf(icon) === -1) {
        filtered[prefix].push(icon)
    }
})

// Parse each collection
let code = ''
Object.keys(filtered).forEach(prefix => {
    let collection = new Collection()
    if (!collection.loadFromFile(locate(prefix))) {
        console.error('Error loading collection', prefix)
        return
    }

    code += collection.scriptify({
        icons: filtered[prefix],
        optimize: true,
        pretty: pretty
    })
})

// Save code
fs.writeFileSync(output, code, 'utf8')
console.log('Saved bundle to', output, ' (' + code.length + ' bytes)')
