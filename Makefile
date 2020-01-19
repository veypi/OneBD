version:
	@grep -p 'Version = ' onebd.go|cut -f2 -d'"'

tag:
	@git tag `grep -p 'Version = ' onebd.go|cut -f2 -d'"'`
	@git tag | grep -v ^v
