#!/usr/bin/env node

'use strict'

// Usage:
// yarn add -D @iconify/json-tools @iconify/json
// node bundle_icons.js
// https://iconify.design/docs/icon-bundles/

const fs = require('fs')
const path = require('path')
const { Collection } = require('@iconify/json-tools')

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
    'ic:round-person'
]

const output = path.normalize(path.join(__dirname, '../static/assets/icons.js'))
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
    if (!collection.loadIconifyCollection(prefix)) {
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