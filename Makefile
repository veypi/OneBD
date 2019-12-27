version:
	grep -P 'Version = ' onebd.go|cut -f2 -d'"'

tag:
	@git tag `grep -P 'Version = ' onebd.go|cut -f2 -d'"'`
	@git tag | grep -v ^v
