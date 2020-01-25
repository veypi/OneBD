version:
	@grep -p 'Version = ' onebd.go|cut -f2 -d'"'

tag:
	@awk -F '"' '/Version/ {print $$2;system("git tag "$$2);system("git push origin "$$2)}' onebd.go

dropTag:
	@awk -F '"' '/Version/ {print $$2;system("git tag -d "$$2);system("git push origin :refs/tags/"$$2)}' onebd.go

