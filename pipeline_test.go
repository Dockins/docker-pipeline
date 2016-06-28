package main

import "testing"

func TestParseCommand(t *testing.T) {
	p, err := Parse([]byte(`
build:
    image: maven:3.3.3-jdk-8
    commands:
    -   echo hello    
    -   mvn -version    
    `))

	// TODO compare actual with expected Pipeline{"build":Stage{...}}

	if err != nil {
		t.Errorf("Failed to parse yml content %s", err)
		return
	}
	if len(p) != 1 {
		t.Errorf("Failed to parse yml content, len= %s", len(p))
		return
	}
	s, ok := p["build"]
	if !ok {
		t.Errorf("Failed to parse yml content - no 'build' stage")
		return
	}

	if s.Order != 0 {
		t.Errorf("Failed to parse yml content - 'build' stage should get Order=0")
		return
	}

	switch c := s.Exec.(type) {
	case Command:
		if c.Image != "maven:3.3.3-jdk-8" {
			t.Errorf("unexpected Image %s", c.Image)
		}
		if len(c.Commands) != 2 {
			t.Errorf("unexpected Commands %s", c.Commands)
		}
		if c.Commands[1] != "mvn -version" {
			t.Errorf("unexpected Command %s", c.Commands[1])
		}
	default:
		t.Errorf("Failed to parse yml content - 'build' stage isn't a Command")
	}

}
