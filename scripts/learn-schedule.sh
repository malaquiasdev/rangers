#!/usr/bin/env bash
set -eu

label=com.rangers.learn
plist="$HOME/Library/LaunchAgents/$label.plist"
state="$HOME/.local/state/rangers/learn"
dominio="gui/$(id -u)"

case "${1:-status}" in
  install)
    mkdir -p "$state" "$HOME/Library/LaunchAgents"
    cat > "$plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>$label</string>
  <key>ProgramArguments</key>
  <array>
    <string>$HOME/.local/bin/rangers</string>
    <string>learn</string>
  </array>
  <key>StartInterval</key>
  <integer>7200</integer>
  <key>RunAtLoad</key>
  <true/>
  <key>EnvironmentVariables</key>
  <dict>
    <key>HOME</key>
    <string>$HOME</string>
    <key>PATH</key>
    <string>$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin</string>
  </dict>
  <key>StandardOutPath</key>
  <string>$state/launchd.log</string>
  <key>StandardErrorPath</key>
  <string>$state/launchd.log</string>
</dict>
</plist>
PLIST
    launchctl bootout "$dominio/$label" 2>/dev/null || true
    launchctl bootstrap "$dominio" "$plist"
    echo "agendado: a cada 2 horas, log em $state/launchd.log"
    ;;
  uninstall)
    launchctl bootout "$dominio/$label" 2>/dev/null || true
    rm -f "$plist"
    echo "agendamento removido"
    ;;
  status)
    launchctl print "$dominio/$label" 2>&1 | grep -E "state =|runs =|last exit code|run interval" || echo "não agendado"
    ;;
  *)
    echo "uso: learn-schedule.sh install|uninstall|status"
    exit 2
    ;;
esac
