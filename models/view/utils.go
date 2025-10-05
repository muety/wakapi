package view

import "strings"

func GetLanguageIcon(language string) string {
	// https://icon-sets.iconify.design/mdi/
	langs := map[string]string{
		"c++":        "language-cpp",
		"cpp":        "language-cpp",
		"go":         "language-go",
		"haskell":    "language-haskell",
		"html":       "language-html5",
		"java":       "language-java",
		"javascript": "language-javascript",
		"jsx":        "language-javascript",
		"kotlin":     "language-kotlin",
		"lua":        "language-lua",
		"php":        "language-php",
		"python":     "language-python",
		"r":          "language-r",
		"ruby":       "language-ruby",
		"rust":       "language-rust",
		"swift":      "language-swift",
		"typescript": "language-typescript",
		"tsx":        "language-typescript",
		"markdown":   "language-markdown",
		"vue":        "vuejs",
		"react":      "react",
		"bash":       "bash",
		"json":       "code-json",
		"nix":        "nix",
	}
	if match, ok := langs[strings.ToLower(language)]; ok {
		return "mdi:" + match
	}
	return ""
}

func GetOidcProviderIcon(provider string) string {
	providers := map[string]string{
		"github":    "codicon:github-inverted",
		"gitlab":    "devicon-plain:gitlab",
		"codeberg":  "devicon-plain:codeberg",
		"google":    "devicon-plain:google",
		"microsoft": "mdi:microsoft",
		"facebook":  "devicon-plain:facebook",
		"okta":      "devicon-plain:okta",
	}
	match, _ := providers[strings.ToLower(provider)]
	return match
}
