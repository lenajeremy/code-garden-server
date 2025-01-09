package docker

type Language string

var SupportedLanguages = []Language{
	"python", "node", "go",
}

var LanguageToImageMap = map[Language]string{
	"python": "code-garden-python",
	"node":   "code-garden-node",
	"go":     "code-garden-go",
	"java":   "openjdk:11",
	"ruby":   "ruby:2.7",
	"swift": "swiftlang-6.0.0",
}

var LanguageToDockerFileMap = map[Language]string{
	"python": "Dockerfile_python",
	"node":   "Dockerfile_node",
	"go":     "Dockerfile_go",
	//"java":   "Dockerfile.java",
	//"ruby":   "Dockerfile.ruby",
}
