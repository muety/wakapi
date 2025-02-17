const colors = require("tailwindcss/colors");

module.exports = {
  theme: {
    extend: {
      keyframes: {
        "caret-blink": {
          "0%,70%,100%": { opacity: "1" },
          "20%,50%": { opacity: "0" },
        },
      },
      colors: {
        green: colors.emerald,
      },
      width: {
        16: "4rem",
      },
      animation: {
        "caret-blink": "caret-blink 1.25s ease-out infinite",
      },
    },
  },
  content: ["./views/*.tpl.html"],
  safelist: [
    "newsbox-default",
    "newsbox-warning",
    "newsbox-danger",
    "leaderboard-self",
    "leaderboard-default",
    "leaderboard-gold",
    "leaderboard-silver",
    "leaderboard-bronze",
  ],
};
