package dizzclient

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func execDizz(args ...string) (string, error) {
	cmd := exec.Command(FindDizzBinary(), args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("dizz %s: %w\n%s", strings.Join(args, " "), err, ansiRe.ReplaceAllString(stderr.String(), ""))
	}
	return strings.TrimSpace(ansiRe.ReplaceAllString(stdout.String(), "")), nil
}

func Initialize() error {
	_, err := execDizz("init")
	return err
}

func Status() (*Summary, error) {
	raw, err := execDizz("status")
	if err != nil {
		return nil, err
	}
	return ParseStatusOutput(raw), nil
}

func LogDump() ([]Symbol, error) {
	raw, err := execDizz("log", "--dump")
	if err != nil {
		return nil, err
	}
	return ParseLogOutput(raw), nil
}

func ListIntents() ([]Intent, error) {
	raw, err := execDizz("context", "--intents")
	if err != nil {
		return nil, err
	}
	return ParseIntentsTON([]byte(raw))
}

func ListTodos() ([]Todo, error) {
	raw, err := execDizz("context", "--todos")
	if err != nil {
		return nil, err
	}
	return ParseTodosOutput([]byte(raw))
}

func IntentAdd(msg, typ string, severity int, tags []string) error {
	args := []string{"intent", "add", msg, "--type", typ, "--severity", fmt.Sprintf("%d", severity)}
	for _, tag := range tags {
		args = append(args, "--tags", tag)
	}
	_, err := execDizz(args...)
	return err
}

func IntentResolve(id, note string) error {
	args := []string{"intent", "resolve", id}
	if note != "" {
		args = append(args, "--note", note)
	}
	_, err := execDizz(args...)
	return err
}

func IntentClose(id, note string) error {
	args := []string{"intent", "close", id}
	if note != "" {
		args = append(args, "--note", note)
	}
	_, err := execDizz(args...)
	return err
}

func SnapshotCreate() (string, error) {
	raw, err := execDizz("snapshot")
	if err != nil {
		return "", err
	}
	if hash := ParseSnapshotOutput(raw); hash != "" {
		return "Snapshot created: " + hash, nil
	}
	return raw, nil
}

func SnapshotDiff() (string, error) {
	return execDizz("snapshot", "--diff")
}

func SnapshotList() ([]SnapshotInfo, error) {
	raw, err := execDizz("snapshot", "list")
	if err != nil {
		return nil, err
	}
	return ParseSnapshotListOutput(raw), nil
}

func SnapshotCheckout(hash string) (string, error) {
	return execDizz("snapshot", "checkout", hash)
}

func SnapshotPrune(keep int) (string, error) {
	return execDizz("snapshot", "prune", "--keep", fmt.Sprintf("%d", keep))
}
