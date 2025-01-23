package docker

type Language string

var SupportedLanguages = []Language{
	"python", "typescript", "go", "javascript",
}

var LanguageToImageMap = map[Language]string{
	"python":     "code-garden-python",
	"javascript": "code-garden-node",
	"typescript": "code-garden-node",
	"go":         "code-garden-go",
}

var LanguageToDockerFileMap = map[Language]string{
	"python": "Dockerfile_python",
	"javascript":   "Dockerfile_node",
	"typescript":   "Dockerfile_node",
	"go":     "Dockerfile_go",
	//"java":   "Dockerfile.java",
	//"ruby":   "Dockerfile.ruby",
}
