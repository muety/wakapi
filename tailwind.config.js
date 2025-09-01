const colors = require('tailwindcss/colors')

module.exports = {
    theme: {
        darkMode: 'class',
        extend: {
            colors: {
                green: colors.emerald,
                background: 'var(--background)',
                card: 'rgb(var(--card) / <alpha-value>)',

                muted: 'var(--muted)',

                secondary: 'var(--secondary)',
                primary: 'var(--primary)',
                foreground: 'var(--foreground)',
                focused: 'rgb(var(--focused) / <alpha-value>)',
                accent: 'rgb(var(--accent) / <alpha-value>)',
                danger: 'rgb(var(--danger) / <alpha-value>)',
            },
            width: {
                '16': '4rem',
            }
        }
    },
    content: [
        './views/*.tpl.html',
    ],
    safelist: [
        'newsbox-default',
        'newsbox-warning',
        'newsbox-danger',
        'leaderboard-self',
        'leaderboard-default',
        'leaderboard-gold',
        'leaderboard-silver',
        'leaderboard-bronze',
    ]
}
