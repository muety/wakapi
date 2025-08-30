const colors = require('tailwindcss/colors')

module.exports = {
    theme: {
        darkMode: 'class',
        extend: {
            colors: {
                green: colors.emerald,
                background: 'var(--background)',
                card: 'var(--card)',

                muted: 'var(--muted)',

                secondary: 'var(--secondary)',
                primary: 'var(--primary)',
                foreground: 'var(--foreground)',
                accent: 'var(--accent)',
                accentHover: 'var(--accentHover)',
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
