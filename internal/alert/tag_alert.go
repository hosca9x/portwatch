package alert

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// TagAlerter emits alerts for port entries that carry specific tags.
type TagAlerter struct {
	w       io.Writer
	tags    map[ports.Tag]struct{}
	tagMap  *ports.TagMap
}

// NewTagAlerter creates a TagAlerter that fires when an entry's key has any of
// the watched tags. If w is nil it defaults to os.Stdout.
func NewTagAlerter(w io.Writer, tagMap *ports.TagMap, watched ...ports.Tag) *TagAlerter {
	if w == nil {
		w = os.Stdout
	}
	set := make(map[ports.Tag]struct{}, len(watched))
	for _, tag := range watched {
		set[tag] = struct{}{}
	}
	return &TagAlerter{w: w, tags: set, tagMap: tagMap}
}

// Notify checks each entry; if its key carries a watched tag an alert line is
// written to the writer.
func (a *TagAlerter) Notify(entries []ports.PortEntry) {
	for _, e := range entries {
		matched := a.matchedTags(e.Key())
		if len(matched) == 0 {
			continue
		}
		sort.Strings(matched)
		fmt.Fprintf(a.w, "[tag-alert] port %d/%s matched tags: %s\n",
			e.Port, e.Proto, strings.Join(matched, ","))
	}
}

func (a *TagAlerter) matchedTags(key string) []string {
	var out []string
	for _, tag := range a.tagMap.Tags(key) {
		if _, ok := a.tags[tag]; ok {
			out = append(out, string(tag))
		}
	}
	return out
}
