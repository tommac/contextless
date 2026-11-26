// This file keeps idea-vault/ out of the Go module github.com/tommac/contextless.
// Note titles here are sentences, and the module format refuses some
// characters in file names (an apostrophe, for one). With those files inside
// the module, "go install github.com/tommac/contextless/cmd/cx@latest" failed.
// There is no Go code in this folder.
module github.com/tommac/contextless/idea-vault

go 1.26.0
