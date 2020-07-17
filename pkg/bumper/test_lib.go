package bumper

import "testing"

func TestDetectPid(t *testing.T) {
	//d := DetectPid(15)
}

func TestDetectCommand(t *testing.T) {
	_, err := DetectCommand("command")
	if err != nil {
		t.Log("Command detection failed {}", err)
		t.Fail()
	}
}
