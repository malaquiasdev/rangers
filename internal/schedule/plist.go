package schedule

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const plistSource = `<?xml version="1.0" encoding="UTF-8"?>
` + `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" ` +
	`"http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{xml .Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{xml .Binary}}</string>
    <string>learn</string>
  </array>
  <key>StartInterval</key>
  <integer>{{.Interval}}</integer>
  <key>RunAtLoad</key>
  <true/>
  <key>EnvironmentVariables</key>
  <dict>
    <key>HOME</key>
    <string>{{xml .Home}}</string>
    <key>PATH</key>
    <string>{{xml .Path}}</string>
  </dict>
  <key>StandardOutPath</key>
  <string>{{xml .Log}}</string>
  <key>StandardErrorPath</key>
  <string>{{xml .Log}}</string>
</dict>
</plist>
`

var plistTemplate = template.Must(
	template.New("plist").Funcs(template.FuncMap{"xml": escapeXML}).Parse(plistSource))

type plistData struct {
	Label    string
	Binary   string
	Home     string
	Path     string
	Log      string
	Interval int
}

func (j *job) data() plistData {
	folders := []string{
		filepath.Join(j.home, ".local", "bin"),
		"/opt/homebrew/bin", "/usr/local/go/bin", "/usr/local/bin", "/usr/bin", "/bin",
	}
	return plistData{
		Label:    label,
		Binary:   j.binary,
		Home:     j.home,
		Path:     strings.Join(folders, ":"),
		Log:      j.log(),
		Interval: intervalSeconds,
	}
}

func (j *job) writePlist() error {
	for _, dir := range []string{j.state, filepath.Dir(j.plist)} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	var rendered bytes.Buffer
	if err := plistTemplate.Execute(&rendered, j.data()); err != nil {
		return err
	}
	return os.WriteFile(j.plist, rendered.Bytes(), 0o600)
}

func escapeXML(text string) (string, error) {
	var escaped bytes.Buffer
	err := xml.EscapeText(&escaped, []byte(text))
	return escaped.String(), err
}
