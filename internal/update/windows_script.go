package update

import (
	"fmt"
	"strings"
)

// windowsUpdaterScript is the batch file that waits for this process to exit,
// moves newPath over targetPath and starts it again from workDir with args.
func windowsUpdaterScript(targetPath, newPath, workDir string, args []string) string {
	argString := formatArgs(args)
	return fmt.Sprintf(`@echo off
setlocal
set "TARGET=%s"
set "NEWFILE=%s"
set "WORKDIR=%s"
cd /D "%%WORKDIR%%"
:wait
ping 127.0.0.1 -n 2 >nul 2>nul
:loop
move /Y "%%NEWFILE%%" "%%TARGET%%" >nul 2>nul
if errorlevel 1 (
  ping 127.0.0.1 -n 3 >nul 2>nul
  goto loop
)
start "" /b "%%TARGET%%"%s
start "" /b cmd /c "del /q ""%%~f0"""
exit /b
`, escapeForBatch(targetPath), escapeForBatch(newPath), escapeForBatch(workDir), argString)
}

func formatArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = ` "` + quotedForBatch.Replace(arg) + `"`
	}
	return strings.Join(parts, "")
}

// Inside double quotes a batch file reads everything literally except a
// percent sign and the quote itself.
var quotedForBatch = strings.NewReplacer("%", "%%", `"`, `""`)

func escapeForBatch(path string) string {
	replacer := strings.NewReplacer(
		"^", "^^",
		"&", "^&",
		"|", "^|",
		"<", "^<",
		">", "^>",
		"%", "%%",
		"\"", "\"\"",
	)
	return replacer.Replace(path)
}
