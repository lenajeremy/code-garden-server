package docker

type Language string

var SupportedLanguages = []Language{
	"python", "node", "go",
}

var LanguageImageMap = map[Language]string{
	"python": "python:3.8",
	"node":   "node:14",
	"go":     "golang:1.16",
	"java":   "openjdk:11",
	"ruby":   "ruby:2.7",
}

var LanguageToDockerFileMap = map[Language]string{
	"python": "Dockerfile_python",
	"node":   "Dockerfile_node",
	"go":     "Dockerfile_go",
	//"java":   "Dockerfile.java",
	//"ruby":   "Dockerfile.ruby",
}
