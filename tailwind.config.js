const colors = require('tailwindcss/colors')

module.exports = {
    theme: {
        darkMode: 'class',
        extend: {
            colors: {
                green: colors.emerald,
                background: 'rgb(var(--background) / <alpha-value>)',
                card: 'rgb(var(--card) / <alpha-value>)',
                muted: 'rgb(var(--muted) / <alpha-value>)',
                secondary: 'rgb(var(--secondary) / <alpha-value>)',
                primary: 'rgb(var(--primary) / <alpha-value>)',
                foreground: 'rgb(var(--foreground) / <alpha-value>)',
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
