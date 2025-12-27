package server

import (
	"bytes"
	"net/http"
	"steam-fast-api/internal/cache"
	"steam-fast-api/internal/models"
	"steam-fast-api/internal/proxy"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func writeJSONString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '"' || b == '\\' {
			buf.WriteByte('\\')
			buf.WriteByte(b)
		} else if b < 0x20 {
			switch b {
			case '\n':
				buf.WriteString(`\n`)
			case '\r':
				buf.WriteString(`\r`)
			case '\t':
				buf.WriteString(`\t`)
			default:
				buf.WriteString(`\u00`)
				buf.WriteByte("0123456789abcdef"[b>>4])
				buf.WriteByte("0123456789abcdef"[b&0xF])
			}
		} else {
			buf.WriteByte(b)
		}
	}
	buf.WriteByte('"')
}

func HandleLookup(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	id := uint32(id64)
	reg := cache.Current.Load()
	name, appType, exists := reg.Get(id)
	var meta *proxy.AppMetadata
	if !exists {
		meta, err = proxy.GetMetadata(id)
		if err != nil {
			w.WriteHeader(404)
			return
		}
		name, appType = meta.Name, meta.AppType
		reg.Discover(id, name, appType)
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteByte('[')
	buf.WriteString(strconv.FormatUint(uint64(id), 10))
	buf.WriteByte(',')
	writeJSONString(buf, name)
	if appType&models.TypeGame != 0 {
		var dlcIDs []uint32
		if cached, ok := reg.Links.Load(id); ok {
			dlcIDs = cached.([]uint32)
		} else {
			if meta == nil {
				meta, _ = proxy.GetMetadata(id)
			}
			if meta != nil {
				dlcIDs = meta.DLCList
				reg.Links.Store(id, dlcIDs)
			}
		}
		if len(dlcIDs) > 0 {
			buf.WriteString(",[")
			for i, dID := range dlcIDs {
				if i > 0 {
					buf.WriteByte(',')
				}
				dName, _, dExists := reg.Get(dID)
				if !dExists {
					if dMeta, err := proxy.GetMetadata(dID); err == nil {
						dName = dMeta.Name
						reg.Discover(dID, dName, dMeta.AppType)
					}
				}
				buf.WriteByte('[')
				buf.WriteString(strconv.FormatUint(uint64(dID), 10))
				buf.WriteByte(',')
				writeJSONString(buf, dName)
				buf.WriteByte(']')
			}
			buf.WriteByte(']')
		}
	} else if appType&models.TypeDLC != 0 {
		var pID uint32
		if cached, ok := reg.Links.Load(id); ok {
			pID = cached.(uint32)
		} else {
			if meta == nil {
				meta, _ = proxy.GetMetadata(id)
			}
			if meta != nil {
				pID = meta.ParentID
				reg.Links.Store(id, pID)
			}
		}
		if pID != 0 {
			pName, _, pExists := reg.Get(pID)
			if !pExists {
				if pMeta, err := proxy.GetMetadata(pID); err == nil {
					pName = pMeta.Name
					reg.Discover(pID, pName, pMeta.AppType)
				}
			}
			buf.WriteString(",[")
			buf.WriteString(strconv.FormatUint(uint64(pID), 10))
			buf.WriteByte(',')
			writeJSONString(buf, pName)
			buf.WriteByte(']')
		}
	}
	buf.WriteByte(']')
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if len(query) < 2 {
		w.WriteHeader(400)
		return
	}
	reg := cache.Current.Load()
	f := func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsNumber(c) }
	tokens := strings.FieldsFunc(query, f)
	if len(tokens) == 0 {
		w.WriteHeader(400)
		return
	}
	candidates, ok := reg.Search[tokens[0]]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteByte('[')
	count := 0
	for _, id := range candidates {
		name := reg.Names[id]
		lowerName := strings.ToLower(name)
		match := true
		for i := 1; i < len(tokens); i++ {
			if !strings.Contains(lowerName, tokens[i]) {
				match = false
				break
			}
		}
		if match {
			if count > 0 {
				buf.WriteByte(',')
			}
			buf.WriteByte('[')
			buf.WriteString(strconv.FormatUint(uint64(id), 10))
			buf.WriteByte(',')
			writeJSONString(buf, name)
			buf.WriteByte(']')
			count++
			if count >= 10 {
				break
			}
		}
	}
	buf.WriteByte(']')
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}
