//
// static.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-12-11 15:37
// Distributed under terms of the MIT license.
//

package middlewares

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils/logv"
)

func Hanlder404(x *rest.X, err error) error {
	if err != nil {
		x.WriteHeader(404)
		x.Write([]byte(err.Error()))
	}
	return nil
}

// need to define *path in url variable
// like: app.Router('/*path',http.MethodGet,static.Static("./static", "./404.html"))
func Static(directory string, file404 string) func(*rest.X) {
	dir, err := os.Stat(directory)
	if err != nil {
		logv.Panic().Err(err).Send()
		return nil
	}
	if !dir.IsDir() {
		return func(x *rest.X) {
			f, err := os.Open(directory)
			if err != nil {
				panic(err)
			}
			info, err := f.Stat()
			if err != nil {
				x.WriteHeader(http.StatusNotFound)
				return
			}
			x.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(info.Name())))
			http.ServeContent(x, x.Request, info.Name(), info.ModTime(), f)
		}
	}
	var fs http.FileSystem = http.Dir(directory)
	return func(x *rest.X) {
		name := strings.TrimSuffix(x.Params.GetStr("path"), "/")
		f, info, err := handleDirOpen(fs.Open(name))
		if file404 != "" && err != nil {
			// handler name/+ ./404.html ./index.html
			if file404[0] == '.' {
				f, info, err = handleDirOpen(fs.Open(name + file404[1:]))
			} else {
				f, info, err = handleDirOpen(fs.Open(file404))
			}
		}
		if err != nil {
			x.WriteHeader(http.StatusNotFound)
			logv.Debug().Err(err).Send()
			return
		}
		defer f.Close()
		x.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(info.Name())))
		http.ServeContent(x, x.Request, info.Name(), info.ModTime(), f.(io.ReadSeeker))
	}
}

func handleDirOpen(f fs.File, err error) (fs.File, fs.FileInfo, error) {
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	if info.IsDir() {
		f.Close()
		return nil, nil, fs.ErrNotExist
	}
	return f, info, nil
}

func EmbedFile(f []byte, contentType string) func(*rest.X) {
	return func(x *rest.X) {
		x.Header().Set("Content-Type", contentType)
		_, err := x.Write(f)
		if err != nil {
			logv.Warn().Msgf("write file failed: %s", err.Error())
		}
	}
}

// need to define *path in url variable
func EmbedDir(dir embed.FS, fsPrefix string, file404 string) func(*rest.X) {
	if !strings.HasSuffix(fsPrefix, "/") {
		fsPrefix += "/"
	}
	return func(x *rest.X) {
		name := strings.TrimSuffix(fsPrefix+x.Params.GetStr("path"), "/")
		f, info, err := handleDirOpen(dir.Open(name))
		if file404 != "" && err != nil {
			// handler name/+ ./404.html ./index.html
			if file404[0] == '.' {
				f, info, err = handleDirOpen(dir.Open(name + file404[1:]))
			} else {
				f, info, err = handleDirOpen(dir.Open(fsPrefix + file404))
			}
		}
		if err != nil {
			x.ResponseWriter().WriteHeader(http.StatusNotFound)
			logv.Debug().Err(err).Send()
			return
		}
		defer f.Close()
		x.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(info.Name())))
		http.ServeContent(x, x.Request, info.Name(), info.ModTime(), f.(io.ReadSeeker))
	}
}
