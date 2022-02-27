package sseclt

import "strings"

type Event map[string]string

func (e Event) AddLine(line string) error {
	// https://html.spec.whatwg.org/multipage/server-sent-events.html
	if line[0] == ':' { // Comment
		return nil
	}
	split := strings.SplitN(line, ":", 2)
	var field string
	var value string
	if len(split) < 2 {
		field = line
		value = ""
	} else {
		field = split[0]
		value = strings.TrimPrefix(split[1], " ")
	}
	e[field] = value
	return nil
}
