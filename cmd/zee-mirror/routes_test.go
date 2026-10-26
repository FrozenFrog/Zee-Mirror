package main

import (
	"testing"

	"zee-mirror/internal/router"
	"zee-mirror/pkg/i18n"
)

// Every registered command has a translated description in the catalog, and
// CommandInfo.Description keeps the DefaultLang text.
func TestRegisteredCommandsAreTranslated(t *testing.T) {
	r := router.NewRouter(nil)
	setupRoutes(r)
	cmds := r.GetAllCommandsFlat()
	if len(cmds) == 0 {
		t.Fatal("no commands registered")
	}
	for _, c := range cmds {
		want, ok := i18n.Lookup(i18n.DefaultLang, i18n.CommandKey(c.Name))
		if !ok {
			t.Errorf("command %q has no %s key", c.Name, i18n.CommandKey(c.Name))
			continue
		}
		if c.Description != want {
			t.Errorf("command %q Description = %q, want %q", c.Name, c.Description, want)
		}
	}

	for _, l := range i18n.Languages {
		menu := botMenuCommands(r, l.Code)
		if len(menu) == 0 || len(menu) > 100 {
			t.Fatalf("%s: menu has %d commands", l.Code, len(menu))
		}
		for _, bc := range menu {
			if bc.Description == "" || len(bc.Description) > 256 {
				t.Errorf("%s: /%s has invalid description %q", l.Code, bc.Command, bc.Description)
			}
		}
	}
	en := map[string]string{}
	for _, bc := range botMenuCommands(r, "en") {
		en[bc.Command] = bc.Description
	}
	if en["mirror"] != "Upload a file from a URL to Google Drive" {
		t.Errorf("en mirror description = %q", en["mirror"])
	}
	if _, ok := en["authorize"]; ok {
		t.Error("admin commands must not be in the menu")
	}
}
