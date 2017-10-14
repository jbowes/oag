package translator

import "testing"

func TestFormatVar(t *testing.T) {
	tcs := [][]string{
		{"id", "id"},
		{"thingId", "thingID"},
		{"ThingId", "thingID"},
		{"ThingThinger", "thingThinger"},
		{"thing_thinger", "thingThinger"},
		{"thing-thinger", "thingThinger"},
		{"thing thinger", "thingThinger"},
		{"thing", "thinger", "thingThinger"},
	}

	for _, tc := range tcs {
		t.Run(tc[len(tc)-1], func(t *testing.T) {
			if out := formatVar(tc[:len(tc)-1]...); out != tc[len(tc)-1] {
				t.Error("wanted:", tc[len(tc)-1], "got:", out)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tcs := [][]string{
		{"id", "ID"},
		{"thingId", "ThingID"},
		{"ThingId", "ThingID"},
		{"ThingThinger", "ThingThinger"},
		{"thing_thinger", "ThingThinger"},
		{"thing-thinger", "ThingThinger"},
		{"thing thinger", "ThingThinger"},
		{"thing", "thinger", "ThingThinger"},
	}

	for _, tc := range tcs {
		t.Run(tc[len(tc)-1], func(t *testing.T) {
			if out := formatID(tc[:len(tc)-1]...); out != tc[len(tc)-1] {
				t.Error("wanted:", tc[len(tc)-1], "got:", out)
			}
		})
	}
}
