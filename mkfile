sources = expand.go graph.go lex.go mk.go recipe.go remote.go rules.go

mk: ${sources}
    go build 

test:V:
    go test

mk.1: mk.1.md
    pandoc -s -t man -o $target $prereq