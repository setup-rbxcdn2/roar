package docs

import (
	"encoding/json"
	"net/http"
	"sync"
)

const robloxDatatypesURL = "https://api.github.com/repos/Roblox/creator-docs/contents/content/en-us/reference/engine/datatypes"

type ghContentEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var (
	datatypeCacheOnce sync.Once
	datatypeCacheList []string
	datatypeCacheSet  map[string]struct{}
	resolveCache      sync.Map
)

func fetchRobloxDatatypes() []string {
	resp, err := http.Get(robloxDatatypesURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var entries []ghContentEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.Type == "file" && len(e.Name) > 5 && e.Name[len(e.Name)-5:] == ".yaml" {
			names = append(names, e.Name[:len(e.Name)-5])
		}
	}
	return names
}

func loadDatatypeCache() {
	datatypeCacheOnce.Do(func() {
		datatypeCacheList = fetchRobloxDatatypes()
		datatypeCacheSet = make(map[string]struct{}, len(datatypeCacheList))
		for _, n := range datatypeCacheList {
			datatypeCacheSet[n] = struct{}{}
		}
	})
}

func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return i
}

func resolveRobloxDatatype(name string) string {
	if cached, ok := resolveCache.Load(name); ok {
		return cached.(string)
	}

	loadDatatypeCache()

	if _, ok := datatypeCacheSet[name]; ok {
		resolveCache.Store(name, name)
		return name
	}

	best := name
	bestLen := -1
	for _, candidate := range datatypeCacheList {
		l := commonPrefixLen(name, candidate)
		if l > bestLen || (l == bestLen && len(candidate) < len(best)) {
			best = candidate
			bestLen = l
		}
	}

	resolveCache.Store(name, best)
	return best
}