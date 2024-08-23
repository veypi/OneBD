#
# Makefile
# Copyright (C) 2024 veypi <i@veypi.com>
# 2024-08-13 17:16
# Distributed under terms of the MIT license.
#


cache=~/test/app/

cmd=go run clis.go -l info --dir ${cache} -y --repo testapp

test:clean
	@${cmd} model new -n app
	@${cmd} model gen
	@${cmd} app  new
	@${cmd} api new -n user.app
	@${cmd} api new -n user.auth

clean:
	@rm -rf ${cache}/*
