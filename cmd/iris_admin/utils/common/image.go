package common

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Mrs4s/go-cqhttp/internal/download"
	"html/template"
	"os"
	"path"
	"strings"

	tmpl "github.com/GoAdminGroup/go-admin/template"
	"github.com/Mrs4s/MiraiGo/binary"
	log "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/global"
	"github.com/Mrs4s/go-cqhttp/internal/cache"
)

// GetImageWithCache 从缓存获取图片信息的模板
func GetImageWithCache(data global.MSG) template.HTML {
	var b []byte
	var err error
	url := data["url"].(string)
	file := data["file"].(string)
	if strings.HasSuffix(file, ".image") {
		var f []byte
		f, err = hex.DecodeString(strings.TrimSuffix(file, ".image"))
		b = cache.Image.Get(f)
	}

	if b == nil {
		if !global.PathExists(path.Join(global.ImagePath, file)) {
			return returnImagefaild(url)
		}
		b, err = os.ReadFile(path.Join(global.ImagePath, file))
	}
	if err == nil {
		r := binary.NewReader(b)
		r.ReadBytes(16)
		msg := global.MSG{
			"size":     r.ReadInt32(),
			"filename": r.ReadString(),
			"url":      r.ReadString(),
		}
		local := path.Join(global.CachePath, file+path.Ext(msg["filename"].(string)))
		if !global.PathExists(local) {
			d, err := download.Request{URL: msg["url"].(string)}.Bytes()
			if err == nil {
				f, _ := os.OpenFile(local, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o0644)
				_, _ = f.Write(d)
				_ = f.Close()
			} else {
				log.Warnf("下载图片 %v 时出现错误: %v", msg["url"], err)
				return returnImagefaild(url)
			}
		}
		msg["file"] = local
		img, _ := os.ReadFile(local)
		return returnImageSuccess(img)
	}
	return returnImagefaild(url)
}

func returnImageSuccess(data []byte) template.HTML {
	imgsrcBase64 := fmt.Sprintf("data:image/gif;base64,%s", base64.StdEncoding.EncodeToString(data))
	return tmpl.HTML(fmt.Sprintf(`<img src="%s" height="120"/><br/>`, imgsrcBase64))
}

func returnImagefaild(url string) template.HTML {
	return tmpl.HTML(fmt.Sprintf(`[CQ:image,url=<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>]`, url, url))
}
