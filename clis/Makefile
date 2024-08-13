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

clean:
	@rm -rf ${cache}/*
