package entities

type Language string

var LanguageImageMap = map[Language]string{
	"python": "python:3.8",
	"node":   "node:14",
	"go":     "golang:1.16",
	"java":   "openjdk:11",
	"ruby":   "ruby:2.7",
}
