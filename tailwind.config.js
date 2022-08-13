module.exports = {
    purge: {
        enabled: true,
        mode: 'all',
        content: ['./views/*.tpl.html'],
        safelist: [
            'newsbox-default',
            'newsbox-warning',
            'newsbox-danger',
        ]
    },
}