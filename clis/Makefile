#
# Makefile
# Copyright (C) 2024 veypi <i@veypi.com>
# 2024-08-13 17:16
# Distributed under terms of the MIT license.
#


cache=~/test/app/

cmd=go run clis.go -l debug --dir ${cache} -y --repo testapp

test:clean
	@${cmd} app  new
	@${cmd} model new
	@${cmd} model new -n app/app
	@${cmd} model new -n user
	@${cmd} model new -n user.cache
	@${cmd} model gen
	@${cmd} api new -n single
	@${cmd} api gen

clean:
	@rm -rf ${cache}/*
