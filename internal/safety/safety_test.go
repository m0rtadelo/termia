package safety

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		command string
		want    Level
	}{
		{"ls -la", Safe},
		{"cat file.txt", Safe},
		{"du -sh *", Safe},
		{"grep -rn foo .", Safe},
		{"cp a b", Caution},
		{"mv a b", Caution},
		{"chmod 644 file", Caution},
		{"sudo apt install curl", Caution},
		{"echo hi > out.txt", Caution},
		{"rm file.txt", Caution},
		{"rm -rf /tmp/build", Danger},
		{"rm -rf /", Danger},
		{"dd if=/dev/zero of=/dev/sda", Danger},
		{"mkfs.ext4 /dev/sdb1", Danger},
		{"sudo shutdown -h now", Danger},
		{"chmod -R 777 /", Danger},
		{"git reset --hard HEAD~1", Danger},
		{":(){ :|:& };:", Danger},
	}
	for _, c := range cases {
		if got := Classify(c.command); got != c.want {
			t.Errorf("Classify(%q) = %v, want %v", c.command, got, c.want)
		}
	}
}

func TestLevelString(t *testing.T) {
	if Safe.String() != "SAFE" || Caution.String() != "CAUTION" || Danger.String() != "DANGER" {
		t.Error("unexpected Level.String output")
	}
}
