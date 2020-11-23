site:V: index.html manual.html contributors.html license.html


index.html: README.md
    pandoc -s --to html5 --from markdown -c style.css --metadata title="Mk: make remade" -o $target <(sed -e 's/\.md/.html/g' $prereq)

manual.html: mk.1.md
	pandoc -s --to html5 --from markdown -c style.css --metadata title="Mk(1)" -o $target <(sed -e 's/\.md/.html/g' $prereq)

%.html: %.md
	pandoc -s --to html5 --from markdown -c style.css -o $target --metadata title="Mk: $(basename $target .html)" <(sed -e 's/\.md/.html/g' $prereq)