package docker

type Language string

var SupportedLanguages = []Language{
	"python",
	// "typescript", "go", "javascript", "ruby", "rust", "swift", "cpp",
}

var LanguageToImageMap = map[Language]string{
	"python": "code-garden-python",
	//"javascript": "code-garden-node",
	//"typescript": "code-garden-node",
	//"go":         "code-garden-go",
	//"rust":       "code-garden-rust",
	//"swift":      "code-garden-swift",
	//"ruby":       "code-garden-ruby",
	//"cpp": "code-garden-cpp",
}

var LanguageToDockerFileMap = map[Language]string{
	"python": "Dockerfile_python",
	//"javascript": "Dockerfile_node",
	//"typescript": "Dockerfile_node",
	//"go":         "Dockerfile_go",
	//"rust":       "Dockerfile_rust",
	//"ruby":       "Dockerfile_ruby",
	//"swift":      "Dockerfile_swift",
	//"cpp": "Dockerfile_cpp",
	//"
}
